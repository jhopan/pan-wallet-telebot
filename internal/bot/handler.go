package bot

import (
	"context"
	"strconv"
	"strings"
)

// WalletAPI adalah ketergantungan ke Pan-Wallet.
type WalletAPI interface {
	Resolve(ctx context.Context, telegramID string) (Wallet, error)
	Balance(ctx context.Context, publicID string) (Wallet, error)
	CreateTopup(ctx context.Context, publicID string, amount int64) (Topup, error)
	LinkConfirm(ctx context.Context, code, telegramID string) (Wallet, error)
}

// Wallet adalah data saldo yang dibutuhkan bot.
type Wallet struct {
	PublicID  string
	Balance   int64
	Held      int64
	Available int64
	Status    string
	IsNew     bool
}

// Topup adalah data invoice top-up.
type Topup struct {
	Total  int64
	PayURL string
}

// Sessions menyimpan public_id per chat.
type Sessions interface {
	SetPublicID(chatID int64, publicID string)
	PublicID(chatID int64) (string, bool)
}

// Handler memproses satu pesan Telegram.
type Handler struct {
	API      WalletAPI
	Sessions Sessions
}

// Reply adalah jawaban untuk satu pesan.
type Reply struct {
	Text string
}

// Handle memproses pesan masuk dan mengembalikan teks balasan.
func (handler *Handler) Handle(ctx context.Context, chatID int64, telegramID, text string) Reply {
	command, args := ParseCommand(text)

	// Pesan bukan perintah: tetap pastikan pelanggan terdaftar, lalu bantu.
	if command == "" {
		if _, ok := handler.Sessions.PublicID(chatID); !ok {
			if _, err := handler.ensureRegistered(ctx, chatID, telegramID); err != nil {
				return Reply{Text: "Maaf, layanan saldo sedang tidak bisa dihubungi. Coba lagi sebentar."}
			}
		}
		return Reply{Text: HelpMessage()}
	}

	switch command {
	case "start":
		wallet, err := handler.ensureRegistered(ctx, chatID, telegramID)
		if err != nil {
			return Reply{Text: "Maaf, akun kamu belum bisa dibuat. Coba lagi sebentar."}
		}
		return Reply{Text: WelcomeMessage(wallet.PublicID, wallet.IsNew)}

	case "saldo":
		wallet, err := handler.ensureRegistered(ctx, chatID, telegramID)
		if err != nil {
			return Reply{Text: "Saldo belum bisa dibaca. Coba lagi sebentar."}
		}
		return Reply{Text: BalanceMessage(wallet.PublicID, wallet.Balance, wallet.Held, wallet.Available)}

	case "topup":
		amount, err := strconv.ParseInt(strings.TrimSpace(args), 10, 64)
		if err != nil || amount < 1000 {
			return Reply{Text: "Format: /topup 50000\nNominal minimal Rp 1.000."}
		}
		wallet, err := handler.ensureRegistered(ctx, chatID, telegramID)
		if err != nil {
			return Reply{Text: "Akun belum siap. Coba lagi sebentar."}
		}
		topup, err := handler.API.CreateTopup(ctx, wallet.PublicID, amount)
		if err != nil {
			return Reply{Text: "Top-up gagal: " + err.Error()}
		}
		return Reply{Text: TopupMessage(topup.Total, topup.PayURL, 15)}

	case "ikat":
		code := strings.TrimSpace(args)
		if code == "" {
			return Reply{Text: "Format: /ikat 482913\nKode didapat dari halaman web saat memasukkan email."}
		}
		wallet, err := handler.API.LinkConfirm(ctx, code, telegramID)
		if err != nil {
			return Reply{Text: "Gagal mengikat: " + err.Error()}
		}
		handler.Sessions.SetPublicID(chatID, wallet.PublicID)
		return Reply{Text: "Berhasil. Akun terhubung ke ID " + wallet.PublicID + "."}

	case "bantuan", "help":
		return Reply{Text: HelpMessage()}
	}

	return Reply{Text: "Perintah tidak dikenal.\n\n" + HelpMessage()}
}

// ensureRegistered memastikan pelanggan punya wallet, lalu simpan di sesi.
func (handler *Handler) ensureRegistered(ctx context.Context, chatID int64, telegramID string) (Wallet, error) {
	wallet, err := handler.API.Resolve(ctx, telegramID)
	if err != nil {
		return Wallet{}, err
	}
	handler.Sessions.SetPublicID(chatID, wallet.PublicID)
	return wallet, nil
}
