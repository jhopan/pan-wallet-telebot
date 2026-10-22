package bot

import (
	"fmt"
	"strconv"
	"strings"
)

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
			"Tersedia: Rp %s\nTertahan: Rp %s\nTotal: Rp %s",
		publicID, FormatRupiah(available), FormatRupiah(held), FormatRupiah(balance),
	)
}

// TopupMessage menyusun pesan invoice top-up.
func TopupMessage(total int64, payURL string, minutes int) string {
	return fmt.Sprintf(
		"Top-up dibuat\n\nNominal: Rp %s\nBayar di sini:\n%s\n\nBerlaku %d menit.",
		FormatRupiah(total), payURL, minutes,
	)
}

// HelpMessage menjelaskan perintah yang tersedia.
func HelpMessage() string {
	return "Perintah yang tersedia:\n\n" +
		"/saldo - lihat saldo kamu\n" +
		"/topup <nominal> - isi saldo, contoh /topup 50000\n" +
		"/ikat <kode> - hubungkan email atau WhatsApp ke akun ini\n" +
		"/bantuan - tampilkan pesan ini"
}

// WelcomeMessage menyapa pelanggan baru.
func WelcomeMessage(publicID string, isNew bool) string {
	if isNew {
		return fmt.Sprintf(
			"Akun kamu dibuat.\n\nID Pelanggan: %s\n\nSimpan ID ini. Saldo masih Rp 0.\n"+
				"Ketik /topup 50000 untuk isi saldo.",
			publicID,
		)
	}
	return fmt.Sprintf("Halo lagi. ID Pelanggan kamu: %s", publicID)
}
