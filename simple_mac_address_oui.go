// Package simple_mac_address_oui resolves the organizationally unique identifier
// (OUI) portion of a MAC address against an in-memory IEEE assignment table.
//
// The table is intentionally small and self-contained: this library is meant to
// answer "who owns this prefix?" for the handful of vendors a caller actually
// cares about, without pulling in a multi-megabyte registry or any third-party
// dependency. Callers register the prefixes they need, then look them up.
package simple_mac_address_oui

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidMAC is returned by ParseMAC and Lookup when the input is not a
// well-formed 48-bit MAC address in one of the accepted formats.
var ErrInvalidMAC = errors.New("invalid MAC address")

// OUI is the 24-bit organizationally unique identifier assigned by the IEEE.
// It is stored in the low three bytes of the uint32; the high byte is always
// zero so that comparisons and lookups are simple integer operations.
type OUI uint32

// MAC is a parsed 48-bit hardware address. The six bytes are kept in
// transmission order, so MAC[0] is the most significant octet.
type MAC [6]byte

// Table maps OUI prefixes to vendor names. The zero value is not usable;
// construct one with NewTable.
type Table struct {
	entries map[OUI]string
}

// NewTable returns an empty Table ready for Register calls.
func NewTable() *Table {
	return &Table{entries: make(map[OUI]string)}
}

// Register associates the given MAC prefix with a vendor name. The prefix is
// parsed with ParseMAC, so any of the accepted textual forms work, but only the
// first three bytes are used as the OUI key. Registering the same OUI twice
// overwrites the previous name, which lets callers correct a bad entry without
// rebuilding the table.
func (t *Table) Register(prefix, vendor string) error {
	mac, err := ParseMAC(prefix)
	if err != nil {
		return err
	}
	t.entries[ouiFromMAC(mac)] = vendor
	return nil
}

// Lookup returns the vendor name registered for the OUI of mac. It returns
// ErrInvalidMAC if mac cannot be parsed and a plain error if no vendor is
// registered for the prefix. The error is deliberately not a sentinel because
// callers usually only care whether the lookup succeeded.
func (t *Table) Lookup(mac string) (string, error) {
	parsed, err := ParseMAC(mac)
	if err != nil {
		return "", err
	}
	vendor, ok := t.entries[ouiFromMAC(parsed)]
	if !ok {
		return "", fmt.Errorf("no vendor registered for OUI %s", ouiFromMAC(parsed).String())
	}
	return vendor, nil
}

// Len reports how many distinct OUI prefixes are registered.
func (t *Table) Len() int {
	return len(t.entries)
}

// ouiFromMAC extracts the 24-bit OUI from a parsed MAC address.
func ouiFromMAC(m MAC) OUI {
	return OUI(uint32(m[0])<<16 | uint32(m[1])<<8 | uint32(m[2]))
}

// String formats the OUI as six uppercase hex digits, e.g. "001122".
func (o OUI) String() string {
	return fmt.Sprintf("%06X", uint32(o))
}

// String formats the MAC as six colon-separated uppercase hex octets,
// e.g. "00:11:22:33:44:55".
func (m MAC) String() string {
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", m[0], m[1], m[2], m[3], m[4], m[5])
}

// ParseMAC accepts the common textual MAC forms: colon-separated, hyphen-
// separated, Cisco-style dotted quads (aabb.ccdd.eeff), and bare 12-digit hex.
// It rejects anything that does not resolve to exactly six bytes. The
// separator is normalised to colons before splitting, which keeps the parser
// small and means callers do not have to care which convention their data uses.
func ParseMAC(s string) (MAC, error) {
	var m MAC

	s = strings.TrimSpace(s)
	if s == "" {
		return m, ErrInvalidMAC
	}

	// Normalise separators. Dots are only meaningful in the Cisco form, so we
	// strip them and treat the remaining digits as a contiguous 12-char hex
	// string; colons and hyphens become the canonical colon separator.
	normalised := strings.ReplaceAll(s, ".", "")
	normalised = strings.ReplaceAll(normalised, "-", ":")

	var parts []string
	if strings.Contains(normalised, ":") {
		parts = strings.Split(normalised, ":")
	} else {
		// Bare hex, possibly with no separators at all.
		if len(normalised) != 12 {
			return m, ErrInvalidMAC
		}
		parts = make([]string, 6)
		for i := 0; i < 6; i++ {
			parts[i] = normalised[i*2 : i*2+2]
		}
	}

	if len(parts) != 6 {
		return m, ErrInvalidMAC
	}

	for i, part := range parts {
		if len(part) != 2 {
			return m, ErrInvalidMAC
		}
		b, err := parseHexByte(part)
		if err != nil {
			return m, ErrInvalidMAC
		}
		m[i] = b
	}
	return m, nil
}

// parseHexByte converts exactly two hex digits to a byte. It is stricter than
// strconv.ParseUint because it rejects the "0x" prefix and signs that
// ParseUint would otherwise accept, which matters for a fixed-width field.
func parseHexByte(s string) (byte, error) {
	if len(s) != 2 {
		return 0, ErrInvalidMAC
	}
	var v byte
	for i := 0; i < 2; i++ {
		c := s[i]
		var d byte
		switch {
		case c >= '0' && c <= '9':
			d = c - '0'
		case c >= 'a' && c <= 'f':
			d = c - 'a' + 10
		case c >= 'A' && c <= 'F':
			d = c - 'A' + 10
		default:
			return 0, ErrInvalidMAC
		}
		v = v<<4 | d
	}
	return v, nil
}
