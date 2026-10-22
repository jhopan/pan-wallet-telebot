package bot

import (
	"fmt"
	"strconv"
	"strings"
)

// MinTopup adalah nominal top-up terendah yang diterima Pan-Wallet.
const MinTopup int64 = 25_000

// DisclaimerSaldo menjelaskan saldo tidak bisa diuangkan.
const DisclaimerSaldo = "Saldo hanya bisa dipakai untuk transaksi di layanan kami. Saldo tidak dapat diuangkan atau ditarik tunai."

// Button adalah satu tombol inline Telegram.
type Button struct {
	Text     string
	Callback string
	URL      string
}

// Keyboard adalah susunan tombol inline.
type Keyboard struct {
	Rows [][]Button
}

// TopupKeyboard menyediakan pilihan nominal siap tekan.
func TopupKeyboard() Keyboard {
	return Keyboard{Rows: [][]Button{
		{
			{Text: "Rp 25.000", Callback: "topup:25000"},
			{Text: "Rp 50.000", Callback: "topup:50000"},
		},
		{
			{Text: "Rp 100.000", Callback: "topup:100000"},
			{Text: "Rp 200.000", Callback: "topup:200000"},
		},
	}}
}

// MainKeyboard adalah menu utama setelah /start.
func MainKeyboard() Keyboard {
	return Keyboard{Rows: [][]Button{
		{
			{Text: "Cek Saldo", Callback: "saldo"},
			{Text: "Top Up", Callback: "topupmenu"},
		},
	}}
}

// ParseCallbackAmount membaca callback "topup:50000" menjadi nominal.
func ParseCallbackAmount(data string) (int64, bool) {
	if !strings.HasPrefix(data, "topup:") {
		return 0, false
	}
	raw := strings.TrimPrefix(data, "topup:")
	amount, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || amount < MinTopup {
		return 0, false
	}
	return amount, true
}

// ParseCommand memisahkan "/perintah argumen" menjadi nama perintah dan argumen.
// Mengembalikan string kosong bila pesan bukan perintah.
func ParseCommand(text string) (string, string) {
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "/") {
		return "", ""
	}
	parts := strings.SplitN(trimmed, " ", 2)
	command := strings.ToLower(strings.TrimPrefix(parts[0], "/"))
	// Buang mention bot, misal /saldo@PanWalletBot
	if at := strings.Index(command, "@"); at > 0 {
		command = command[:at]
	}
	args := ""
	if len(parts) == 2 {
		args = strings.TrimSpace(parts[1])
	}
	return command, args
}

// FormatRupiah memformat nominal menjadi "25.000".
func FormatRupiah(amount int64) string {
	negative := amount < 0
	if negative {
		amount = -amount
	}
	digits := strconv.FormatInt(amount, 10)
	var builder strings.Builder
	length := len(digits)
	for index, digit := range digits {
		if index > 0 && (length-index)%3 == 0 {
			builder.WriteByte('.')
		}
		builder.WriteRune(digit)
	}
	if negative {
		return "-" + builder.String()
	}
	return builder.String()
}

// BalanceMessage menyusun pesan saldo.
func BalanceMessage(publicID string, balance, held, available int64) string {
	return fmt.Sprintf(
		"Saldo kamu\n\nID Pelanggan: %s\n\n"+
			"Tersedia: Rp %s\nTertahan: Rp %s\nTotal: Rp %s\n\n"+
			"%s",
		publicID, FormatRupiah(available), FormatRupiah(held), FormatRupiah(balance),
		DisclaimerSaldo,
	)
}

// TopupMessage menyusun pesan invoice top-up.
func TopupMessage(total int64, payURL string, minutes int) string {
	return fmt.Sprintf(
		"Top-up dibuat\n\nNominal: Rp %s\nBayar di sini:\n%s\n\nBerlaku %d menit.",
		FormatRupiah(total), payURL, minutes,
	)
}

// PaidMessage memberi tahu saldo masuk.
func PaidMessage(amount, balance int64) string {
	return fmt.Sprintf(
		"Saldo masuk\n\nNominal: Rp %s\nSaldo sekarang: Rp %s",
		FormatRupiah(amount), FormatRupiah(balance),
	)
}

// HelpMessage adalah bantuan tanpa bot toko.
func HelpMessage() string { return HelpMessageWithStore("") }

// HelpMessageWithStore menjelaskan peran bot ini dan mengarahkan ke bot toko.
func HelpMessageWithStore(storeBot string) string {
	text := "Bot Saldo Pan-Wallet\n\n" +
		"Bot ini khusus mengurus saldo. Tidak menjual produk.\n\n" +
		"Perintah:\n" +
		"/saldo - lihat saldo kamu\n" +
		"/topup - isi saldo, minimal Rp 25.000\n" +
		"/ikat <kode> - hubungkan email atau WhatsApp ke akun ini\n" +
		"/bantuan - tampilkan pesan ini"

	if strings.TrimSpace(storeBot) != "" {
		text += fmt.Sprintf(
			"\n\nUntuk beli pulsa, paket data, atau lisensi, buka bot toko kami:\n@%s",
			strings.TrimPrefix(storeBot, "@"),
		)
	}
	return text
}

// WelcomeMessage menyapa pelanggan baru dan menjelaskan peran bot.
func WelcomeMessage(publicID string, isNew bool) string {
	if isNew {
		return fmt.Sprintf(
			"Selamat datang di Bot Saldo Pan-Wallet\n\n"+
				"Bot ini khusus untuk:\n"+
				"• daftar akun saldo\n"+
				"• isi saldo (top-up QRIS)\n"+
				"• cek saldo\n\n"+
				"ID Pelanggan kamu: %s\n"+
				"Saldo: Rp 0\n\n"+
				"Minimal top-up Rp 25.000.\n"+
				"%s\n\n"+
				"Tekan Top Up untuk mulai mengisi saldo.",
			publicID, DisclaimerSaldo,
		)
	}
	return fmt.Sprintf(
		"Halo lagi.\n\nID Pelanggan kamu: %s\n\n"+
			"Tekan Cek Saldo atau Top Up di bawah.",
		publicID,
	)
}

// NotTopupHereMessage dipakai bila bot toko dipanggil perintah top-up.
// Bot toko tidak boleh memproses top-up; pelanggan diarahkan ke bot saldo.
func NotTopupHereMessage(balanceBot string) string {
	return fmt.Sprintf(
		"Top-up hanya bisa dilakukan di bot saldo.\n\nBuka: @%s",
		strings.TrimPrefix(balanceBot, "@"),
	)
}
