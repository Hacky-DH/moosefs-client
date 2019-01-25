package mfscli

import (
	"flag"
	"testing"
	"time"
)

func init() {
	flag.Set("logtostderr", "true")
	flag.Set("v", "10")
	flag.Parse()
}

func client() *Client {
	return NewClientPwd("127.0.0.1", "password", false)
}

func session(t *testing.T, cb func(*Client)) {
	c := client()
	defer c.Close()
	err := c.CreateSession()
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseSession()
	cb(c)
}

func TestConnect(t *testing.T) {
	t.Skip()
	c := client()
	if err := c.Connect(); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second)
	c.Close()
}

func TestQuotaControlNoSession(t *testing.T) {
	c := client()
	defer c.Close()
	info := &QuotaInfo{
		inode: 13,
	}
	var err error
	err = c.QuotaControl(info, quotaGet)
	if err == nil {
		t.Error("unexpected")
	}
}

func TestQuotaControl(t *testing.T) {
	t.Skip()
	session(t, func(c *Client) {
		s, _ := ParseBytes("1Ti")
		info := &QuotaInfo{
			inode:   13,
			slength: s,
		}
		err := c.QuotaControl(info, quotaSet)
		if err != nil {
			t.Fatal(err)
		}
		info.slength = 0
		err = c.QuotaControl(info, quotaGet)
		if err != nil {
			t.Fatal(err)
		}
		str := FormatBytes(float64(info.slength), Binary)
		if str != "1.00 TiB" {
			t.Fatal("unexpect ", str)
		}
		err = c.QuotaControl(info, quotaDel)
		if err != nil {
			t.Fatal(err)
		}
		if info.slength != 0 {
			t.Fatal("unexpect ", info.slength)
		}
	})
}
