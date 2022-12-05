package idg

import (
	"io"
	"net"
	"sync"
	"time"
)

var (
	infName string // 网卡名称
	infAddr []byte // 网卡地址

	lasttime uint64     // last time we returned
	clockSeq uint16     // clock sequence for this run
	timeMu   sync.Mutex // uuidV1时间锁
)

// 想指定的位置写入网卡地址
func setNodeInterface(macId []byte) bool {
	if infAddr != nil {
		copy(macId, infAddr)
	}

	iname, addr := getHardwareInterface() // null implementation for js
	if iname != "" && addr != nil {
		infName = iname
		copy(infAddr, addr)
		copy(macId, addr)
		return true
	}

	// 如果始终未获取到
	if addr == nil {
		infName = "_"
		io.ReadFull(random1, infAddr)
		copy(macId, infAddr)
		return true
	}
	return false
}

// 获取 uuid v1 需要的网卡mac地址
func getHardwareInterface() (string, []byte) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return "", nil
	}

	for _, ifs := range netInterfaces {
		if len(ifs.HardwareAddr) >= 6 && (infName == "" || infName == ifs.Name) {
			return ifs.Name, ifs.HardwareAddr
		}
	}
	return "", nil
}

const (
	lillian    = 2299160          // Julian day of 15 Oct 1582
	unix       = 2440587          // Julian day of 1 Jan 1970
	epoch      = unix - lillian   // Days between epochs
	g1582      = epoch * 86400    // seconds between epochs
	g1582ns100 = g1582 * 10000000 // 100s of a nanoseconds between epochs
)

func getTime() (int64, uint16, error) {
	timeMu.Lock()
	defer timeMu.Unlock()

	t := time.Now()

	// If we don't have a clock sequence already, set one.
	if clockSeq == 0 {
		setClockSequence(-1)
	}
	now := uint64(t.UnixNano()/100) + g1582ns100

	// If time has gone backwards with this clock sequence then we
	// increment the clock sequence
	if now <= lasttime {
		clockSeq = ((clockSeq + 1) & 0x3fff) | 0x8000
	}
	lasttime = now
	return int64(now), clockSeq, nil
}

func setClockSequence(seq int) {
	if seq == -1 {
		var b [2]byte
		io.ReadFull(random1, b[:]) // clock sequence
		seq = int(b[0])<<8 | int(b[1])
	}
	oldSeq := clockSeq
	clockSeq = uint16(seq&0x3fff) | 0x8000 // Set our variant
	if oldSeq != clockSeq {
		lasttime = 0
	}
}
