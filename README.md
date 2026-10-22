# Pan-Wallet Telegram Bot

Bot Telegram untuk saldo Pan-Wallet. Bot ini **klien**: semua saldo disimpan di
service Pan-Wallet, bot hanya menampilkan dan meneruskan perintah.

## Perintah

| Perintah | Fungsi |
|---|---|
| `/start` | daftar otomatis, dapat ID pelanggan |
| `/saldo` | lihat saldo |
| `/topup 50000` | buat invoice top-up QRIS |
| `/ikat 482913` | hubungkan email/WhatsApp ke akun ini |
| `/bantuan` | daftar perintah |

## Jalan lokal

```bash
cp .env.example .env      # isi BOT_TOKEN dan PANWALLET_URL
go build -o bot.exe ./cmd/bot
./bot.exe
```

Prasyarat: Pan-Wallet sudah jalan di `PANWALLET_URL`.

## Struktur

```text
cmd/bot/            entrypoint, long polling Telegram
internal/bot/       handler perintah + format pesan
internal/walletapi/ klien HTTP ke Pan-Wallet
internal/store/     sesi public_id per chat (memori)
internal/config/    baca environment
```

## Test

```bash
go test ./...
```
