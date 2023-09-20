package utool_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/NightmareZero/nzgoutil/utool"
)

func TestRenewableContextTimeout(t *testing.T) {
	rc, cancel := utool.WithRenewableTimeout(context.Background(), time.Second*5)
	defer cancel()
	fmt.Printf("rc: %+v\n", rc)
	tc := 0
	go func() {
		for {
			select {
			case <-rc.Done():
				return
			default:
				fmt.Printf("do something : %v\n", tc)
				tc++
				time.Sleep(time.Second)
			}
		}
	}()
	go func() {
		time.Sleep(time.Second * 1)
		rc.Renew(time.Second * 5)
		time.Sleep(time.Second * 3)
		rc.Renew(time.Second * 5)
	}()
	<-rc.Done()
	fmt.Printf("final tc: %v", tc)
}
