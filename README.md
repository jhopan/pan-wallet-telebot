# Pan-Wallet Telegram Bot

Bot Telegram **khusus saldo** untuk Pan-Wallet. Bot ini **klien**: semua saldo
disimpan di service Pan-Wallet, bot hanya menampilkan dan meneruskan perintah.

## Prinsip: satu bot, satu tanggung jawab

Bot ini **hanya** mengurus:

- registrasi akun saldo
- top-up QRIS
- cek saldo

Bot ini **tidak** menjual pulsa, paket data, atau lisensi. Itu tugas bot toko
masing-masing. Kalau pelanggan butuh produk, bot ini mengarahkan ke bot toko.

```
Bot AgenPulsa   -> jual pulsa/paket. Cek saldo, tapi TIDAK top-up.
Bot Xray        -> jual lisensi.      Cek saldo, tapi TIDAK top-up.
Bot Pan-Wallet  -> registrasi + top-up. HANYA di sini top-up bisa.
```

## Ketentuan saldo

- **Minimal top-up Rp 25.000.**
- **Saldo tidak dapat diuangkan** atau ditarik tunai. Saldo hanya bisa dipakai
  untuk transaksi di layanan Jhopan (pulsa, paket data, lisensi, dan produk lain).

## Perintah

| Perintah | Fungsi |
|---|---|
| `/start` | daftar otomatis, dapat ID pelanggan |
| `/saldo` | lihat saldo |
| `/topup` | tampilkan tombol nominal (25rb / 50rb / 100rb / 200rb) |
| `/topup 50000` | langsung buat invoice nominal tertentu |
| `/ikat 482913` | hubungkan email/WhatsApp ke akun ini |
| `/bantuan` | daftar perintah + arahkan ke bot toko |

Tombol inline juga tersedia: **Cek Saldo** dan **Top Up**.

## Jalan lokal

```bash
cp .env.example .env      # isi BOT_TOKEN, PANWALLET_URL, STORE_BOT
go build -o bot.exe ./cmd/bot
./bot.exe
```

Prasyarat: Pan-Wallet sudah jalan di `PANWALLET_URL`.

## Struktur

```text
cmd/bot/            entrypoint, long polling Telegram + tombol inline
internal/bot/       handler perintah, callback tombol, format pesan
internal/walletapi/ klien HTTP ke Pan-Wallet
internal/store/     sesi public_id per chat (memori)
internal/config/    baca environment
```

## Test

```bash
go test ./...
```

## Untuk bot toko (AgenPulsa, Xray, dll)

Bot toko **tidak boleh** memproses top-up. Alurnya:

1. Simpan `public_id` pelanggan setelah `/start` pertama.
2. Sebelum order, panggil `GET /v1/public/{public_id}`.
3. Kalau `available` kurang, balas:
   `Saldo kurang. Top up di @PanWalletBot` + tombol link ke bot saldo.
4. Jangan lanjutkan order.

Endpoint cek saldo: lihat `docs/API.md` di repo Pan-Wallet, bagian 7.
