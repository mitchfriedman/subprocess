package subprocess

import (
	"bytes"
	"fmt"
	"sync"
)

type logger struct {
	logLock sync.RWMutex
	bytes.Buffer
}

func (l *logger) Printf(line string, format ...interface{}) {
	l.logLock.Lock()
	defer l.logLock.Unlock()
	s := fmt.Sprintf(line, format)
	_, _ = l.Write([]byte(s + "\n"))
}
