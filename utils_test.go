package mfscli

import (
	"testing"
)

func TestFormatBytes(t *testing.T) {
	if "999.00 B" != FormatBytes(999, Binary) {
		t.Error("expect error")
	}
	if "1000.00 B" != FormatBytes(1000, Binary) {
		t.Error("expect error")
	}
	if "1.00 KiB" != FormatBytes(1024, Binary) {
		t.Error("expect error")
	}
	if "1.00 TiB" != FormatBytes(1099511627776, Binary) {
		t.Error("expect error")
	}
	if "909.49 TiB" != FormatBytes(1000000000000000, Binary) {
		t.Error("expect error")
	}
	if "1.00 PiB" != FormatBytes(1125899906842624, Binary) {
		t.Error("expect error")
	}
	if "999.00 bit" != FormatBytes(999, Decimal) {
		t.Error("expect error")
	}
	if "1.00 kbit" != FormatBytes(1000, Decimal) {
		t.Error("expect error")
	}
	if "1.02 kbit" != FormatBytes(1024, Decimal) {
		t.Error("expect error")
	}
	if "1.10 Tbit" != FormatBytes(1099511627776, Decimal) {
		t.Error("expect error")
	}
	if "1.00 Pbit" != FormatBytes(1000000000000000, Decimal) {
		t.Error("expect error")
	}
	if "1.13 Pbit" != FormatBytes(1125899906842624, Decimal) {
		t.Error("expect error")
	}
}
