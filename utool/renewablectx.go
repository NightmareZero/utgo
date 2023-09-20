package utool

import (
	"context"
	"time"
)

var _ context.Context = &RenewableContext{}

type RenewableContext struct {
	ctx    context.Context
	cancel context.CancelFunc
	timer  *time.Timer
}

func newRenewableContext(parent context.Context, timeout time.Duration) *RenewableContext {
	ctx, cancel := context.WithCancel(parent)
	return &RenewableContext{
		ctx:    ctx,
		cancel: cancel,
		timer:  time.NewTimer(timeout),
	}
}

func (rc *RenewableContext) Done() <-chan struct{} {
	return rc.ctx.Done()
}

func (*RenewableContext) Deadline() (deadline time.Time, ok bool) {
	return deadline, false
}

func (rc *RenewableContext) Err() error {
	return rc.ctx.Err()
}

func (rc *RenewableContext) Value(key interface{}) interface{} {
	return rc.ctx.Value(key)
}

func (rc *RenewableContext) Renew(timeout time.Duration) {
	rc.timer.Reset(timeout)
}

func (rc *RenewableContext) Cancel() {
	defer rc.cancel()
	defer rc.timer.Stop()
}

func CtxWithRenewableTimeout(parent context.Context, timeout time.Duration) (*RenewableContext, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	rc := newRenewableContext(ctx, timeout)
	go func() {
		for {
			select {
			case <-rc.timer.C:
				rc.Cancel()
			case <-rc.ctx.Done():
				rc.Cancel()
				return
			}
		}
	}()
	return rc, cancel
}
