package state

import "testing"

func TestThreadIDToSessionID_Deterministic(t *testing.T) {
	threadID := "1234567890123456789"

	id1 := ThreadIDToSessionID(threadID)
	id2 := ThreadIDToSessionID(threadID)

	if id1 != id2 {
		t.Errorf("Expected same session ID, got %s and %s", id1, id2)
	}

	t.Logf("Thread ID: %s -> Session ID: %s", threadID, id1)
}

func TestThreadIDToSessionID_DifferentThreads(t *testing.T) {
	thread1 := "1234567890123456789"
	thread2 := "9876543210987654321"

	id1 := ThreadIDToSessionID(thread1)
	id2 := ThreadIDToSessionID(thread2)

	if id1 == id2 {
		t.Errorf("Expected different session IDs for different threads, got same: %s", id1)
	}

	t.Logf("Thread %s -> %s", thread1, id1)
	t.Logf("Thread %s -> %s", thread2, id2)
}
