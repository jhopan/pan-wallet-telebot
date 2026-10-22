package store

import (
	"sync"
)

// Session menyimpan public_id pelanggan selama bot berjalan.
// Sengaja hanya di memori: saldo tetap di Pan-Wallet, bukan di bot.
type Session struct {
	mu       sync.RWMutex
	byChatID map[int64]string
}

// New membuat penyimpanan sesi.
func New() *Session {
	return &Session{byChatID: map[int64]string{}}
}

// SetPublicID menyimpan public_id untuk satu chat Telegram.
func (session *Session) SetPublicID(chatID int64, publicID string) {
	session.mu.Lock()
	session.byChatID[chatID] = publicID
	session.mu.Unlock()
}

// PublicID mengambil public_id untuk satu chat.
func (session *Session) PublicID(chatID int64) (string, bool) {
	session.mu.RLock()
	defer session.mu.RUnlock()
	value, ok := session.byChatID[chatID]
	return value, ok
}
