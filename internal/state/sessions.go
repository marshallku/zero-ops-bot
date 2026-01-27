package state

import (
    "github.com/google/uuid"
)

// Namespace UUID for generating deterministic session IDs from thread IDs
var sessionNamespace = uuid.MustParse("a1b2c3d4-e5f6-7890-abcd-ef1234567890")

// ThreadIDToSessionID generates a deterministic UUID from a Discord thread ID.
// Same thread ID always produces the same session UUID.
func ThreadIDToSessionID(threadID string) string {
    return uuid.NewSHA1(sessionNamespace, []byte(threadID)).String()
}
