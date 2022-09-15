package idg

import (
	"sync"
	"time"
)

const (
	zeroTime    int64 = 1577808001000 // 雪花id时间戳起始点
	gidLen      uint8 = 10            // 生成器序号长度
	serialLen   uint8 = 12            // 自增筛选
	gidLimit    int64 = 0x3FF         // 1024  10 bits
	serialLimit int64 = 0xFFF         // 4096  10 bits
	timeOffset  uint8 = gidLen + serialLen
)

type snowflaker struct {
	l      sync.Mutex // 工作锁
	time   int64      // 时间戳
	gid    int64      // 生成器序号 (0~1023)
	serial int64      // 自增序列号
}

func (w *snowflaker) Get() int64 {
	w.l.Lock()
	defer w.l.Unlock()

	now := time.Now().UnixMilli()
	if w.time == now {
		w.serial++
		if w.serial >= serialLimit {
			// 等待时间轮转
			for now <= w.time {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		w.serial = 0
		w.time = now
	}
	newId := int64((now-zeroTime)<<timeOffset | (w.gid << serialLen) | (w.serial))
	return newId
}
