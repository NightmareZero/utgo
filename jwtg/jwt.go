package jwtg

import (
	"fmt"
	"time"

	"github.com/NightmareZero/nzgoutil/idg"
	"github.com/gbrlsnchs/jwt/v3"
)

// 生成Jwt生成器
func NewJwtGenrator[T any](key []byte, container T) (jg *JwtGenerator[T], err error) {
	if len(key) < 128 {
		return nil, fmt.Errorf("too short for private key, must be 128 bytes at least")
	}

	jg = &JwtGenerator[T]{key: key}
	jg.alg = jwt.NewHS256(key)

	if jg.ExpMinute == 0 {
		jg.ExpMinute = 30
	}

	return
}

type JwtGenerator[T any] struct {
	key       []byte
	alg       *jwt.HMACSHA
	ExpMinute int
}

type SignOption struct {
	ExpMinute int // 如果不设置，将使用JwtGenerator的ExpMinute
}

func (g *JwtGenerator[T]) Sign(u T, opt SignOption) (token []byte, err error) {
	exp := g.ExpMinute
	if opt.ExpMinute > 0 {
		exp = opt.ExpMinute
	}

	now := time.Now()
	pl := JwtToken[T]{
		Payload: jwt.Payload{
			Issuer:         "nz",
			Subject:        "token",
			Audience:       jwt.Audience{},
			ExpirationTime: jwt.NumericDate(now.Add(time.Duration(exp) * time.Minute)),
			NotBefore:      jwt.NumericDate(now.Add(-30 * time.Second)),
			IssuedAt:       jwt.NumericDate(now),
			JWTID:          idg.UuidV1().Str22(),
		},
		Tag: u,
	}

	token, err = jwt.Sign(pl, g.alg)
	return
}

func (g *JwtGenerator[T]) Refresh(token []byte, opt SignOption) (newToken []byte, err error) {
	t, err := g.Verify(token)
	if err != nil {
		return
	}
	now := time.Now()
	t.ExpirationTime = jwt.NumericDate(now.Add(time.Duration(opt.ExpMinute) * time.Minute))
	t.NotBefore = jwt.NumericDate(now.Add(-30 * time.Second))

	newToken, err = jwt.Sign(t, g.alg)
	return
}

func (g *JwtGenerator[T]) Refresh2(token []byte, opt SignOption, onRefresh func(t JwtToken[T]) (T, error)) (newToken []byte, err error) {
	t, err := g.Verify(token)
	if err != nil {
		return
	}

	if onRefresh != nil {
		t.Tag, err = onRefresh(t)
		if err != nil {
			return
		}
	}

	now := time.Now()
	t.ExpirationTime = jwt.NumericDate(now.Add(time.Duration(opt.ExpMinute) * time.Minute))
	t.NotBefore = jwt.NumericDate(now.Add(-30 * time.Second))

	newToken, err = jwt.Sign(t, g.alg)
	return
}

func (g *JwtGenerator[T]) Verify(b []byte) (token JwtToken[T], err error) {
	token = JwtToken[T]{}
	_, err = jwt.Verify(b, g.alg, &token)
	return
}

type JwtToken[T any] struct {
	jwt.Payload
	Tag T `json:"t,omitempty"`
}
