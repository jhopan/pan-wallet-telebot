# AGENTS.md

Bot Telegram untuk Pan-Wallet. Bot ini **klien** Pan-Wallet, bukan pemegang saldo.
Semua data saldo ada di service Pan-Wallet; bot hanya menampilkan dan meneruskan perintah.

## Ruang lingkup

Termasuk: repo ini, kode bot, handler perintah Telegram, klien HTTP ke Pan-Wallet, penyimpanan sesi lokal.

Tidak termasuk: logika saldo, ledger, top-up, pembayaran. Itu milik repo `Pan-Wallet`.
Jangan pernah menghitung atau menyimpan saldo di repo ini.

## Aturan arsitektur

- Bot tidak punya database saldo. Kalau butuh menyimpan apa pun, simpan hanya
  penanda sesi (langkah percakapan atau public_id), bukan angka uang.
- Identitas pelanggan = `public_id` 8 karakter dari Pan-Wallet, bukan Telegram ID.
  Telegram ID hanya dipakai untuk memanggil `POST /v1/identity/resolve`.
- Semua nominal dikirim sebagai integer rupiah. Jangan pakai float.
- Bot tidak memegang API key produk. Bot memakai endpoint publik Pan-Wallet
  (`/v1/public/*`, `/v1/identity/*`) yang tidak butuh token.
- Jangan menampilkan daftar ID pelanggan. ID itu rahasia pelanggan.

## Endpoint Pan-Wallet yang dipakai

| Kebutuhan | Endpoint |
|---|---|
| Pelanggan chat pertama kali | `POST /v1/identity/resolve` |
| Cek saldo | `GET /v1/public/{public_id}` |
| Buat top-up | `POST /v1/public/{public_id}/topup` |
| Ikat akun lain | `POST /v1/identity/link/confirm` |
| Lihat channel tertaut | `GET /v1/public/{public_id}/identities` |

Dokumentasi lengkap: repo Pan-Wallet, `docs/API.md`.

## Alur wajib

1. Setiap pesan masuk, panggil `identity/resolve` dengan `telegram_id` pengirim.
2. Simpan `public_id` hasilnya di sesi.
3. Kalau `is_new` true, sapa pelanggan dan tampilkan ID-nya.
4. Perintah lain memakai `public_id` dari sesi, bukan dari input pengguna.
5. Perintah `/ikat <kode>` meneruskan kode ke `identity/link/confirm`.

## Runtime

- Go 1.26.
- Konfigurasi lewat environment: `BOT_TOKEN`, `PANWALLET_URL`.
- Jangan menaruh token bot di source, tes, atau commit.
- Perintah lokal: `go test ./...`, `go build ./cmd/bot`.

## Keamanan

- Token bot Telegram hanya di `.env` lokal atau environment service.
- `.env` tidak boleh masuk git.
- Jangan log isi pesan pelanggan lengkap (bisa memuat data pribadi).
- Batasi panjang pesan keluar; Telegram memotong di 4096 karakter.

## Git

- Jangan commit `.env`, berkas sesi, atau data pelanggan.
- Jangan commit atau push tanpa diminta.
