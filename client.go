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
	password  string
	Subdir    string //remote subdir
	rootPath  string //local root path
	uid       uint32
	gid       uint32
	sessionId uint32
	sync.Mutex
	Version
}

func NewClientPwd(addr, pwd string, heartbeat bool) (c *Client) {
	c = &Client{
		password: pwd,
		uid:      uint32(os.Getuid()),
		gid:      uint32(os.Getgid()),
		Subdir:   "/",
		rootPath: "/mnt/client",
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
		if len(c.password) > 0 {
			buf, err = c.doCmd(CLTOMA_FUSE_REGISTER, FUSE_REGISTER_BLOB_ACL,
				REGISTER_GETRANDOM)
			if err == nil && len(buf) == 32 {
				pwMd5 := md5.Sum([]byte(c.password))
				md := md5.New()
				md.Write(buf[:16])
				md.Write(pwMd5[:])
				md.Write(buf[16:])
				pwFinal = md.Sum(nil)
			}
		}
		buf, err = c.doCmd(CLTOMA_FUSE_REGISTER, FUSE_REGISTER_BLOB_ACL,
			REGISTER_NEWSESSION, c.Version, len(c.rootPath), c.rootPath,
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

type quotaMode int

const (
	quotaGet quotaMode = iota
	quotaSet
	quotaDel
)

func (c *Client) QuotaControl(info *QuotaInfo, mode quotaMode) (err error) {
	if info == nil {
		return
	}
	if mode == quotaGet {
		info.qflags = 0
	} else {
		// set or del all quota
		info.qflags = 0xff
	}
	var buf []byte
	if mode == quotaSet {
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
	mode uint16) (inode uint32, err error) {
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
	err = c.checkBuf(buf, 0, 8)
	if err != nil {
		glog.Error(err)
		return
	}
	UnPack(buf[4:], &inode)
	glog.V(8).Infof("mkdir name %s inode %d parent %d", name, inode, parent)
	return
}

func (c *Client) Mknod(parent uint32, name string, mode uint16) (inode uint32, err error) {
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
	err = c.checkBuf(buf, 0, 8)
	if err != nil {
		glog.Error(err)
		return
	}
	UnPack(buf[4:], &inode)
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
	class uint8
	inode uint32
	name  string
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
		info.name = string(buf[pos : pos+int(sz)])
		pos += int(sz)
		UnPack(buf[pos:], &info.inode, &info.class)
		pos += 5
		infoMap[info.inode] = info
		glog.V(10).Infof("readdir inode %d name %s", info.inode, info.name)
	}
	glog.V(8).Infof("readdir parent %d len %d", parent, len(infoMap))
	return
}
