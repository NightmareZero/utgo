package miniofs_test

import (
	"testing"
	"time"

	"github.com/NightmareZero/nzgoutil/fio"
)

func beforeTest() {
	// conf.LoadConf("../../../../run/config.yaml")
	// fioinit.Init(conf.Conf.Platform.MinIO)
}

func TestWrite(t *testing.T) {
	beforeTest()
	bucket, err := fio.FileSystem.Bucket("test1")
	if err != nil {
		t.Fatal(err)
	}

	file, err := bucket.OpenFile("test.txt")
	if err != nil {
		t.Fatal("fail on open file", err)
	}

	_, err = file.Write([]byte("hello world"))
	if err != nil {
		t.Fatal("fail on write data", err)
	}

	err = file.Close()
	if err != nil {
		t.Fatal("fail on close", err)
	}

	time.Sleep(5 * time.Second)

}

func TestList(t *testing.T) {
	beforeTest()
	bucket, err := fio.FileSystem.Bucket("test1")
	if err != nil {
		t.Fatal(err)
	}

	list, err := bucket.List("/")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(len(list), " files found:")
	for i, v := range list {
		t.Log(i, ": ", v.Name())
	}

}

func TestInfo(t *testing.T) {
	beforeTest()

	mis, err := fio.FileSystem.Bucket("test1")
	if err != nil {
		t.Fatal(err)
	}
	bs, err := mis.Info()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", bs)

}
