package mfscli

/*
MIT License

Copyright (c) 2019 DHacky
*/

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

func maclient() *MAClient {
	return NewMAClientPwd("127.0.0.1", "password", false)
}

func session(t *testing.T, cb func(*MAClient)) {
	c := maclient()
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
	c := maclient()
	if err := c.Connect(); err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second)
	c.Close()
}

func TestQuotaControlNoSession(t *testing.T) {
	t.Skip()
	c := maclient()
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
	session(t, func(c *MAClient) {
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
	session(t, func(c *MAClient) {
		_, err := c.Statfs()
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestAccess(t *testing.T) {
	t.Skip()
	session(t, func(c *MAClient) {
		err := c.Access(MFS_ROOT_ID, 0777)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestLookup(t *testing.T) {
	t.Skip()
	session(t, func(c *MAClient) {
		_, err := c.Lookup(MFS_ROOT_ID, ".")
		if err != nil {
			t.Fatal(err)
		}
		_, err = c.Lookup(MFS_ROOT_ID, "notexist")
		if err == nil {
			t.Fatal("unexpect")
		}
	})
}

func TestMkdir(t *testing.T) {
	t.Skip()
	session(t, func(c *MAClient) {
		n := "testdir"
		c.Rmdir(MFS_ROOT_ID, n)
		_, err := c.Mkdir(MFS_ROOT_ID, n, 0755)
		if err != nil {
			t.Fatal(err)
		}
		_, err = c.Lookup(MFS_ROOT_ID, n)
		if err != nil {
			t.Fatal(err)
		}
		c.Rmdir(MFS_ROOT_ID, n)
	})
}

func TestMknod(t *testing.T) {
	t.Skip()
	session(t, func(c *MAClient) {
		n := "testfile"
		c.Unlink(MFS_ROOT_ID, n)
		_, err := c.Mknod(MFS_ROOT_ID, n, 0744)
		if err != nil {
			t.Fatal(err)
		}
		_, err = c.Lookup(MFS_ROOT_ID, n)
		if err != nil {
			t.Fatal(err)
		}
		c.Unlink(MFS_ROOT_ID, n)
	})
}

func TestReaddir(t *testing.T) {
	t.Skip()
	session(t, func(c *MAClient) {
		_, err := c.Readdir(MFS_ROOT_ID)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestCreate(t *testing.T) {
	t.Skip()
	session(t, func(c *MAClient) {
		n := "testfile"
		c.Unlink(MFS_ROOT_ID, n)
		fi, err := c.Create(MFS_ROOT_ID, n, 0744)
		if err != nil {
			t.Fatal(err)
		}
		_, err = c.Open(fi.Inode, 7)
		if err != nil {
			t.Fatal(err)
		}
		c.Unlink(MFS_ROOT_ID, n)
	})
}

func TestAttr(t *testing.T) {
	t.Skip()
	session(t, func(c *MAClient) {
		n := "testfile"
		c.Unlink(MFS_ROOT_ID, n)
		fi, err := c.Create(MFS_ROOT_ID, n, 0744)
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
		c.Unlink(MFS_ROOT_ID, n)
	})
}

func TestRemoveAllSession(t *testing.T) {
	t.Skip()
	c := maclient()
	defer c.Close()
	sess, _ := c.ListSession()
	for _, s := range sess {
		c.RemoveSession(s)
	}
}

func TestPurge(t *testing.T) {
	t.Skip()
	session(t, func(c *MAClient) {
		n := "testfile"
		c.Unlink(MFS_ROOT_ID, n)
		fi, err := c.Create(MFS_ROOT_ID, n, 0744)
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
	session(t, func(c *MAClient) {
		_, err := c.ReaddirAttr(MFS_ROOT_ID)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestDirStats(t *testing.T) {
	t.Skip()
	session(t, func(c *MAClient) {
		_, err := c.GetDirStats(MFS_ROOT_ID)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestRWChunk(t *testing.T) {
	t.Skip()
	session(t, func(c *MAClient) {
		n := "testfile"
		c.Unlink(MFS_ROOT_ID, n)
		fi, err := c.Create(MFS_ROOT_ID, n, 0744)
		if err != nil {
			t.Fatal(err)
		}
		_, err = c.Open(fi.Inode, 7)
		if err != nil {
			t.Fatal(err)
		}
		cs, err := c.WriteChunk(fi.Inode, 0, CHUNKOPFLAG_CANMODTIME)
		if err != nil {
			t.Fatal(err)
		}
		err = c.WriteChunkEnd(cs.ChunkId, fi.Inode, 0, cs.Length,
			CHUNKOPFLAG_CANMODTIME)
		if err != nil {
			t.Fatal(err)
		}
		_, err = c.ReadChunk(fi.Inode, 0, CHUNKOPFLAG_CANMODTIME)
		if err != nil {
			t.Fatal(err)
		}
		c.Unlink(MFS_ROOT_ID, n)
	})
}

func TestFileInfoSize(t *testing.T) {
	fi := FileInfo{
		Type: TYPE_DIRECTORY,
		Size: 3001298,
	}
	if fi.GetSize() != "12.98 GiB" {
		t.Error("unexpect")
	}
	fi.Type = TYPE_FILE
	if fi.GetSize() != "3001298" {
		t.Error("unexpect")
	}
}
