package config

import (
	"fmt"
	"os"
	"strings"
)

// Settings adalah konfigurasi bot dari environment.
type Settings struct {
	BotToken     string
	PanWalletURL string
	// StoreBot adalah username bot toko (tanpa @) untuk diarahkan pelanggan.
	StoreBot string
}

// Load membaca konfigurasi dan memvalidasi.
func Load() (Settings, error) {
	settings := Settings{
		BotToken:     strings.TrimSpace(os.Getenv("BOT_TOKEN")),
		PanWalletURL: strings.TrimSpace(os.Getenv("PANWALLET_URL")),
		StoreBot:     strings.TrimSpace(os.Getenv("STORE_BOT")),
	}
	if settings.BotToken == "" {
		return Settings{}, fmt.Errorf("BOT_TOKEN wajib diisi")
	}
	if settings.PanWalletURL == "" {
		settings.PanWalletURL = "http://127.0.0.1:8080"
	}
	return settings, nil
}
