package mfscli

import (
	"fmt"
	"strings"
)

type Mode int

const (
	Binary Mode = iota
	Decimal
)

func FormatBytes(b float64, mode Mode) string {
	units, suff := float64(1024), "B|KiB|MiB|GiB|TiB|PiB|EiB|ZiB|YiB"
	if mode == Decimal {
		units, suff = float64(1000), "bit|kbit|Mbit|Gbit|Tbit|Pbit|Ebit|Zbit|Ybit"
	}
	for _, s := range strings.Split(suff, "|") {
		if b < units {
			return fmt.Sprintf("%.2f %s", b, s)
		}
		b /= units
	}
	return "NotSupported"
}
