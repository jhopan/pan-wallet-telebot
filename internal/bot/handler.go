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
	// StoreBot adalah username bot toko untuk diarahkan dari /bantuan.
	StoreBot string
}

// Reply adalah jawaban untuk satu pesan.
type Reply struct {
	Text     string
	Keyboard *Keyboard
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
		menu := MainKeyboard()
		return Reply{Text: handler.helpText(), Keyboard: &menu}
	}

	switch command {
	case "start":
		wallet, err := handler.ensureRegistered(ctx, chatID, telegramID)
		if err != nil {
			return Reply{Text: "Maaf, akun kamu belum bisa dibuat. Coba lagi sebentar."}
		}
		menu := MainKeyboard()
		return Reply{Text: WelcomeMessage(wallet.PublicID, wallet.IsNew), Keyboard: &menu}

	case "saldo":
		wallet, err := handler.ensureRegistered(ctx, chatID, telegramID)
		if err != nil {
			return Reply{Text: "Saldo belum bisa dibaca. Coba lagi sebentar."}
		}
		menu := MainKeyboard()
		return Reply{
			Text:     BalanceMessage(wallet.PublicID, wallet.Balance, wallet.Held, wallet.Available),
			Keyboard: &menu,
		}

	case "topup":
		// Dengan argumen: /topup 50000. Tanpa argumen: tampilkan tombol nominal.
		trimmed := strings.TrimSpace(args)
		if trimmed == "" {
			keyboard := TopupKeyboard()
			return Reply{Text: "Pilih nominal top-up:", Keyboard: &keyboard}
		}
		amount, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil || amount < 1000 {
			keyboard := TopupKeyboard()
			return Reply{Text: "Nominal minimal Rp 1.000.\n\nPilih nominal top-up:", Keyboard: &keyboard}
		}
		return handler.createTopup(ctx, chatID, telegramID, amount)

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
		menu := MainKeyboard()
		return Reply{Text: handler.helpText(), Keyboard: &menu}
	}

	menu := MainKeyboard()
	return Reply{Text: "Perintah tidak dikenal.\n\n" + handler.helpText(), Keyboard: &menu}
}

// HandleCallback memproses tekan tombol inline.
func (handler *Handler) HandleCallback(ctx context.Context, chatID int64, telegramID, data string) Reply {
	switch data {
	case "saldo":
		return handler.Handle(ctx, chatID, telegramID, "/saldo")

	case "topupmenu":
		keyboard := TopupKeyboard()
		return Reply{Text: "Pilih nominal top-up:", Keyboard: &keyboard}
	}

	if amount, ok := ParseCallbackAmount(data); ok {
		return handler.createTopup(ctx, chatID, telegramID, amount)
	}
	return Reply{Text: "Tombol tidak dikenali."}
}

// createTopup membuat invoice dan mengembalikan link bayar.
func (handler *Handler) createTopup(ctx context.Context, chatID int64, telegramID string, amount int64) Reply {
	wallet, err := handler.ensureRegistered(ctx, chatID, telegramID)
	if err != nil {
		return Reply{Text: "Akun belum siap. Coba lagi sebentar."}
	}
	topup, err := handler.API.CreateTopup(ctx, wallet.PublicID, amount)
	if err != nil {
		return Reply{Text: "Top-up gagal: " + err.Error()}
	}
	return Reply{Text: TopupMessage(topup.Total, topup.PayURL, 15)}
}

// helpText memakai StoreBot bila dikonfigurasi.
func (handler *Handler) helpText() string {
	return HelpMessageWithStore(handler.StoreBot)
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
