package idg

import (
	"encoding/base64"
	"encoding/binary"
	"sync"
	"time"
)

const (
	iotDevLen       uint8 = 16
	iotDevLimit     int64 = 0xFFFF // 设备序号最大值 65535  16 bits
	iotDevAreaLimit int64 = 0xFFFF // 区域(站点序号最大值) 65535 16 bits

)

type iotIdSerial struct {
	l      sync.Mutex // 工作锁
	time   int64      // 时间戳
	serial int64      // 自增序列 (0~1023)
	dev    int64      // 设备序号 (28位数，12位devArea+16位dev)
}

func (i *iotIdSerial) next() int64 {
	i.l.Lock()
	defer i.l.Unlock()

	now := time.Now().UnixMilli()
	if i.time == now {
		i.serial++
		if i.serial >= snowSerialLimit {
			// 等待时间轮转
			for now <= i.time {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		i.serial = 0
		i.time = now
	}
	return now
}

func (i *iotIdSerial) Uuid() uuid {
	// 生成id
	tId := int64((i.next()-zeroTime)<<int64(iotDevLen) | (i.serial))

	var idg [16]byte = [16]byte{}

	binary.BigEndian.PutUint64(idg[8:16], uint64(tId))
	binary.BigEndian.PutUint64(idg[0:8], uint64(i.dev))
	return idg
}

// 生成一个16位字符串
func (i *iotIdSerial) Str() string {
	tId := int64((i.next()-zeroTime)<<int64(iotDevLen) | (i.serial))

	var idg [16]byte = [16]byte{}

	binary.BigEndian.PutUint64(idg[8:16], uint64(tId))
	binary.BigEndian.PutUint64(idg[0:8], uint64(i.dev))

	return base64.RawURLEncoding.EncodeToString(idg[4:16])
}
