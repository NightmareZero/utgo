package miniofs_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/NightmareZero/nzgoutil/fio"
	"github.com/NightmareZero/nzgoutil/fio/fioinit"
)

func beforeTest() {
	var cfg = fio.MinIO{
		Endpoint: "127.0.0.1:9000",
		Access:   "admin",
		Secret:   "password",
	}
	fioinit.Init(cfg)
}

func TestWrite(t *testing.T) {
	beforeTest()
	bucket, err := fio.FileSystem.Bucket("test", true)
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

func TestMerge(t *testing.T) {
	beforeTest()
	fil, err := os.ReadFile("/home/user/tmp/log1.log")
	if err != nil {
		t.Fatal(err)
	}

	bucket, err := fio.FileSystem.Bucket("test", false)
	if err != nil {
		t.Fatal(err)
	}

	file1, err := bucket.OpenFile("test1.txt")
	if err != nil {
		t.Fatal("fail on open file", err)
	}

	_, err = file1.Write(fil)
	if err != nil {
		t.Fatal("fail on write data", err)
	}

	err = file1.Close()
	if err != nil {
		t.Fatal("fail on close", err)
	}

	file2, err := bucket.OpenFile("test2.txt")
	if err != nil {
		t.Fatal("fail on open file", err)
	}

	_, err = file2.Write(fil)
	if err != nil {
		t.Fatal("fail on write data", err)
	}

	err = file2.Close()
	if err != nil {
		t.Fatal("fail on close", err)
	}

	time.Sleep(5 * time.Second)

	// merge
	_, err = bucket.MergeFile(context.Background(), fio.MergeOption{
		Path: "test/test.txt",
	}, fio.MergeOption{
		Path: "test/test1.txt",
	}, fio.MergeOption{
		Path: "test/test2.txt",
	}, fio.MergeOption{
		Path: "test/test3.txt",
	})
	if err != nil {
		t.Fatal("fail on merge", err)
	}

	time.Sleep(5 * time.Second)

}

func TestList(t *testing.T) {
	beforeTest()
	bucket, err := fio.FileSystem.Bucket("test1", false)
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

	mis, err := fio.FileSystem.Bucket("test1", false)
	if err != nil {
		t.Fatal(err)
	}
	bs, err := mis.Info()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", bs)

}

func TestQuota(t *testing.T) {
	beforeTest()

	mis, err := fio.FileSystem.Bucket("test1", false)
	if err != nil {
		t.Fatal(err)
	}
	var qv int64 = 512
	err = mis.SetConfig(fio.BucketConfig{
		Quota: &qv,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("set quota ok")

}
