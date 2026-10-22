package bot

import "testing"

func TestWelcomeMentionsBalanceBotRole(t *testing.T) {
	text := WelcomeMessage("JKCUX8CK", true)

	for _, want := range []string{"JKCUX8CK", "saldo", "top-up"} {
		if !contains(text, want) {
			t.Errorf("sapaan tidak memuat %q:\n%s", want, text)
		}
	}
	// Harus jelas ini bot saldo, bukan bot toko.
	if !contains(text, "Bot Saldo") {
		t.Errorf("harus menyebut dirinya Bot Saldo:\n%s", text)
	}
}

func TestHelpRedirectsToStoreBot(t *testing.T) {
	text := HelpMessageWithStore("AgenPulsaBot")

	if !contains(text, "@AgenPulsaBot") {
		t.Errorf("bantuan harus menyebut bot toko:\n%s", text)
	}
	if !contains(text, "top-up") && !contains(text, "topup") {
		t.Errorf("bantuan harus menyebut top-up:\n%s", text)
	}
}

func TestHelpWithoutStoreBotStillWorks(t *testing.T) {
	text := HelpMessageWithStore("")
	if text == "" {
		t.Fatal("bantuan tidak boleh kosong")
	}
	if contains(text, "@") {
		t.Errorf("tanpa bot toko tidak boleh ada mention:\n%s", text)
	}
}

func TestTopupKeyboardHasNominalButtons(t *testing.T) {
	keyboard := TopupKeyboard()

	if len(keyboard.Rows) == 0 {
		t.Fatal("keyboard harus punya tombol")
	}
	var labels []string
	for _, row := range keyboard.Rows {
		for _, button := range row {
			labels = append(labels, button.Text)
		}
	}
	joined := ""
	for _, label := range labels {
		joined += label + " "
	}
	for _, want := range []string{"10.000", "25.000", "50.000", "100.000"} {
		if !contains(joined, want) {
			t.Errorf("tombol tidak memuat %q: %s", want, joined)
		}
	}
}

func TestTopupKeyboardCallbackData(t *testing.T) {
	keyboard := TopupKeyboard()
	for _, row := range keyboard.Rows {
		for _, button := range row {
			if button.Callback == "" {
				t.Errorf("tombol %q tanpa callback", button.Text)
			}
			if len(button.Callback) > 64 {
				t.Errorf("callback terlalu panjang: %q", button.Callback)
			}
		}
	}
}

func TestParseCallbackAmount(t *testing.T) {
	amount, ok := ParseCallbackAmount("topup:50000")
	if !ok || amount != 50000 {
		t.Fatalf("amount = %d, ok=%v", amount, ok)
	}

	for _, bad := range []string{"", "topup:", "saldo:1", "topup:abc", "topup:0", "topup:500"} {
		if _, ok := ParseCallbackAmount(bad); ok {
			t.Errorf("%q seharusnya ditolak", bad)
		}
	}
}

func TestPaidMessageShowsNewBalance(t *testing.T) {
	text := PaidMessage(50001, 150001)
	for _, want := range []string{"50.001", "150.001"} {
		if !contains(text, want) {
			t.Errorf("pesan lunas tidak memuat %q:\n%s", want, text)
		}
	}
}
