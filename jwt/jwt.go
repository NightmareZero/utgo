package jwt3

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"
	"time"

	"github.com/NightmareZero/nzgoutil/idg"
	"github.com/gbrlsnchs/jwt/v3"
)

// 生成Jwt生成器
func NewJwtGenrator[T any](key []byte, container T) (jg *JwtGenerator[T], err error) {
	jg = &JwtGenerator[T]{
		key: key,
	}

	keyReader := bytes.NewReader(key)
	jg.pvKey, err = ecdsa.GenerateKey(elliptic.P256(), keyReader)
	if err != nil {
		return nil, fmt.Errorf("NewJwtGenrator: generate key error %w", err)
	}
	jg.pubKey = jg.pvKey.PublicKey

	jg.alg = jwt.NewES256(
		jwt.ECDSAPublicKey(&jg.pubKey),
		jwt.ECDSAPrivateKey(jg.pvKey),
	)

	return
}

type JwtGenerator[T any] struct {
	key    []byte
	pvKey  *ecdsa.PrivateKey
	pubKey ecdsa.PublicKey
	alg    *jwt.ECDSASHA
}

func (g *JwtGenerator[T]) NewToken(u T) (token []byte, err error) {
	now := time.Now()
	pl := JwtToken[T]{
		Payload: jwt.Payload{
			Issuer:         "dt01",
			Subject:        "login",
			Audience:       jwt.Audience{},
			ExpirationTime: jwt.NumericDate(now.Add(3 * time.Hour)),
			NotBefore:      jwt.NumericDate(now.Add(30 * time.Minute)),
			IssuedAt:       jwt.NumericDate(now),
			JWTID:          idg.UuidV1().Str22(),
		},
		Tag: u,
	}
	token, err = jwt.Sign(pl, g.alg)
	return
}

type JwtToken[T any] struct {
	jwt.Payload
	Tag T `json:"t,omitempty"`
}
