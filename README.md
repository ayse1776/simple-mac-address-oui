# simple-mac-address-oui

Resolve the organizationally unique identifier (OUI) portion of a MAC address against an in-memory table of IEEE vendor assignments.

```go
package main

import (
	"fmt"
	"log"

	oui "simple-mac-address-oui"
)

func main() {
	table := oui.NewTable()
	if err := table.Register("00:11:22:00:00:00", "Acme Corp"); err != nil {
		log.Fatal(err)
	}

	vendor, err := table.Lookup("00:11:22:33:44:55")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(vendor) // Acme Corp
}
```

## Why this exists

The full IEEE OUI registry is a large file that changes regularly. Applications that only need to recognise a handful of hardware vendors end up shipping that whole registry, or hitting a network service, for a lookup that is really just a prefix match on three bytes. This library keeps the table in memory and lets the caller decide which prefixes matter. The trade-off is explicit: you get a tiny, dependency-free resolver, but you are responsible for keeping the entries you register current.

## The awkward edge

A MAC address identifies a device, but only its first three bytes identify the vendor. Two devices from the same vendor share an OUI and differ in the last three bytes, so `Lookup` ignores the device-specific portion entirely. That is the intended behaviour, but it means a lookup cannot distinguish two devices from the same vendor, and a table entry registered for one device will match every device from that OUI.

## API

- `NewTable() *Table` — create an empty table.
- `(*Table) Register(prefix, vendor string) error` — associate a MAC prefix with a vendor. Accepts any format `ParseMAC` accepts; only the first three bytes are used. Re-registering an OUI overwrites the previous vendor.
- `(*Table) Lookup(mac string) (string, error)` — return the vendor for the OUI of `mac`. Returns `ErrInvalidMAC` for an unparseable address, and a descriptive error when the OUI is not registered.
- `(*Table) Len() int` — number of distinct OUI prefixes registered.
- `ParseMAC(s string) (MAC, error)` — parse a MAC in colon, hyphen, Cisco dotted-quad, or bare 12-digit hex form. Returns `ErrInvalidMAC` on malformed input.
- `MAC` — a `[6]byte` hardware address; `String` formats it as `AA:BB:CC:DD:EE:FF`.
- `OUI` — a `uint32` holding the 24-bit prefix; `String` formats it as six uppercase hex digits.
- `ErrInvalidMAC` — sentinel error for malformed MAC input.
