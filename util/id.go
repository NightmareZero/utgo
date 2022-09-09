package util

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"
)

var (
	random1 = rand.Reader
)

// 生成短id
func ShortId() string {
	uuid := uuidR16()
	return base64.RawURLEncoding.EncodeToString(uuid[:])
}

func Uuid() string {
	uuid := uuidR16()
	target := [36]byte{}

	hex.Encode(target[:], uuid[:4])
	target[8] = '-'
	hex.Encode(target[9:13], uuid[4:6])
	target[13] = '-'
	hex.Encode(target[14:18], uuid[6:8])
	target[18] = '-'
	hex.Encode(target[19:23], uuid[8:10])
	target[23] = '-'
	hex.Encode(target[24:], uuid[10:])

	return string(target[:])
}

func uuidR16() (r16 [16]byte) {
	io.ReadFull(random1, r16[:])

	r16[6] = (r16[6] & 0x0f) | 0x40 // Version 4
	r16[8] = (r16[8] & 0x3f) | 0x80 // Variant is 10

	return
}
