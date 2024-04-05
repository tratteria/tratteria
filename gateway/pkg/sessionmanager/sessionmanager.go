package sessionmanager

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"sync"
	"time"
)

type UserSession struct {
	Email   string
	Expires time.Time
}

var (
	sessionStore = make(map[string]UserSession)
	storeLock    sync.RWMutex
)

func generateSessionID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		log.Fatal(err)
	}

	return hex.EncodeToString(b)
}

func SaveSession(sessionData UserSession) string {
	sessionID := generateSessionID()
	storeLock.Lock()
	defer storeLock.Unlock()
	sessionStore[sessionID] = sessionData

	return sessionID
}

func GetSession(sessionID string) (UserSession, bool) {
	storeLock.RLock()
	defer storeLock.RUnlock()
	sessionData, exists := sessionStore[sessionID]

	return sessionData, exists
}

func DeleteSession(sessionID string) {
	storeLock.Lock()
	defer storeLock.Unlock()
	delete(sessionStore, sessionID)
}
