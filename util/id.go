package util

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
)

var (
	random1 = rand.Reader
)

type uuid [16]byte

// 生成短id(22个字符)
func UuidV4_22() string {
	uuid := uuidV4()
	return base64.RawURLEncoding.EncodeToString(uuid[:])
}

func UuidV1_22() string {
	uuid := uuidV1()
	return base64.RawURLEncoding.EncodeToString(uuid[:])
}

func NewSnowflaker(generatorId int64) (*snowflaker, error) {
	if generatorId < 0 || generatorId > gidLimit {
		return nil, errors.New("生成器序号过大")
	}
	// 生成一个新节点
	return &snowflaker{
		time:   0,
		gid:    generatorId,
		serial: 0,
	}, nil
}

// 生成不带 '-' 的uuid
func UuidV4_32() string {
	uuid := uuidV4()
	target := [32]byte{}

	encodeToString(uuid[:], target[:], false)

	return string(target[:])
}

// 生成不带 '-' 的uuid
func UuidV1_32() string {
	uuid := uuidV1()
	target := [32]byte{}

	encodeToString(uuid[:], target[:], false)

	return string(target[:])
}

// 生成UUID (以"-"分隔的 UUID.V1)
func UuidV1() string {
	uuid := uuidV1()
	target := [36]byte{}

	encodeToString(uuid[:], target[:], true)

	return string(target[:])
}

// 生成UUID (以"-"分隔的 UUID.V4)
func UuidV4() string {
	uuid := uuidV4()
	target := [36]byte{}

	encodeToString(uuid[:], target[:], true)

	return string(target[:])
}

// 生成uuidV4需要的16位byte随机数
func uuidV4() (r16 uuid) {
	io.ReadFull(random1, r16[:])

	r16[6] = (r16[6] & 0x0f) | 0x40 // Version 4
	r16[8] = (r16[8] & 0x3f) | 0x80 // Variant is 10

	return
}

// 生成uuidV1需要的16位byte数
func uuidV1() (mu uuid) {
	now, seq, err := getTime()
	if err != nil {
		return
	}

	timeLow := uint32(now & 0xffffffff)
	timeMid := uint16((now >> 32) & 0xffff)
	timeHi := uint16((now >> 48) & 0x0fff)
	timeHi |= 0x1000 // Version 1

	binary.BigEndian.PutUint32(mu[0:], timeLow)
	binary.BigEndian.PutUint16(mu[4:], timeMid)
	binary.BigEndian.PutUint16(mu[6:], timeHi)
	binary.BigEndian.PutUint16(mu[8:], seq)

	setNodeInterface(mu[10:])

	return
}

func encodeToString(uuid []byte, buf []byte, hyphen bool) {
	if hyphen {
		hex.Encode(buf[:], uuid[:4])
		hex.Encode(buf[9:13], uuid[4:6])
		hex.Encode(buf[14:18], uuid[6:8])
		hex.Encode(buf[19:23], uuid[8:10])
		hex.Encode(buf[24:], uuid[10:])

		buf[8] = '-'
		buf[13] = '-'
		buf[18] = '-'
		buf[23] = '-'
	} else {
		hex.Encode(buf[:], uuid[:])
	}
}

// 设置 uuidV1 需要的网卡名称
func SetV1InterfaceName(name string) {
	infName = name
}
