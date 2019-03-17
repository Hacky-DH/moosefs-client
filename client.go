package mfscli

/*
MIT License

Copyright (c) 2019 DHacky
*/

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"os"
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
	// the default Subdir is /
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

func NewClient() (c *Client, err error) {
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

// check master connection and the length of path
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

// path based lookup
// if success, parent is parent inode of path
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
	glog.V(8).Infof("client lookup path %s result: parent %d path %s inode %d",
		path, parent, p, info.Inode)
	return
}

func (c *Client) Open(path string, flags uint8) (f *File, err error) {
	_, info, err := c.lookup(path)
	if err != nil {
		return
	}
	info, err = c.mc.Open(info.Inode, flags)
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
	return c.Open(path, WANT_READ|WANT_WRITE)
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

// write one chunk by one
func (f *File) Write(buf []byte, offset uint64) (n uint32, err error) {
	size := uint32(len(buf))
	for size > 0 {
		chindx := uint32(offset >> MFSCHUNKBITS)
		cs, e := f.client.mc.WriteChunk(f.info.Inode, chindx, 0)
		if e != nil {
			err = fmt.Errorf("write chunk failed: %v", e)
			return
		}
		off := uint32(offset & MFSCHUNKMASK)
		sz := MFSCHUNKSIZE - off
		if sz > size {
			sz = size
		}
		glog.V(10).Infof("client write chunk cindex %d buf[%d:%d] off %d",
			chindx, n, sz, off)
		var rs uint32
		rs, err = cs.Write(buf[n:n+sz], offset)
		if err != nil || rs != sz {
			err = fmt.Errorf("write data to chunkserver failed: %v", err)
			return
		}
		length := uint64(off + sz)
		err = f.client.mc.WriteChunkEnd(cs.ChunkId, f.info.Inode,
			chindx, length, 0)
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

// read one chunk by one
func (f *File) Read(buf []byte, offset uint64) (n uint32, err error) {
	if offset >= f.info.Size {
		err = fmt.Errorf("read offset is longer than file length")
		return
	}
	size := uint32(len(buf))
	glog.V(10).Infof("client read file size %d offset %d", f.info.Size, offset)
	for size > 0 {
		chindx := uint32(offset >> MFSCHUNKBITS)
		cs, e := f.client.mc.ReadChunk(f.info.Inode, chindx, 0)
		if e != nil {
			err = fmt.Errorf("read chunk failed: %v", e)
			return
		}
		off := uint32(offset & MFSCHUNKMASK)
		sz := MFSCHUNKSIZE - off
		if sz > size {
			sz = size
		}
		glog.V(10).Infof("client read chunk cindex %d buf[%d:%d] off %d",
			chindx, n, sz, off)
		var rs uint32
		rs, err = cs.Read(buf[n:n+sz], uint64(off))
		if err != nil || rs != sz {
			err = fmt.Errorf("read data from chunkserver failed: %v", err)
			return
		}
		size -= sz
		n += sz
		offset += uint64(sz)
	}
	return
}

func (c *Client) Mkdir(path string) (err error) {
	_, info, err := c.lookup(path)
	if err == nil {
		// already exists
		return
	}
	_, info, err = c.lookup(filepath.Dir(path))
	if err != nil {
		// parent dir is not exists
		return
	}
	_, err = c.mc.Mkdir(info.Inode, filepath.Base(path), 0755)
	if err != nil {
		return
	}
	return
}

func (c *Client) Rmdir(path string) (err error) {
	p, _, err := c.lookup(path)
	if err != nil {
		return
	}
	return c.mc.Rmdir(p, filepath.Base(path))
}

func (c *Client) Readdir(path string) (infoMap ReaddirInfoMap, err error) {
	_, info, err := c.lookup(path)
	if err != nil {
		return
	}
	return c.mc.Readdir(info.Inode)
}

func (c *Client) GetDirStats(path string) (ds *DirStats, err error) {
	_, info, err := c.lookup(path)
	if err != nil {
		return
	}
	return c.mc.GetDirStats(info.Inode)
}

func (c *Client) Chmod(path string, mode uint16) (fi *FileInfo, err error) {
	_, info, err := c.lookup(path)
	if err != nil {
		return
	}
	return c.mc.Chmod(info.Inode, mode)
}

func (c *Client) Chown(path string, uid, gid uint32) (fi *FileInfo, err error) {
	_, info, err := c.lookup(path)
	if err != nil {
		return
	}
	return c.mc.Chown(info.Inode, uid, gid)
}

// write local file to mfs
func (c *Client) WriteFile(localPath, path string) (err error) {
	f, err := os.Open(localPath)
	if err != nil {
		return
	}
	defer f.Close()
	info, err := os.Stat(localPath)
	if err != nil {
		return
	}
	size := uint64(info.Size())
	file, err := c.OpenOrCreate(path)
	if err != nil {
		return
	}
	buf := make([]byte, MFSCHUNKSIZE)
	var n int
	var wn uint32
	var off uint64
	for {
		n, err = f.Read(buf)
		if err != nil {
			return
		}
		if n <= 0 {
			break
		}
		wn, err = file.Write(buf[:n], off)
		if err != nil {
			return
		}
		off += uint64(wn)
		glog.Infof("write file percent %.2f%%", float64(off*100)/float64(size))
		if off >= size {
			break
		}
	}
	glog.V(5).Infof("write file %s to mfs %s size %d", localPath, path, off)
	return
}

// read mfs file to local file
func (c *Client) ReadFile(path, localPath string) (err error) {
	file, err := c.Open(path, WANT_READ)
	if err != nil {
		return
	}
	dst, err := os.Create(localPath)
	if err != nil {
		return
	}
	buf := make([]byte, MFSCHUNKSIZE)
	var n uint32
	var wn int
	var off uint64
	for {
		sz := MFSCHUNKSIZE - (off & MFSCHUNKMASK)
		if sz > (file.info.Size - off) {
			sz = file.info.Size - off
		}
		n, err = file.Read(buf[:sz], off)
		if err != nil {
			return
		}
		if n <= 0 {
			break
		}
		wn, err = dst.Write(buf[:n])
		if err != nil {
			return
		}
		off += uint64(wn)
		glog.Infof("read file percent %.2f%%",
			float64(off*100)/float64(file.info.Size))
		if off >= file.info.Size {
			break
		}
	}
	err = dst.Close()
	if err != nil {
		return
	}
	glog.V(5).Infof("read mfs %s to local file %s size %d", path, localPath, off)
	return
}
