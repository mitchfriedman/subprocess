package subprocess

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"sync"
	"syscall"
	"time"

	"github.com/kr/pty"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh/terminal"
)

var ErrTimeout = errors.New("timeout expecting results")

const DefaultTimeout = 30 * time.Second

type SubProcess struct {
	command  *exec.Cmd
	ctx      context.Context
	pty      *os.File
	log      *logger
	oldState *terminal.State
}

func NewSubProcess(command string, args ...string) (*SubProcess, error) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, command, args...)

	return &SubProcess{
		command: cmd,
		log:     &logger{},
		ctx:     ctx,
	}, nil
}

func (s *SubProcess) listenForShutdown(signals chan os.Signal, errs chan error, stop chan struct{}) {
	for {
		select {
		case e := <-errs:
			log.Printf("failed with error: %v", e)
			stop <- struct{}{}
			return

		case sig := <-signals:
			switch sig {
			case syscall.SIGWINCH:
				if err := pty.InheritSize(os.Stdin, s.pty); err != nil {
					// probably not worth shutting down the process over this error, so let's log and move on
					log.Printf("error resizing pty: %s", err)
				}

			default:
				stop <- struct{}{}
				return
			}
		}
	}
}

func waitForCommandCompletion(cmd *exec.Cmd, errs chan error, stop chan struct{}) {
	err := cmd.Wait()
	if err != nil {
		errs <- err
	}
	stop <- struct{}{}
}

func (s *SubProcess) Interact() {
	errs := make(chan error)
	stop := make(chan struct{}, 1)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGINT, syscall.SIGWINCH, syscall.SIGTSTP)

	_, cancel := context.WithCancel(s.ctx)

	go s.listenForShutdown(signals, errs, stop)
	go waitForCommandCompletion(s.command, errs, stop)
	go io.Copy(os.Stdout, s.pty)
	go io.Copy(s.pty, os.Stdin)

	<-stop
	cancel()
	close(stop)
	_ = s.pty.Close()
}

func (s *SubProcess) LogOutput() string {
	return s.log.String()
}

func (s *SubProcess) Start() error {
	p, err := pty.Start(s.command)
	if err != nil {
		return err
	}
	s.pty = p
	s.oldState, err = terminal.MakeRaw(int(os.Stdin.Fd()))
	return err
}

func (s *SubProcess) Close() error {
	defer func() {
		_ = terminal.Restore(int(os.Stdin.Fd()), s.oldState)
	}()
	if s.command != nil && s.command.Process != nil {
		return s.command.Process.Kill()
	}
	return nil
}

func (s *SubProcess) Send(value string) error {
	_, err := s.pty.Write([]byte(value))
	return err
}

func (s *SubProcess) SendLine(value string) error {
	return s.Send(value + "\r\n")
}

func (s *SubProcess) ExpectWithTimeout(expression *regexp.Regexp, duration time.Duration) (bool, error) {
	expressions := []*regexp.Regexp{
		expression,
	}
	index, err := s.ExpectExpressionsWithTimeout(expressions, duration)
	return index == 0, err
}

func (s *SubProcess) Expect(expression *regexp.Regexp) (bool, error) {
	return s.ExpectWithTimeout(expression, DefaultTimeout)
}

func (s *SubProcess) ExpectExpressions(expressions []*regexp.Regexp) (int, error) {
	return s.ExpectExpressionsWithTimeout(expressions, DefaultTimeout)
}

func (s *SubProcess) ExpectExpressionsWithTimeout(expressions []*regexp.Regexp, timeout time.Duration) (int, error) {
	errs := make(chan error, 1)
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(timeout))
	ctx, cancelFunc := context.WithCancel(ctx)

	var output bytes.Buffer
	var rwLock sync.RWMutex

	var wg sync.WaitGroup

	wg.Add(1)
	go s.readOutput(ctx, &wg, &output, &rwLock, errs)

	var index = -1
	var e error

OUTER:
	for {
		select {
		case <-ctx.Done():
			e = ErrTimeout

		case err := <-errs:
			s.log.Printf("error reading from pty: %v", err)
			e = errors.Wrap(err, "error reading from pty")
			break OUTER

		case <-time.After(50 * time.Microsecond): // TODO: adjust this
			rwLock.RLock()
			b := output.Bytes()
			rwLock.RUnlock()

			for i, r := range expressions {
				if r.Find(b) != nil {
					index = i
					break OUTER
				}
			}
		}
	}

	cancelFunc()
	wg.Wait()
	return index, e
}

func (s *SubProcess) readOutput(ctx context.Context, wg *sync.WaitGroup, buf io.Writer, lock *sync.RWMutex, errs chan error) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			var temp bytes.Buffer

			n, err := io.Copy(&temp, s.pty)
			if err != nil && err != io.EOF {
				errs <- err
				close(errs)
				return
			}

			if n > 0 {
				lock.Lock()
				_, _ = buf.Write(temp.Bytes())
				lock.Unlock()
			}
		}
	}
}
