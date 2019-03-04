package mfscli

/*
MIT License

Copyright (c) 2019 DHacky
*/

import (
	"fmt"
	"github.com/golang/glog"
)

// tool is querying information from mfs
func NewTools(addr string) *MAClient {
	return NewMAClientPwd(addr, "", false)
}

func (c *MAClient) MasterVersion() error {
	buf, err := c.doCmd(ANTOAN_GET_VERSION)
	if err != nil {
		return err
	}
	var ver uint32
	UnPack(buf, &ver)
	c.Version = GetVersion(ver)
	if c.Version.LessThan(3, 0, 72) {
		return fmt.Errorf("client only support mfsmaster version >= 3.0.72")
	}
	glog.V(5).Infof("mfsmaster version %s", c.Version)
	return nil
}

type QuotaInfo struct {
	size                               int
	inode                              uint32
	path                               string
	graceperiod                        uint32
	exceeded, qflags                   uint8
	stimestamp                         uint32
	sinodes                            uint32 // soft
	slength, ssize, srealsize          uint64
	hinodes                            uint32 // hard
	hlength, hsize, hrealsize          uint64
	currinodes                         uint32 // current
	currlength, currsize, currrealsize uint64
}

// get the max retio, only care soft quota
func (info *QuotaInfo) Usage() (current, quota string, retio float64) {
	c := float64(info.currlength)
	q := float64(info.slength)
	current = FormatBytes(c, Binary)
	quota = FormatBytes(q, Binary)
	if info.slength != 0 {
		r := c / q
		if r >= retio {
			retio = r
		}
	}
	if info.ssize != 0 {
		c := float64(info.currsize)
		q := float64(info.ssize)
		r := c / q
		if r > retio {
			retio = r
			current = FormatBytes(c, Binary)
			quota = FormatBytes(q, Binary)
		}
	}
	if info.sinodes != 0 {
		c := float64(info.currinodes)
		q := float64(info.sinodes)
		r := c / q
		if r > retio {
			retio = r
			current = FormatBytes(c, Decimal)
			quota = FormatBytes(q, Decimal)
		}
	}
	return
}

func (c *MAClient) UnPackQuota(buf []byte) *QuotaInfo {
	if len(buf) <= 102 {
		return nil
	}
	q := new(QuotaInfo)
	var leng uint32
	UnPack(buf[q.size:q.size+8], &q.inode, &leng)
	q.size += 8
	q.path = string(buf[q.size : q.size+int(leng)])
	q.size += int(leng)
	UnPack(buf[q.size:q.size+94], &q.graceperiod, &q.exceeded, &q.qflags,
		&q.stimestamp, &q.sinodes, &q.slength, &q.ssize, &q.srealsize,
		&q.hinodes, &q.hlength, &q.hsize, &q.hrealsize,
		&q.currinodes, &q.currlength, &q.currsize, &q.currrealsize)
	q.size += 94
	return q
}

type QuotaInfoMap map[string]*QuotaInfo

func (c *MAClient) AllQuotaInfo() (quota QuotaInfoMap, err error) {
	err = c.MasterVersion()
	if err != nil {
		return
	}
	buf, err := c.doCmd(CLTOMA_QUOTA_INFO)
	if err != nil {
		return
	}
	var pos int
	quota = make(QuotaInfoMap)
	for pos < len(buf) {
		q := c.UnPackQuota(buf[pos:])
		if q == nil {
			break
		}
		quota[q.path] = q
		pos += q.size
		glog.V(10).Infof("quota inode %d path %s", q.inode, q.path)
	}
	glog.V(5).Infof("quota number %d", len(quota))
	return
}

// get all usage of mfs by sending command to mfs master
func GetUsage(masterAddr string) (QuotaInfoMap, error) {
	m := NewTools(masterAddr)
	defer m.Close()
	return m.AllQuotaInfo()
}
