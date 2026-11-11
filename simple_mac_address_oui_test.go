package simple_mac_address_oui

import (
	"errors"
	"testing"
)

func TestParseMACAcceptedForms(t *testing.T) {
	want := MAC{0x00, 0x11, 0x22, 0x33, 0x44, 0x55}
	cases := []struct {
		name string
		in   string
	}{
		{"colon uppercase", "00:11:22:33:44:55"},
		{"colon lowercase", "00:11:22:33:44:55"},
		{"hyphen", "00-11-22-33-44-55"},
		{"cisco dotted", "0011.2233.4455"},
		{"bare hex", "001122334455"},
		{"surrounding whitespace", "  00:11:22:33:44:55  "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseMAC(tc.in)
			if err != nil {
				t.Fatalf("ParseMAC(%q) returned error: %v", tc.in, err)
			}
			if got != want {
				t.Fatalf("ParseMAC(%q) = %v, want %v", tc.in, got, want)
			}
		})
	}
}

func TestParseMACRejectsMalformed(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"too few octets", "00:11:22:33:44"},
		{"too many octets", "00:11:22:33:44:55:66"},
		{"short octet", "0:11:22:33:44:55"},
		{"long octet", "000:11:22:33:44:55"},
		{"non-hex digit", "00:11:22:33:44:5g"},
		{"bare hex wrong length", "0011223344"},
		{"bare hex odd length", "00112233445"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseMAC(tc.in)
			if !errors.Is(err, ErrInvalidMAC) {
				t.Fatalf("ParseMAC(%q) error = %v, want ErrInvalidMAC", tc.in, err)
			}
		})
	}
}

func TestMACString(t *testing.T) {
	m := MAC{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	if got, want := m.String(), "AA:BB:CC:DD:EE:FF"; got != want {
		t.Fatalf("MAC.String() = %q, want %q", got, want)
	}
}

func TestOUIString(t *testing.T) {
	o := OUI(0x001122)
	if got, want := o.String(), "001122"; got != want {
		t.Fatalf("OUI.String() = %q, want %q", got, want)
	}
}

func TestRegisterAndLookup(t *testing.T) {
	tbl := NewTable()
	if err := tbl.Register("00:11:22:33:44:55", "Acme Corp"); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if err := tbl.Register("AA-BB-CC-00-00-01", "Globex"); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if got, err := tbl.Lookup("00:11:22:99:88:77"); err != nil || got != "Acme Corp" {
		t.Fatalf("Lookup = (%q, %v), want (Acme Corp, nil)", got, err)
	}
	if got, err := tbl.Lookup("aabbccffeedd"); err != nil || got != "Globex" {
		t.Fatalf("Lookup = (%q, %v), want (Globex, nil)", got, err)
	}
	if tbl.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", tbl.Len())
	}
}

func TestRegisterOverwrites(t *testing.T) {
	tbl := NewTable()
	if err := tbl.Register("00:11:22:00:00:00", "First"); err != nil {
		t.Fatal(err)
	}
	if err := tbl.Register("00:11:22:ff:ff:ff", "Second"); err != nil {
		t.Fatal(err)
	}
	if tbl.Len() != 1 {
		t.Fatalf("Len() = %d, want 1 after overwrite", tbl.Len())
	}
	got, err := tbl.Lookup("00:11:22:12:34:56")
	if err != nil || got != "Second" {
		t.Fatalf("Lookup = (%q, %v), want (Second, nil)", got, err)
	}
}

func TestRegisterRejectsInvalidPrefix(t *testing.T) {
	tbl := NewTable()
	if err := tbl.Register("not-a-mac", "Nope"); !errors.Is(err, ErrInvalidMAC) {
		t.Fatalf("Register error = %v, want ErrInvalidMAC", err)
	}
	if tbl.Len() != 0 {
		t.Fatalf("Len() = %d, want 0 after failed register", tbl.Len())
	}
}

func TestLookupUnknownOUI(t *testing.T) {
	tbl := NewTable()
	if err := tbl.Register("00:11:22:00:00:00", "Acme"); err != nil {
		t.Fatal(err)
	}
	_, err := tbl.Lookup("de:ad:be:ef:00:01")
	if err == nil {
		t.Fatal("Lookup of unregistered OUI returned nil error")
	}
}

func TestLookupInvalidMAC(t *testing.T) {
	tbl := NewTable()
	_, err := tbl.Lookup("zz:zz:zz:zz:zz:zz")
	if !errors.Is(err, ErrInvalidMAC) {
		t.Fatalf("Lookup error = %v, want ErrInvalidMAC", err)
	}
}

func TestLookupOnlyUsesFirstThreeBytes(t *testing.T) {
	tbl := NewTable()
	if err := tbl.Register("00:11:22:00:00:00", "Acme"); err != nil {
		t.Fatal(err)
	}
	for _, mac := range []string{
		"00:11:22:00:00:01",
		"00:11:22:ff:ff:fe",
		"00:11:22:12:34:56",
	} {
		got, err := tbl.Lookup(mac)
		if err != nil || got != "Acme" {
			t.Fatalf("Lookup(%q) = (%q, %v), want (Acme, nil)", mac, got, err)
		}
	}
}
