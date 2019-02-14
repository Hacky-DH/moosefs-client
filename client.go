package mfscli

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"path/filepath"
	"strings"
)

var (
	masterAddr   string
	masterPsw    string
	masterSubdir string
)

func init() {
	flag.StringVar(&masterAddr, "H", "", "mfs master host")
	flag.StringVar(&masterPsw, "P", "", "mfs master password")
	flag.StringVar(&masterSubdir, "p", "", "mfs remote sub path in mfs tree")
}

type Client struct {
	mc        *MAClient
	Cwd       string
	currInode uint32
}

type File struct {
	Path   string
	info   *FileInfo
	client *Client
}

func NewClientFull(addr, password, subDir string) (c *Client, err error) {
	c = &Client{
		mc:        NewMAClientPwd(addr, password, true),
		Cwd:       "/",
		currInode: MFS_ROOT_ID,
	}
	if len(subDir) > 0 {
		if !filepath.IsAbs(subDir) {
			subDir = filepath.Join(string(filepath.Separator), subDir)
		}
		c.mc.Subdir = subDir
	}
	err = c.mc.CreateSession()
	if err != nil {
		c.mc.Close()
		return
	}
	return
}

func NewCLient() (c *Client, err error) {
	return NewClientFull(masterAddr, masterPsw, masterSubdir)
}

// close client, not file
func (c *Client) Close() {
	if c.mc != nil {
		c.mc.CloseSession()
		c.mc.Close()
		c.mc = nil
	}
}

func (c *Client) check(path string) (p string, err error) {
	if c.mc == nil {
		err = fmt.Errorf("client is closed")
		return
	}
	p = filepath.Clean(path)
	if len(p) >= MFS_PATH_MAX {
		err = fmt.Errorf("path is too long")
		return
	}
	return
}

func (c *Client) lookup(path string) (parent uint32, info *FileInfo, err error) {
	p, err := c.check(path)
	if err != nil {
		return
	}
	curr := c.currInode
	if filepath.IsAbs(p) {
		curr = MFS_ROOT_ID
	}
	pa := strings.Split(p, string(filepath.Separator))
	for _, part := range pa {
		if len(part) == 0 {
			continue
		}
		info, err = c.mc.Lookup(curr, part)
		if err != nil {
			return
		}
		parent = curr
		curr = info.Inode
	}
	if info == nil && curr == MFS_ROOT_ID {
		info, err = c.mc.GetAttr(curr)
		if err != nil {
			return
		}
		parent = curr
	}
	glog.V(5).Infof("client lookup path %s result: parent %d path %s inode %d",
		path, parent, p, info.Inode)
	return
}

func (c *Client) Open(path string) (f *File, err error) {
	_, info, err := c.lookup(path)
	if err != nil {
		return
	}
	info, err = c.mc.Open(info.Inode, 1)
	if err != nil {
		return
	}
	f = &File{
		Path:   path,
		info:   info,
		client: c,
	}
	return
}

func (c *Client) Create(path string) (f *File, err error) {
	_, info, err := c.lookup(filepath.Dir(path))
	if err != nil {
		return
	}
	fi, err := c.mc.Create(info.Inode, filepath.Base(path), 0666)
	if err != nil {
		return
	}
	f = &File{
		Path:   path,
		info:   fi,
		client: c,
	}
	return
}

func (c *Client) OpenOrCreate(path string) (f *File, err error) {
	_, _, err = c.lookup(path)
	if err != nil {
		return c.Create(path)
	}
	return c.Open(path)
}

func (c *Client) Unlink(path string) (err error) {
	p, _, err := c.lookup(path)
	if err != nil {
		return
	}
	return c.mc.Unlink(p, filepath.Base(path))
}

func (c *Client) Chdir(path string) (err error) {
	_, info, err := c.lookup(path)
	if err != nil {
		return
	}
	if info.IsDir() {
		if filepath.IsAbs(path) {
			c.Cwd = path
		} else {
			c.Cwd = filepath.Join(c.Cwd, path)
		}
		c.currInode = info.Inode
	} else {
		err = fmt.Errorf("chdir path %s is not a directory", path)
	}
	return
}

const (
	MFSBLOCKSINCHUNK  = 0x400
	MFSCHUNKSIZE      = 0x04000000
	MFSCHUNKMASK      = 0x03FFFFFF
	MFSCHUNKBITS      = 26
	MFSCHUNKBLOCKMASK = 0x03FF0000
	MFSBLOCKSIZE      = 0x10000
	MFSBLOCKMASK      = 0x0FFFF
	MFSBLOCKNEGMASK   = 0x7FFF0000
	MFSBLOCKBITS      = 16
	MFSCRCEMPTY       = 0xD7978EEB
	MFSHDRSIZE        = 0x2000
)

func (f *File) Length() string {
	return f.info.GetSize()
}

func (f *File) Write(buf []byte, offset uint64) (n uint32, err error) {
	size := uint32(len(buf))
	for size > 0 {
		chindx := uint32(offset >> MFSCHUNKBITS)
		cs, e := f.client.mc.WriteChunk(f.info.Inode, chindx, CHUNKOPFLAG_CANMODTIME)
		if e != nil {
			err = fmt.Errorf("write chunk failed: %v", e)
			return
		}
		off := uint32(offset & MFSCHUNKMASK)
		sz := MFSCHUNKMASK - off
		if sz > size {
			sz = size
		}
		var rs uint32
		rs, err = cs.Write(buf[n:n+sz], off)
		if err != nil || rs != sz {
			err = fmt.Errorf("write data to chunkserver failed: %v", err)
			return
		}
		length := uint64(off + sz)
		err = f.client.mc.WriteChunkEnd(cs.ChunkId, f.info.Inode,
			chindx, length, CHUNKOPFLAG_CANMODTIME)
		if err != nil {
			err = fmt.Errorf("write end chunk failed: %v", err)
			return
		}
		size -= sz
		n += sz
		offset += uint64(sz)
	}
	return
}
