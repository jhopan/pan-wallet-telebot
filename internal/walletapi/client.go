package walletapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Client memanggil endpoint publik Pan-Wallet.
// Bot tidak memegang API key: semua endpoint yang dipakai bersifat publik.
type Client struct {
	baseURL string
	http    *http.Client
}

// Wallet adalah jawaban cek saldo.
type Wallet struct {
	PublicID  string `json:"public_id"`
	Balance   int64  `json:"balance"`
	Held      int64  `json:"held"`
	Available int64  `json:"available"`
	Currency  string `json:"currency"`
	Status    string `json:"status"`
	IsNew     bool   `json:"is_new"`
}

// Topup adalah jawaban pembuatan invoice.
type Topup struct {
	InvoiceID string `json:"invoice_id"`
	Total     int64  `json:"total"`
	Status    string `json:"status"`
	PayURL    string `json:"pay_url"`
	ExpiresAt int64  `json:"expires_at"`
}

// New membuat client dengan base URL Pan-Wallet.
func New(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

// Resolve mencari wallet dari identitas Telegram. Membuat baru bila belum ada.
func (client *Client) Resolve(ctx context.Context, telegramID string) (Wallet, error) {
	var wallet Wallet
	err := client.post(ctx, "/v1/identity/resolve", map[string]any{
		"provider":    "telegram",
		"external_id": telegramID,
	}, &wallet)
	return wallet, err
}

// Balance membaca saldo berdasarkan public_id.
func (client *Client) Balance(ctx context.Context, publicID string) (Wallet, error) {
	var wallet Wallet
	request, err := http.NewRequestWithContext(ctx, http.MethodGet,
		client.baseURL+"/v1/public/"+publicID, nil)
	if err != nil {
		return Wallet{}, err
	}
	response, err := client.http.Do(request)
	if err != nil {
		return Wallet{}, fmt.Errorf("hubungi Pan-Wallet: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return Wallet{}, fmt.Errorf("saldo gagal dibaca (HTTP %d)", response.StatusCode)
	}
	if err := json.NewDecoder(response.Body).Decode(&wallet); err != nil {
		return Wallet{}, fmt.Errorf("baca jawaban: %w", err)
	}
	return wallet, nil
}

// CreateTopup membuat invoice top-up untuk pelanggan.
func (client *Client) CreateTopup(ctx context.Context, publicID string, amount int64) (Topup, error) {
	var topup Topup
	err := client.post(ctx, "/v1/public/"+publicID+"/topup", map[string]any{"amount": amount}, &topup)
	return topup, err
}

// LinkConfirm mengonfirmasi kode ikat yang dikirim pelanggan.
func (client *Client) LinkConfirm(ctx context.Context, code, telegramID string) (Wallet, error) {
	var wallet Wallet
	err := client.post(ctx, "/v1/identity/link/confirm", map[string]any{
		"code":        code,
		"telegram_id": telegramID,
	}, &wallet)
	return wallet, err
}

func (client *Client) post(ctx context.Context, path string, payload any, target any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode body: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.baseURL+path, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := client.http.Do(request)
	if err != nil {
		return fmt.Errorf("hubungi Pan-Wallet: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 {
		var failure struct {
			Error string `json:"error"`
		}
		json.NewDecoder(response.Body).Decode(&failure)
		if failure.Error != "" {
			return fmt.Errorf("%s", failure.Error)
		}
		return fmt.Errorf("Pan-Wallet menolak (HTTP %d)", response.StatusCode)
	}
	if target != nil {
		if err := json.NewDecoder(response.Body).Decode(target); err != nil {
			return fmt.Errorf("baca jawaban: %w", err)
		}
	}
	return nil
}

// PayURL mengubah path relatif dari Pan-Wallet menjadi URL absolut.
// Pan-Wallet mengembalikan pay_url relatif (misal /pay/abc), sedangkan
// pelanggan perlu URL lengkap agar bisa diklik dari Telegram.
func (client *Client) PayURL(path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if path == "" {
		return ""
	}
	return client.baseURL + path
}
