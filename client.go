package mfscli

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/golang/glog"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type Client struct {
	conn      net.Conn
	addr      string
	Password  string
	Subdir    string //remote subdir
	RootPath  string //local root path
	uid       uint32
	gid       uint32
	sessionId uint32
	sync.Mutex
	Version
}

func NewClientPwd(addr, pwd string, heartbeat bool) (c *Client) {
	c = &Client{
		Password: pwd,
		uid:      uint32(os.Getuid()),
		gid:      uint32(os.Getgid()),
		Subdir:   "/",
		RootPath: "/mnt/client",
	}
	ip := strings.Split(addr, ":")
	if len(ip) < 2 {
		//mfs client port
		c.addr = addr + ":9421"
	}
	if heartbeat {
		go c.heartbeat()
	}
	return
}

func NewClient(addr string) *Client {
	return NewClientPwd(addr, "", true)
}

func (c *Client) Connect() (err error) {
	if c.conn != nil {
		return
	}
	var conn net.Conn
	c.Lock()
	defer c.Unlock()
	for i := 0; i < 3; i++ {
		conn, err = net.DialTimeout("tcp", c.addr, time.Minute)
		if err == nil {
			c.conn = conn
			break
		}
		glog.V(8).Infof("connect mfs master error: %v retry #%d", err, i+1)
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	if err != nil {
		return
	}
	glog.V(8).Infof("connect mfs master %s successfully", c.addr)
	return
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

func (c *Client) heartbeat() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	nop := func() error {
		if err := c.Connect(); err != nil {
			return fmt.Errorf("connect error %s", err.Error())
		}
		c.Lock()
		defer c.Unlock()
		msg := PackCmd(ANTOAN_NOP, 0)
		if err := c.Send(msg); err != nil {
			return fmt.Errorf("connect error %s", err.Error())
		}
		glog.V(10).Info("sent heartbeat nop")
		return nil
	}
	for {
		select {
		case <-ticker.C:
			if err := nop(); err != nil {
				glog.Error(err)
				c.Close()
			}
		}
	}
}

func (c *Client) Send(msg []byte) error {
	if err := c.Connect(); err != nil {
		return fmt.Errorf("connect to mfs master error %s", err.Error())
	}
	c.Lock()
	defer c.Unlock()
	startSend := 0
	for startSend < len(msg) {
		sent, err := c.conn.Write(msg[startSend:])
		if err != nil {
			c.Close()
			return err
		}
		startSend += sent
	}
	return nil
}

func (c *Client) Recv(buf []byte) (n int, err error) {
	if err = c.Connect(); err != nil {
		fmt.Errorf("connect to mfs master error %s", err.Error())
		return
	}
	c.Lock()
	defer c.Unlock()
	n, err = io.ReadFull(c.conn, buf)
	if err != nil {
		c.Close()
	}
	return
}

func (c *Client) doCmd(cmd uint32, args ...interface{}) (r []byte, err error) {
	msg := PackCmd(cmd, args...)
	if err = c.Send(msg); err != nil {
		err = fmt.Errorf("send error %s", err.Error())
		return
	}
	buf := make([]byte, 8)
	var rcmd, size uint32 = ANTOAN_NOP, 4
	for rcmd == ANTOAN_NOP && size == 4 {
		_, err = c.Recv(buf)
		if err != nil {
			err = fmt.Errorf("cmd recv error %s", err.Error())
			return
		}
		read(bytes.NewBuffer(buf), &rcmd, &size)
	}
	if rcmd != cmd+1 {
		err = fmt.Errorf("mfs master cmd %d bad answer rcmd %d", cmd, rcmd)
		return
	}
	glog.V(10).Infof("command %d size %d", cmd, size)
	if size > 0 {
		buf = make([]byte, size)
		if _, err = c.Recv(buf); err != nil {
			err = fmt.Errorf("data recv error %s", err.Error())
			return
		}
		r = buf
	}
	return
}

func getStatus(buf []byte) (err error) {
	if len(buf) < 1 {
		err = fmt.Errorf("got wrong size %d<1 from mfsmaster", len(buf))
		return
	}
	var code uint8
	UnPack(buf, &code)
	if code != 0 {
		err = fmt.Errorf("got error from mfsmaster: %s", MFSStrerror(code))
		return
	}
	return
}

func (c *Client) checkBuf(buf []byte, id, minsize int) (err error) {
	if len(buf) >= 4 {
		var d uint32
		UnPack(buf[:4], &d)
		if int(d) != id {
			err = fmt.Errorf("got unexpected query id %d!=%d from mfsmaster", d, id)
			return
		}
	}
	if len(buf) < minsize {
		err = fmt.Errorf("got wrong size %d<%d from mfsmaster", len(buf), minsize)
		return
	}
	return
}

func (c *Client) CreateSession() (err error) {
	err = c.MasterVersion()
	if err != nil {
		glog.Error(err)
		return
	}
	var buf []byte
	if c.sessionId == 0 {
		pwFinal := make([]byte, 16)
		if len(c.Password) > 0 {
			buf, err = c.doCmd(CLTOMA_FUSE_REGISTER, FUSE_REGISTER_BLOB_ACL,
				REGISTER_GETRANDOM)
			if err == nil && len(buf) == 32 {
				pwMd5 := md5.Sum([]byte(c.Password))
				md := md5.New()
				md.Write(buf[:16])
				md.Write(pwMd5[:])
				md.Write(buf[16:])
				pwFinal = md.Sum(nil)
			}
		}
		buf, err = c.doCmd(CLTOMA_FUSE_REGISTER, FUSE_REGISTER_BLOB_ACL,
			REGISTER_NEWSESSION, c.Version, len(c.RootPath), c.RootPath,
			len(c.Subdir)+1, c.Subdir+"\000", pwFinal)
	} else {
		buf, err = c.doCmd(CLTOMA_FUSE_REGISTER, FUSE_REGISTER_BLOB_ACL,
			REGISTER_RECONNECT, c.sessionId, c.Version)
	}
	if err != nil {
		glog.Error(err)
		return
	}
	if len(buf) == 1 {
		err = getStatus(buf)
		if err != nil {
			glog.Error(err)
			return
		}
		if c.sessionId != 0 {
			glog.V(8).Infof("reuse session id %d", c.sessionId)
			return
		}
	}
	if len(buf) < 43 {
		err = fmt.Errorf("got wrong size %d<43 from mfsmaster", len(buf))
		glog.Error(err)
		return
	}
	var id uint32
	UnPack(buf[4:], &id)
	if 0 != c.sessionId {
		c.CloseSession()
	}
	c.sessionId = id
	glog.V(8).Infof("create new session id %d", id)
	return
}

func (c *Client) CloseSession() (err error) {
	if c.sessionId == 0 {
		return
	}
	buf, err := c.doCmd(CLTOMA_FUSE_REGISTER, FUSE_REGISTER_BLOB_ACL,
		REGISTER_CLOSESESSION, c.sessionId)
	if err != nil {
		glog.Error(err)
		return
	}
	err = getStatus(buf)
	if err != nil {
		glog.Error(err)
		return
	}
	glog.V(8).Infof("close session id %d", c.sessionId)
	c.sessionId = 0
	return
}

func (c *Client) RemoveSession(sessionId uint32) (err error) {
	buf, err := c.doCmd(CLTOMA_SESSION_COMMAND, uint8(0), sessionId)
	if err != nil {
		glog.Error(err)
		return
	}
	err = getStatus(buf)
	if err != nil {
		glog.Error(err)
		return
	}
	glog.V(8).Infof("remove session id %d", sessionId)
	return
}

func (c *Client) ListSession() (ids []uint32, err error) {
	buf, err := c.doCmd(CLTOMA_SESSION_LIST, uint8(2))
	if err != nil {
		glog.Error(err)
		return
	}
	if len(buf) <= 2 {
		return
	}
	var stats uint16
	UnPack(buf, &stats)
	if stats != 16 {
		err = fmt.Errorf("list session got wrong stats %d!=16 from mfsmaster", stats)
		glog.Error(err)
		return
	}
	if len(buf) < 188 {
		err = fmt.Errorf("list session got small size %d<188 from mfsmaster", len(buf))
		glog.Error(err)
		return
	}
	ids = make([]uint32, 0)
	var id uint32
	pos := 2
	for pos < len(buf) {
		UnPack(buf[pos:], &id)
		ids = append(ids, id)
		glog.V(8).Infof("list session id %d", id)
		pos += 21
		UnPack(buf[pos:], &id) // ileng
		pos += 4 + int(id)
		UnPack(buf[pos:], &id) // pleng
		pos += 4 + int(id) + 27 + 128
	}
	return
}

type QuotaMode int

const (
	QuotaGet QuotaMode = iota
	QuotaSet
	QuotaDel
)

func (c *Client) QuotaControl(info *QuotaInfo, mode QuotaMode) (err error) {
	if info == nil {
		return
	}
	if mode == QuotaGet {
		info.qflags = 0
	} else {
		// set or del all quota
		info.qflags = 0xff
	}
	var buf []byte
	if mode == QuotaSet {
		buf, err = c.doCmd(CLTOMA_FUSE_QUOTACONTROL, 0, info.inode, info.qflags,
			info.graceperiod, info.sinodes, info.slength, info.ssize, info.srealsize,
			info.hinodes, info.hlength, info.hsize, info.hrealsize)
	} else {
		buf, err = c.doCmd(CLTOMA_FUSE_QUOTACONTROL, 0, info.inode, info.qflags)
	}
	if err != nil {
		glog.Error(err)
		return
	}
	if len(buf) == 5 {
		err = c.checkBuf(buf, 0, 5)
		if err != nil {
			glog.Error(err)
			return
		}
		err = getStatus(buf[4:])
		glog.Error(err)
		return
	}
	err = c.checkBuf(buf, 0, 93)
	if err != nil {
		glog.Error(err)
		return
	}
	UnPack(buf[9:], &info.sinodes, &info.slength, &info.ssize, &info.srealsize,
		&info.hinodes, &info.hlength, &info.hsize, &info.hrealsize,
		&info.currinodes, &info.currlength, &info.currsize, &info.currrealsize)
	cr, q, r := info.Usage()
	glog.V(8).Infof("quota control success, %s %s %.2f%%", cr, q, r)
	return
}

type StatInfo struct {
	TotalSpace    uint64
	AvailSpace    uint64
	TrashSpace    uint64
	ReservedSpace uint64
	Inodes        uint32
}

func (c *Client) Statfs() (st *StatInfo, err error) {
	buf, err := c.doCmd(CLTOMA_FUSE_STATFS, 0)
	if err != nil {
		glog.Error(err)
		return
	}
	err = c.checkBuf(buf, 0, 40)
	if err != nil {
		glog.Error(err)
		return
	}
	st = new(StatInfo)
	UnPack(buf[4:], &st.TotalSpace, &st.AvailSpace, &st.TrashSpace,
		&st.ReservedSpace, &st.Inodes)
	return
}

func (c *Client) Access(inode uint32, mode uint16) (err error) {
	buf, err := c.doCmd(CLTOMA_FUSE_ACCESS, 0, inode, c.uid, 1, c.gid, mode)
	if err != nil {
		glog.Error(err)
		return
	}
	err = c.checkBuf(buf, 0, 5)
	if err != nil {
		glog.Error(err)
		return
	}
	err = getStatus(buf[4:])
	if err != nil {
		glog.Error(err)
		return
	}
	return
}

func (c *Client) Lookup(parent uint32, name string) (inode uint32, err error) {
	if len(name) > MFS_NAME_MAX {
		err = fmt.Errorf("name length is too long")
		glog.Error(err)
		return
	}
	buf, err := c.doCmd(CLTOMA_FUSE_LOOKUP, 0, parent, uint8(len(name)),
		name, c.uid, 1, c.gid)
	if err != nil {
		glog.Error(err)
		return
	}
	if len(buf) == 5 {
		err = c.checkBuf(buf, 0, 5)
		if err != nil {
			glog.Error(err)
			return
		}
		err = getStatus(buf[4:])
		glog.Error(err)
		return
	}
	err = c.checkBuf(buf, 0, 8)
	if err != nil {
		glog.Error(err)
		return
	}
	UnPack(buf[4:], &inode)
	return
}

func (c *Client) Mkdir(parent uint32, name string,
	mode uint16) (fi *FileInfo, err error) {
	if len(name) > MFS_NAME_MAX {
		err = fmt.Errorf("name length is too long")
		glog.Error(err)
		return
	}
	buf, err := c.doCmd(CLTOMA_FUSE_MKDIR, 0, parent, uint8(len(name)),
		name, mode, uint16(0), c.uid, 1, c.gid, uint8(0))
	if err != nil {
		glog.Error(err)
		return
	}
	if len(buf) == 5 {
		err = c.checkBuf(buf, 0, 5)
		if err != nil {
			glog.Error(err)
			return
		}
		err = getStatus(buf[4:])
		glog.Error(err)
		return
	}
	err = c.checkBuf(buf, 0, 35)
	if err != nil {
		glog.Error(err)
		return
	}
	var inode uint32
	UnPack(buf[4:], &inode)
	_, fi, err = parseFileInfo(inode, buf[8:])
	if err != nil {
		glog.Error(err)
		return
	}
	glog.V(8).Infof("mkdir name %s inode %d parent %d", name, inode, parent)
	return
}

func (c *Client) Mknod(parent uint32, name string,
	mode uint16) (fi *FileInfo, err error) {
	if len(name) > MFS_NAME_MAX {
		err = fmt.Errorf("name length is too long")
		glog.Error(err)
		return
	}
	buf, err := c.doCmd(CLTOMA_FUSE_MKNOD, 0, parent, uint8(len(name)),
		name, uint8(1), mode, uint16(0), c.uid, 1, c.gid, 0)
	if err != nil {
		glog.Error(err)
		return
	}
	if len(buf) == 5 {
		err = c.checkBuf(buf, 0, 5)
		if err != nil {
			glog.Error(err)
			return
		}
		err = getStatus(buf[4:])
		glog.Error(err)
		return
	}
	err = c.checkBuf(buf, 0, 35)
	if err != nil {
		glog.Error(err)
		return
	}
	var inode uint32
	UnPack(buf[4:], &inode)
	_, fi, err = parseFileInfo(inode, buf[8:])
	if err != nil {
		glog.Error(err)
		return
	}
	glog.V(8).Infof("mknod name %s inode %d parent %d", name, inode, parent)
	return
}

func (c *Client) remove(parent uint32, name string, cmd uint32) (err error) {
	if len(name) > MFS_NAME_MAX {
		err = fmt.Errorf("name length is too long")
		glog.Error(err)
		return
	}
	buf, err := c.doCmd(cmd, 0, parent, uint8(len(name)),
		name, c.uid, 1, c.gid)
	if err != nil {
		glog.Error(err)
		return
	}
	err = c.checkBuf(buf, 0, 5)
	if err != nil {
		glog.Error(err)
		return
	}
	err = getStatus(buf[4:])
	if err != nil {
		glog.Error(err)
		return
	}
	glog.V(8).Infof("remove name %s parent %d", name, parent)
	return
}

func (c *Client) Rmdir(parent uint32, name string) (err error) {
	return c.remove(parent, name, CLTOMA_FUSE_RMDIR)
}

func (c *Client) Unlink(parent uint32, name string) (err error) {
	return c.remove(parent, name, CLTOMA_FUSE_UNLINK)
}

type ReaddirInfo struct {
	Type  uint8
	Inode uint32
	Name  string
}

type ReaddirInfoMap map[uint32]*ReaddirInfo

func (c *Client) Readdir(parent uint32) (infoMap ReaddirInfoMap, err error) {
	//max entries 0xffff
	buf, err := c.doCmd(CLTOMA_FUSE_READDIR, 0, parent, c.uid, 1, c.gid,
		uint8(0), 0xffff, uint64(0))
	if err != nil {
		glog.Error(err)
		return
	}
	if len(buf) == 5 {
		err = c.checkBuf(buf, 0, 5)
		if err != nil {
			glog.Error(err)
			return
		}
		err = getStatus(buf[4:])
		glog.Error(err)
		return
	}
	err = c.checkBuf(buf, 0, 27)
	if err != nil {
		glog.Error(err)
		return
	}
	pos := 12
	var sz uint8
	infoMap = make(ReaddirInfoMap)
	for pos < len(buf) {
		info := new(ReaddirInfo)
		UnPack(buf[pos:], &sz)
		pos++
		info.Name = string(buf[pos : pos+int(sz)])
		pos += int(sz)
		UnPack(buf[pos:], &info.Inode, &info.Type)
		pos += 5
		infoMap[info.Inode] = info
		glog.V(10).Infof("readdir inode %d name %s", info.Inode, info.Name)
	}
	glog.V(8).Infof("readdir parent %d len %d", parent, len(infoMap))
	return
}

const (
	TYPE_FILE = iota + 1
	TYPE_DIRECTORY
	TYPE_SYMLINK
	TYPE_FIFO
	TYPE_BLOCKDEV
	TYPE_CHARDEV
	TYPE_SOCKET
	TYPE_TRASH
	TYPE_SUSTAINED
)

// flags: 01 noacache 02 noecache 04 allowdatacache
// 		  08 noxattr 10 directmode
// 'floating-point' size
// examples:
//    1200 =  12.00 B
// 1023443 = 234.43 kB
// 2052312 = 523.12 MB
// 3001298 =  12.98 GB
// 4001401 =  14.01 TB
type FileInfo struct {
	Flags uint8
	Type  uint8
	Inode uint32
	Uid   uint32
	Gid   uint32
	Mode  os.FileMode
	NLink uint32
	ATime time.Time
	MTime time.Time
	CTime time.Time
	Size  uint64
}

func (fi *FileInfo) String() string {
	return fmt.Sprintf("inode %d type %d flags 0x%x mode %s uid %d gid %d size %d\n\tatime %v mtime %v ctime %v",
		fi.Inode, fi.Type, fi.Flags, fi.Mode, fi.Uid,
		fi.Gid, fi.Size, fi.ATime, fi.MTime, fi.CTime)
}

func parseFileInfo(inode uint32, buf []byte) (size uint32,
	fi *FileInfo, err error) {
	if len(buf) < 27 {
		err = fmt.Errorf("file info buf length is too short")
		glog.Error(err)
		return
	}
	fi = new(FileInfo)
	fi.Inode = inode
	var mode uint16
	var atime, mtime, ctime, dev uint32
	UnPack(buf, &fi.Flags, &mode, &fi.Uid, &fi.Gid, &atime,
		&mtime, &ctime, &fi.NLink)
	size += 27
	fi.Type = uint8(mode >> 12)
	fi.Mode = os.FileMode(mode & 0x0FFF)
	fi.ATime = time.Unix(int64(atime), 0)
	fi.MTime = time.Unix(int64(mtime), 0)
	fi.CTime = time.Unix(int64(ctime), 0)
	fi.Size = 0
	defer func() {
		glog.V(10).Infof("parseFileInfo %s", fi)
	}()
	switch fi.Type {
	case TYPE_FILE:
		goto readSize
	case TYPE_DIRECTORY:
		fi.Mode |= os.ModeDir
		goto readSize
	case TYPE_SYMLINK:
		fi.Mode |= os.ModeSymlink
		goto readSize
	case TYPE_FIFO:
		fi.Mode |= os.ModeNamedPipe
		return
	case TYPE_SOCKET:
		fi.Mode |= os.ModeSocket
		return
	case TYPE_BLOCKDEV:
		fi.Mode |= os.ModeDevice
		goto readDev
	case TYPE_CHARDEV:
		fi.Mode |= os.ModeCharDevice
		goto readDev
	default:
		return
	}
	return
readSize:
	UnPack(buf[size:], &fi.Size)
	size += 8
	return
readDev:
	UnPack(buf[size:], &dev)
	fi.Size = uint64(dev)
	size += 4
	return
}

// flags 01 read 02 write 04
func (c *Client) Open(inode uint32, flags uint8) (err error) {
	buf, err := c.doCmd(CLTOMA_FUSE_OPEN, 0, inode, c.uid, 1, c.gid, flags)
	if err != nil {
		glog.Error(err)
		return
	}
	if len(buf) == 5 {
		err = c.checkBuf(buf, 0, 5)
		if err != nil {
			glog.Error(err)
			return
		}
		err = getStatus(buf[4:])
		glog.Error(err)
		return
	}
	err = c.checkBuf(buf, 0, 31)
	if err != nil {
		glog.Error(err)
		return
	}
	_, _, err = parseFileInfo(inode, buf[4:])
	if err != nil {
		glog.Error(err)
		return
	}
	return
}

// mknod and open
func (c *Client) Create(parent uint32, name string,
	mode uint16) (fi *FileInfo, err error) {
	if len(name) > MFS_NAME_MAX {
		err = fmt.Errorf("name length is too long")
		glog.Error(err)
		return
	}
	buf, err := c.doCmd(CLTOMA_FUSE_CREATE, 0, parent, uint8(len(name)),
		name, mode, uint16(0), c.uid, 1, c.gid)
	if err != nil {
		glog.Error(err)
		return
	}
	if len(buf) == 5 {
		err = c.checkBuf(buf, 0, 5)
		if err != nil {
			glog.Error(err)
			return
		}
		err = getStatus(buf[4:])
		glog.Error(err)
		return
	}
	err = c.checkBuf(buf, 0, 35)
	if err != nil {
		glog.Error(err)
		return
	}
	var inode uint32
	UnPack(buf[4:], &inode)
	_, fi, err = parseFileInfo(inode, buf[8:])
	if err != nil {
		glog.Error(err)
		return
	}
	glog.V(8).Infof("create name %s inode %d mode %o parent %d",
		name, inode, mode, parent)
	return
}

func (c *Client) GetAttr(inode uint32) (fi *FileInfo, err error) {
	buf, err := c.doCmd(CLTOMA_FUSE_GETATTR, 0, inode)
	if err != nil {
		glog.Error(err)
		return
	}
	if len(buf) == 5 {
		err = c.checkBuf(buf, 0, 5)
		if err != nil {
			glog.Error(err)
			return
		}
		err = getStatus(buf[4:])
		glog.Error(err)
		return
	}
	err = c.checkBuf(buf, 0, 31)
	if err != nil {
		glog.Error(err)
		return
	}
	_, fi, err = parseFileInfo(inode, buf[4:])
	if err != nil {
		glog.Error(err)
		return
	}
	return
}

// for setmask
const (
	SET_WINATTR_FLAG = 1 << iota
	SET_MODE_FLAG
	SET_UID_FLAG
	SET_GID_FLAG
	SET_MTIME_NOW_FLAG
	SET_MTIME_FLAG
	SET_ATIME_FLAG
	SET_ATIME_NOW_FLAG
)

func (c *Client) SetAttr(inode uint32, setmask uint8, mode uint16,
	uid, gid, atime, mtime uint32) (fi *FileInfo, err error) {
	buf, err := c.doCmd(CLTOMA_FUSE_SETATTR, 0, inode, uint8(0), c.uid, 1, c.gid,
		setmask, mode, uid, gid, atime, mtime, uint8(0))
	if err != nil {
		glog.Error(err)
		return
	}
	if len(buf) == 5 {
		err = c.checkBuf(buf, 0, 5)
		if err != nil {
			glog.Error(err)
			return
		}
		err = getStatus(buf[4:])
		glog.Error(err)
		return
	}
	err = c.checkBuf(buf, 0, 31)
	if err != nil {
		glog.Error(err)
		return
	}
	_, fi, err = parseFileInfo(inode, buf[4:])
	if err != nil {
		glog.Error(err)
		return
	}
	return
}

func (c *Client) Chmod(inode uint32, mode uint16) (fi *FileInfo, err error) {
	return c.SetAttr(inode, SET_MODE_FLAG, mode, 0, 0, 0, 0)
}

func (c *Client) Chown(inode uint32, uid, gid uint32) (fi *FileInfo, err error) {
	return c.SetAttr(inode, SET_UID_FLAG|SET_GID_FLAG, 0, uid, gid, 0, 0)
}

func (c *Client) Undel(inode uint32) (err error) {
	buf, err := c.doCmd(CLTOMA_FUSE_UNDEL, 0, inode)
	if err != nil {
		glog.Error(err)
		return
	}
	err = c.checkBuf(buf, 0, 5)
	if err != nil {
		glog.Error(err)
		return
	}
	err = getStatus(buf[4:])
	if err != nil {
		glog.Error(err)
	}
	return
}

func (c *Client) Purge(inode uint32) (err error) {
	buf, err := c.doCmd(CLTOMA_FUSE_PURGE, 0, inode)
	if err != nil {
		glog.Error(err)
		return
	}
	err = c.checkBuf(buf, 0, 5)
	if err != nil {
		glog.Error(err)
		return
	}
	err = getStatus(buf[4:])
	if err != nil {
		glog.Error(err)
	}
	return
}
