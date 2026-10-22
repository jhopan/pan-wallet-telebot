package bot

import (
	"context"
	"testing"
)

func TestBalanceMessageHasDisclaimer(t *testing.T) {
	text := BalanceMessage("ABC12345", 50000, 0, 50000)

	if !contains(text, "tidak dapat diuangkan") {
		t.Fatalf("pesan saldo harus menyebut saldo tidak bisa diuangkan:\n%s", text)
	}
}

func TestWelcomeHasMinimumAndDisclaimer(t *testing.T) {
	text := WelcomeMessage("ABC12345", true)

	for _, want := range []string{"25.000", "tidak dapat diuangkan"} {
		if !contains(text, want) {
			t.Errorf("sapaan tidak memuat %q:\n%s", want, text)
		}
	}
}

func TestTopupRejectsBelow25Thousand(t *testing.T) {
	api := &fakeAPI{resolved: Wallet{PublicID: "ABC12345"}, topup: Topup{Total: 25001, PayURL: "https://pay/x"}}
	handler := newHandler(api)

	reply := handler.Handle(context.Background(), 1, "123456789", "/topup 20000")

	if api.lastTopup != 0 {
		t.Fatalf("tidak boleh buat invoice, dapat %d", api.lastTopup)
	}
	if !contains(reply.Text, "25.000") {
		t.Fatalf("harus menyebut minimal 25.000:\n%s", reply.Text)
	}
}

func TestTopupAccepts25Thousand(t *testing.T) {
	api := &fakeAPI{resolved: Wallet{PublicID: "ABC12345"}, topup: Topup{Total: 25001, PayURL: "https://pay/x"}}
	handler := newHandler(api)

	reply := handler.Handle(context.Background(), 1, "123456789", "/topup 25000")

	if api.lastTopup != 25000 {
		t.Fatalf("nominal = %d, want 25000", api.lastTopup)
	}
	if !contains(reply.Text, "https://pay/x") {
		t.Fatalf("harus kirim link: %s", reply.Text)
	}
}

func TestCallbackRejectsBelow25Thousand(t *testing.T) {
	api := &fakeAPI{resolved: Wallet{PublicID: "ABC12345"}}
	handler := newHandler(api)

	handler.HandleCallback(context.Background(), 1, "123456789", "topup:10000")

	if api.lastTopup != 0 {
		t.Fatalf("callback Rp10.000 harus ditolak, dapat %d", api.lastTopup)
	}
}
