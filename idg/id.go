package idg

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"sync"
	"time"
)

var (
	random1 = rand.Reader

	v7Mu     sync.Mutex // uuidV7 时间与计数器锁
	v7LastMs int64      // uuidV7 上次使用的时间戳(毫秒)
	v7Serial uint16     // uuidV7 同毫秒内计数器
)

type uuid [16]byte

func (u uuid) Str36() string {
	target := [36]byte{}

	encodeToString(u[:], target[:], true)

	return string(target[:])
}

func (u uuid) Str32() string {
	target := [32]byte{}

	encodeToString(u[:], target[:], false)

	return string(target[:])
}

func (u uuid) Str22() string {
	return base64.RawURLEncoding.EncodeToString(u[:])
}

func NewSnowflaker(generatorId int64) (*snowflaker, error) {
	if generatorId < 0 || generatorId > snowGidLimit {
		return nil, errors.New("生成器序号过大")
	}
	return &snowflaker{
		time:   0,
		gid:    generatorId,
		serial: 0,
	}, nil
}

func NewIotIdSerial(devId, devArea int64) (*iotIdSerial, error) {
	if devId < 0 || devId > iotDevLimit {
		return nil, errors.New("设备序号过大")
	}
	if devId < 0 || devArea > iotDevAreaLimit {
		return nil, errors.New("生成器区域号过大")
	}
	return &iotIdSerial{
		time:   0,
		dev:    int64(devArea<<int64(iotDevLen) | devId),
		serial: 0,
	}, nil
}

// 生成uuidV4需要的16位byte随机数
func UuidV4() (r16 uuid) {
	io.ReadFull(random1, r16[:])

	r16[6] = (r16[6] & 0x0f) | 0x40 // Version 4
	r16[8] = (r16[8] & 0x3f) | 0x80 // Variant is 10

	return
}

// 生成uuidV1需要的16位byte数
func UuidV1() (mu uuid) {
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

// 生成uuidV7需要的16位byte数 (RFC 9562: 48bit 毫秒时间戳 + 12bit 计数 + 62bit 随机)
func UuidV7() (r uuid) {
	io.ReadFull(random1, r[:])

	now := time.Now().UnixMilli()

	v7Mu.Lock()
	if now <= v7LastMs {
		now = v7LastMs
		v7Serial++
		if v7Serial > 0x0fff {
			// 当前毫秒计数用尽，等待时间轮转
			for now <= v7LastMs {
				now = time.Now().UnixMilli()
			}
			v7LastMs = now
			v7Serial = 0
		}
	} else {
		v7LastMs = now
		v7Serial = 0
	}
	serial := v7Serial
	v7Mu.Unlock()

	r[0] = byte(now >> 40)
	r[1] = byte(now >> 32)
	r[2] = byte(now >> 24)
	r[3] = byte(now >> 16)
	r[4] = byte(now >> 8)
	r[5] = byte(now)

	r[6] = 0x70 | byte(serial>>8) // Version 7
	r[7] = byte(serial)

	r[8] = (r[8] & 0x3f) | 0x80 // Variant is 10

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
