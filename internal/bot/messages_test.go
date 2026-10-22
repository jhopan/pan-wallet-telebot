package bot

import "testing"

func TestParseCommand(t *testing.T) {
	cases := []struct {
		input   string
		command string
		args    string
	}{
		{"/start", "start", ""},
		{"/saldo", "saldo", ""},
		{"/ikat 482913", "ikat", "482913"},
		{"/topup 50000", "topup", "50000"},
		{"/SALDO", "saldo", ""},
		{"halo", "", ""},
		{"", "", ""},
	}

	for _, c := range cases {
		command, args := ParseCommand(c.input)
		if command != c.command {
			t.Errorf("ParseCommand(%q) command = %q, want %q", c.input, command, c.command)
		}
		if args != c.args {
			t.Errorf("ParseCommand(%q) args = %q, want %q", c.input, args, c.args)
		}
	}
}

func TestFormatRupiah(t *testing.T) {
	cases := map[int64]string{
		0:       "0",
		1000:    "1.000",
		25000:   "25.000",
		150001:  "150.001",
		1000000: "1.000.000",
	}
	for input, want := range cases {
		if got := FormatRupiah(input); got != want {
			t.Errorf("FormatRupiah(%d) = %q, want %q", input, got, want)
		}
	}
}

func TestFormatBalanceMessage(t *testing.T) {
	text := BalanceMessage("JKCUX8CK", 150001, 10000, 140001)

	for _, want := range []string{"JKCUX8CK", "150.001", "10.000", "140.001"} {
		if !contains(text, want) {
			t.Errorf("pesan saldo tidak memuat %q:\n%s", want, text)
		}
	}
}

func TestFormatTopupMessage(t *testing.T) {
	text := TopupMessage(25001, "https://paypan.example/pay/abc", 15)

	for _, want := range []string{"25.001", "https://paypan.example/pay/abc", "15"} {
		if !contains(text, want) {
			t.Errorf("pesan topup tidak memuat %q:\n%s", want, text)
		}
	}
}

func TestHelpMessageListsCommands(t *testing.T) {
	text := HelpMessage()
	for _, want := range []string{"/saldo", "/topup", "/ikat"} {
		if !contains(text, want) {
			t.Errorf("bantuan tidak memuat %q", want)
		}
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
