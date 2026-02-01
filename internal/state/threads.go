package state

import "sync"

// ActiveThreads tracks threads where the bot should respond to all messages
var ActiveThreads sync.Map

func AddThread(threadID string) {
	ActiveThreads.Store(threadID, struct{}{})
}

func IsActiveThread(threadID string) bool {
	_, ok := ActiveThreads.Load(threadID)
	return ok
}

func RemoveThread(threadID string) {
	ActiveThreads.Delete(threadID)
}
