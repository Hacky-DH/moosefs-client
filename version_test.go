package mfscli

/*
MIT License

Copyright (c) 2019 DHacky
*/

import (
	"testing"
)

func TestVersion(t *testing.T) {
	v := ParseVersionString(MFS_VERSION)
	if MFS_VERSION != v.String() {
		t.Error("expect", v)
	}
}
