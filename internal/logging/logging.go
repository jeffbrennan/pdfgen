package logging

import "sync"

var (
	LogChannel = make(chan string)
	LogMu      sync.Mutex
)

func PublishLog(message string) {
	LogMu.Lock()
	defer LogMu.Unlock()
	select {
	case LogChannel <- message:
	default:
	}
}
