package mfscli

import (
	"flag"
	"github.com/golang/glog"
	"testing"
)

func init() {
	flag.Set("logtostderr", "true")
	flag.Parse()
}

func Test_glog(t *testing.T) {
	glog.Info("for test")
}

func Test_packString(t *testing.T) {
	s := Pack("123", []byte("456"))
	if string(s) != "123456" {
		t.Error("error pack string")
	}
}

func Test_pack(t *testing.T) {
	s := Pack(uint8(1), uint16(2), uint32(3), uint64(4), "456")
	var t1 uint8
	var t2 uint16
	var t3 uint32
	var t4 uint64
	t5 := make([]byte, len("456"))
	UnPack(s, &t1, &t2, &t3, &t4, t5)
	if t1 != 1 {
		t.Error("error unpack uint8", t1)
	}
	if t2 != 2 {
		t.Error("error unpack uint16", t2)
	}
	if t3 != 3 {
		t.Error("error unpack uint32", t3)
	}
	if t4 != 4 {
		t.Error("error unpack uint64", t4)
	}
	if string(t5) != "456" {
		t.Error("error unpack string")
	}
}

func Test_packCmd(t *testing.T) {
	s := PackCmd(88, uint8(1), uint16(2), uint32(3), uint64(4), "456")
	var t1 uint8
	var t2 uint16
	var t3 uint32
	var t4 uint64
	t5 := make([]byte, len("456"))
	cmd, err := UnPackCmd(s, &t1, &t2, &t3, &t4, t5)
	if err != nil {
		t.Fatal("error UnPackCmd", err)
	}
	if cmd != 88 {
		t.Error("error UnPackCmd cmd", cmd)
	}
	if t1 != 1 {
		t.Error("error UnPackCmd uint8", t1)
	}
	if t2 != 2 {
		t.Error("error UnPackCmd uint16", t2)
	}
	if t3 != 3 {
		t.Error("error UnPackCmd uint32", t3)
	}
	if t4 != 4 {
		t.Error("error UnPackCmd uint64", t4)
	}
	if string(t5) != "456" {
		t.Error("error UnPackCmd string")
	}
}
