package bot

import (
	"context"
	"testing"
)

func TestCallbackSaldoReturnsBalance(t *testing.T) {
	api := &fakeAPI{resolved: Wallet{PublicID: "CBTEST01", Balance: 75000, Available: 75000}}
	handler := newHandler(api)

	reply := handler.HandleCallback(context.Background(), 1, "123456789", "saldo")

	if !contains(reply.Text, "CBTEST01") || !contains(reply.Text, "75.000") {
		t.Fatalf("callback saldo salah: %s", reply.Text)
	}
}

func TestCallbackTopupMenuShowsNominals(t *testing.T) {
	handler := newHandler(&fakeAPI{resolved: Wallet{PublicID: "X"}})

	reply := handler.HandleCallback(context.Background(), 1, "123456789", "topupmenu")

	if reply.Keyboard == nil {
		t.Fatal("harus mengirim tombol nominal")
	}
	if !contains(reply.Text, "Pilih nominal") {
		t.Fatalf("teks = %q", reply.Text)
	}
}

func TestCallbackTopupAmountCreatesInvoice(t *testing.T) {
	api := &fakeAPI{
		resolved: Wallet{PublicID: "CBTEST01"},
		topup:    Topup{Total: 50001, PayURL: "https://pay/xyz"},
	}
	handler := newHandler(api)

	reply := handler.HandleCallback(context.Background(), 1, "123456789", "topup:50000")

	if api.lastTopup != 50000 {
		t.Fatalf("nominal = %d, want 50000", api.lastTopup)
	}
	if !contains(reply.Text, "https://pay/xyz") {
		t.Fatalf("harus mengirim link bayar: %s", reply.Text)
	}
}

func TestTopupWithoutArgsShowsButtons(t *testing.T) {
	handler := newHandler(&fakeAPI{resolved: Wallet{PublicID: "X"}})

	reply := handler.Handle(context.Background(), 1, "123456789", "/topup")

	if reply.Keyboard == nil {
		t.Fatal("/topup tanpa argumen harus menampilkan tombol")
	}
}

func TestStoreBotMentionedInHelp(t *testing.T) {
	handler := &Handler{API: &fakeAPI{resolved: Wallet{PublicID: "X"}}, Sessions: newMemorySessions(), StoreBot: "AgenPulsaBot"}

	reply := handler.Handle(context.Background(), 1, "123456789", "/bantuan")

	if !contains(reply.Text, "@AgenPulsaBot") {
		t.Fatalf("bantuan harus menyebut bot toko: %s", reply.Text)
	}
}
