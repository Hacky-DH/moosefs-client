package mfscli

/*
MIT License

Copyright (c) 2019 DHacky
*/

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

const MFS_VERSION string = "3.0.103"

type Version uint32

func ParseVersionString(str string) Version {
	arr := strings.Split(str, ".")
	var maj uint16
	var mid, min uint8
	switch len(arr) {
	case 3:
		m, _ := strconv.ParseUint(arr[2], 10, 8)
		min = uint8(m)
		fallthrough
	case 2:
		m, _ := strconv.ParseUint(arr[1], 10, 8)
		mid = uint8(m)
		fallthrough
	case 1:
		m, _ := strconv.ParseUint(arr[0], 10, 16)
		maj = uint16(m)
	}
	return ParseVersionInt(maj, mid, min)
}

func ParseVersionInt(maj uint16, mid, min uint8) Version {
	v := Pack(maj, mid, min)
	var res uint32
	read(bytes.NewBuffer(v), &res)
	return Version(res)
}

func (v Version) ToInt() (maj uint16, mid, min uint8) {
	res := Pack(uint32(v))
	UnPack(res, &maj, &mid, &min)
	return
}

func (v Version) String() string {
	maj, mid, min := v.ToInt()
	return fmt.Sprintf("%d.%d.%d", maj, mid, min)
}

func (v Version) MoreThan(maj uint16, mid, min uint8) bool {
	return v >= ParseVersionInt(maj, mid, min)
}

func (v Version) LessThan(maj uint16, mid, min uint8) bool {
	return v < ParseVersionInt(maj, mid, min)
}

func GetVersion(ver uint32) Version {
	maj, mid, min := Version(ver).ToInt()
	if maj >= 2 {
		min >>= 1
	}
	return ParseVersionInt(maj, mid, min)
}
