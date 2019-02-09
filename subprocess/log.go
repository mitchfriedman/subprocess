package subprocess

import (
	"bytes"
	"fmt"
	"sync"
)

type logger struct {
	logLock sync.RWMutex
	log     bytes.Buffer
}

func (l *logger) Printf(line string, format ...interface{}) {
	l.logLock.Lock()
	defer l.logLock.Unlock()
	s := fmt.Sprintf(line, format)
	l.log.Write([]byte(s + "\n"))
}
