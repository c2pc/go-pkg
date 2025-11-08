package logger

import (
	"sync"

	"github.com/op/go-logging"
)

type logCache struct {
	mutex   sync.RWMutex
	loggers map[string]*logging.Logger
}

// getLogger gets logger for given modules. It creates a new logger for the module if not exists
func (l *logCache) getLogger(module string) *logging.Logger {
	if !initialized.Load() {
		return nil
	}
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	if module == "" {
		return l.loggers["LOG"]
	}
	if _, ok := l.loggers[module]; !ok {
		l.mutex.RUnlock()
		l.addLogger(module, backendFileLeveled)
		l.mutex.RLock()
	}

	return l.loggers[module]
}

func (l *logCache) addLogger(module string, log logging.LeveledBackend) {
	logger := logging.MustGetLogger(module)
	logger.SetBackend(log)
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.loggers[module] = logger
}
