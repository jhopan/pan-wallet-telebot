package bot

import (
	"context"
	"testing"
)

type fakeAPI struct {
	resolved   Wallet
	topup      Topup
	lastTopup  int64
	lastCode   string
	resolveErr error
}

func (api *fakeAPI) Resolve(context.Context, string) (Wallet, error) {
	return api.resolved, api.resolveErr
}
func (api *fakeAPI) Balance(context.Context, string) (Wallet, error) { return api.resolved, nil }
func (api *fakeAPI) CreateTopup(_ context.Context, _ string, amount int64) (Topup, error) {
	api.lastTopup = amount
	return api.topup, nil
}
func (api *fakeAPI) LinkConfirm(_ context.Context, code, _ string) (Wallet, error) {
	api.lastCode = code
	return api.resolved, nil
}

func newHandler(api WalletAPI) *Handler {
	return &Handler{API: api, Sessions: newMemorySessions()}
}

// newMemorySessions adalah implementasi Sessions sederhana untuk uji.
type memorySessions struct{ data map[int64]string }

func newMemorySessions() *memorySessions                            { return &memorySessions{data: map[int64]string{}} }
func (s *memorySessions) SetPublicID(chatID int64, publicID string) { s.data[chatID] = publicID }
func (s *memorySessions) PublicID(chatID int64) (string, bool) {
	v, ok := s.data[chatID]
	return v, ok
}

func TestStartWelcomesNewCustomer(t *testing.T) {
	api := &fakeAPI{resolved: Wallet{PublicID: "JKCUX8CK", IsNew: true}}
	reply := newHandler(api).Handle(context.Background(), 1, "123456789", "/start")

	if !contains(reply.Text, "JKCUX8CK") {
		t.Fatalf("sapaan tidak memuat ID: %s", reply.Text)
	}
	if !contains(reply.Text, "dibuat") {
		t.Fatalf("harus menyebut akun dibuat: %s", reply.Text)
	}
}

func TestSaldoShowsAvailableAmount(t *testing.T) {
	api := &fakeAPI{resolved: Wallet{PublicID: "ABC12345", Balance: 150001, Held: 10000, Available: 140001}}
	reply := newHandler(api).Handle(context.Background(), 1, "123456789", "/saldo")

	for _, want := range []string{"ABC12345", "150.001", "140.001"} {
		if !contains(reply.Text, want) {
			t.Fatalf("pesan saldo tidak memuat %q: %s", want, reply.Text)
		}
	}
}

func TestTopupValidatesAmount(t *testing.T) {
	api := &fakeAPI{resolved: Wallet{PublicID: "ABC12345"}, topup: Topup{Total: 50001, PayURL: "https://pay/abc"}}
	handler := newHandler(api)

	reply := handler.Handle(context.Background(), 1, "123456789", "/topup 500")
	if contains(reply.Text, "https://pay/abc") {
		t.Fatal("nominal di bawah 1000 tidak boleh membuat invoice")
	}
	if api.lastTopup != 0 {
		t.Fatalf("API tidak boleh dipanggil, dapat %d", api.lastTopup)
	}

	reply = handler.Handle(context.Background(), 1, "123456789", "/topup 50000")
	if !contains(reply.Text, "https://pay/abc") {
		t.Fatalf("harus mengirim link bayar: %s", reply.Text)
	}
	if api.lastTopup != 50000 {
		t.Fatalf("nominal terkirim = %d, want 50000", api.lastTopup)
	}
}

func TestIkatForwardsCode(t *testing.T) {
	api := &fakeAPI{resolved: Wallet{PublicID: "ABC12345"}}
	reply := newHandler(api).Handle(context.Background(), 1, "123456789", "/ikat 482913")

	if api.lastCode != "482913" {
		t.Fatalf("kode diteruskan = %q, want 482913", api.lastCode)
	}
	if !contains(reply.Text, "Berhasil") {
		t.Fatalf("harus konfirmasi berhasil: %s", reply.Text)
	}
}

func TestUnknownCommandShowsHelp(t *testing.T) {
	reply := newHandler(&fakeAPI{resolved: Wallet{PublicID: "X"}}).
		Handle(context.Background(), 1, "123456789", "/halo")

	if !contains(reply.Text, "/saldo") {
		t.Fatalf("harus menampilkan bantuan: %s", reply.Text)
	}
}

func TestSessionStoresPublicID(t *testing.T) {
	api := &fakeAPI{resolved: Wallet{PublicID: "STORE123"}}
	handler := newHandler(api)
	handler.Handle(context.Background(), 42, "123456789", "/start")

	publicID, ok := handler.Sessions.PublicID(42)
	if !ok || publicID != "STORE123" {
		t.Fatalf("sesi = %q, ok=%v", publicID, ok)
	}
}
