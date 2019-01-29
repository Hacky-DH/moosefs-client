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
	t.Skip()
	c := client()
	defer c.Close()
	info := &QuotaInfo{
		inode: 13,
	}
	var err error
	err = c.QuotaControl(info, QuotaGet)
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
		err := c.QuotaControl(info, QuotaSet)
		if err != nil {
			t.Fatal(err)
		}
		info.slength = 0
		err = c.QuotaControl(info, QuotaGet)
		if err != nil {
			t.Fatal(err)
		}
		str := FormatBytes(float64(info.slength), Binary)
		if str != "1.00 TiB" {
			t.Fatal("unexpect ", str)
		}
		err = c.QuotaControl(info, QuotaDel)
		if err != nil {
			t.Fatal(err)
		}
		if info.slength != 0 {
			t.Fatal("unexpect ", info.slength)
		}
	})
}

func TestStatfs(t *testing.T) {
	t.Skip()
	session(t, func(c *Client) {
		_, err := c.Statfs()
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestAccess(t *testing.T) {
	t.Skip()
	session(t, func(c *Client) {
		err := c.Access(1, 0777)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestLookup(t *testing.T) {
	t.Skip()
	session(t, func(c *Client) {
		_, err := c.Lookup(1, ".")
		if err != nil {
			t.Fatal(err)
		}
		_, err = c.Lookup(1, "notexist")
		if err == nil {
			t.Fatal("unexpect")
		}
	})
}

func TestMkdir(t *testing.T) {
	t.Skip()
	session(t, func(c *Client) {
		n := "testdir"
		c.Rmdir(1, n)
		_, err := c.Mkdir(1, n, 0755)
		if err != nil {
			t.Fatal(err)
		}
		_, err = c.Lookup(1, n)
		if err != nil {
			t.Fatal(err)
		}
		c.Rmdir(1, n)
	})
}

func TestMknod(t *testing.T) {
	t.Skip()
	session(t, func(c *Client) {
		n := "testfile"
		c.Unlink(1, n)
		_, err := c.Mknod(1, n, 0744)
		if err != nil {
			t.Fatal(err)
		}
		_, err = c.Lookup(1, n)
		if err != nil {
			t.Fatal(err)
		}
		c.Unlink(1, n)
	})
}

func TestReaddir(t *testing.T) {
	t.Skip()
	session(t, func(c *Client) {
		_, err := c.Readdir(1)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestCreate(t *testing.T) {
	t.Skip()
	session(t, func(c *Client) {
		n := "testfile"
		c.Unlink(1, n)
		fi, err := c.Create(1, n, 0744)
		if err != nil {
			t.Fatal(err)
		}
		err = c.Open(fi.Inode, 7)
		if err != nil {
			t.Fatal(err)
		}
		c.Unlink(1, n)
	})
}

func TestAttr(t *testing.T) {
	t.Skip()
	session(t, func(c *Client) {
		n := "testfile"
		c.Unlink(1, n)
		fi, err := c.Create(1, n, 0744)
		if err != nil {
			t.Fatal(err)
		}
		_, err = c.GetAttr(fi.Inode)
		if err != nil {
			t.Fatal(err)
		}
		_, err = c.Chmod(fi.Inode, 0766)
		if err != nil {
			t.Fatal(err)
		}
		c.Unlink(1, n)
	})
}

func TestRemoveAllSession(t *testing.T) {
	t.Skip()
	c := client()
	defer c.Close()
	sess, _ := c.ListSession()
	for _, s := range sess {
		c.RemoveSession(s)
	}
}

func TestPurge(t *testing.T) {
	t.Skip()
	session(t, func(c *Client) {
		n := "testfile"
		c.Unlink(1, n)
		fi, err := c.Create(1, n, 0744)
		if err != nil {
			t.Fatal(err)
		}
		err = c.Purge(fi.Inode)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestReaddirAttr(t *testing.T) {
	t.Skip()
	session(t, func(c *Client) {
		_, err := c.ReaddirAttr(1)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestDirStats(t *testing.T) {
	t.Skip()
	session(t, func(c *Client) {
		_, err := c.GetDirStats(1)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestRWChunk(t *testing.T) {
	t.Skip()
	session(t, func(c *Client) {
		_, err := c.ReadChunk(54, 0, CHUNKOPFLAG_CANMODTIME)
		if err != nil {
			t.Fatal(err)
		}
		cs, err := c.WriteChunk(54, 0, CHUNKOPFLAG_CANMODTIME)
		if err != nil {
			t.Fatal(err)
		}
		err = c.WriteChunkEnd(cs.ChunkId, 54, 0, cs.Length,
			CHUNKOPFLAG_CANMODTIME)
		if err != nil {
			t.Fatal(err)
		}
	})
}
