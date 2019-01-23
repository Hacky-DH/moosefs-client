package mfscli

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
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

// 2.45M  -> 2450000
// 2.45Mi -> 2569011
func ParseBytes(str string) (uint64, error) {
	lastDigit := 0
	str = strings.ToUpper(strings.TrimSpace(str))
	for _, r := range str {
		if !(unicode.IsDigit(r) || r == '.') {
			break
		}
		lastDigit++
	}
	f, err := strconv.ParseFloat(str[:lastDigit], 64)
	if err != nil {
		return 0, err
	}
	const suff = "BKMGTPEZY"
	var unit float64 = 1000
	index := -1
	switch len(str) {
	case lastDigit:
		return uint64(f), nil
	case lastDigit + 2:
		if str[lastDigit+1] != 'I' {
			goto out
		}
		unit = 1024
		fallthrough
	case lastDigit + 1:
		index = strings.IndexRune(suff, rune(str[lastDigit]))
		if index == -1 {
			goto out
		}
		f *= math.Pow(unit, float64(index))
		return uint64(f), nil
	}
out:
	return 0, fmt.Errorf("Unrecognized byte str %s", str)
}
