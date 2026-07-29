//go:build integration

// 本测试文件需要真实 SeaweedFS 实例（S3 网关），默认不参与 `go test ./...`。
// 运行方式：go test -tags=integration ./fio/seaweedfs/...
// 环境变量（可选，覆盖默认连接信息）：
//
//	SEAWEEDFS_TEST_ENDPOINT (默认 localhost:8333)
//	SEAWEEDFS_TEST_ACCESS   (默认 admin)
//	SEAWEEDFS_TEST_SECRET   (默认 secret)

package seaweedfs_test

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/NightmareZero/nzgoutil/fio"
	"github.com/NightmareZero/nzgoutil/fio/fioinit"
	"github.com/NightmareZero/nzgoutil/fio/seaweedfs"
)

// beforeTest 通过 S3 网关连接 SeaweedFS，配置可由环境变量覆盖。
func beforeTest() {
	cfg := fio.SeaweedFS{
		Endpoint: envOr("SEAWEEDFS_TEST_ENDPOINT", "localhost:8333"),
		Access:   envOr("SEAWEEDFS_TEST_ACCESS", "admin"),
		Secret:   envOr("SEAWEEDFS_TEST_SECRET", "secret"),
	}
	if err := fioinit.InitSeaweedFS(cfg); err != nil {
		panic(err)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// waitForObject 轮询 Stat 直到对象可见或超时，替代不稳定的 time.Sleep。
func waitForObject(t *testing.T, b fio.IFileBucket, name string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if _, err := b.Stat(name); err == nil {
			return
		} else if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("waitForObject stat unexpected error: %v", err)
		}
		if time.Now().After(deadline) {
			t.Fatalf("waitForObject timeout for %s", name)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// useBucket 获取/创建桶并注册清理（测试结束销毁桶），避免残留数据。
func useBucket(t *testing.T, name string, create bool) fio.IFileBucket {
	t.Helper()
	b, err := fio.FileSystem.Bucket(name, create)
	if err != nil {
		t.Fatalf("open bucket %s: %v", name, err)
	}
	t.Cleanup(func() {
		// 销毁失败不阻断测试结束，仅忽略
		_ = b.Destroy()
	})
	return b
}

func TestWrite(t *testing.T) {
	beforeTest()
	bucket := useBucket(t, "sw-write", true)

	localPath := writeLocalTestFile(t, 12*1024*1024)
	defer os.Remove(localPath)

	readFile, err := os.Open(localPath)
	if err != nil {
		t.Fatal("fail on open local file", err)
	}
	defer readFile.Close()

	file, err := bucket.OpenFile("test.txt")
	if err != nil {
		t.Fatal("fail on open file", err)
	}

	start := time.Now().UnixMilli()

	buf := make([]byte, 1024*1024)
	w, err := io.CopyBuffer(file, readFile, buf)
	if err != nil {
		t.Fatal("fail on write data", err)
	}

	if err := file.Close(); err != nil {
		t.Fatal("fail on close", err)
	}
	t.Log("copy bytes:", w, "in", time.Now().UnixMilli()-start, "ms")

	waitForObject(t, bucket, "test.txt", 10*time.Second)

	stat, err := bucket.Stat("test.txt")
	if err != nil {
		t.Fatal("fail on stat", err)
	}
	if stat.Size() != int64(w) {
		t.Fatalf("size mismatch: want %d, got %d", w, stat.Size())
	}
}

// TestEmptyFileCreate 验证 OpenFile().Close()（不写不读）会创建空对象，
// 与文件系统语义一致。
func TestEmptyFileCreate(t *testing.T) {
	beforeTest()
	bucket := useBucket(t, "sw-empty", true)

	file, err := bucket.OpenFile("empty.txt")
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	waitForObject(t, bucket, "empty.txt", 10*time.Second)

	stat, err := bucket.Stat("empty.txt")
	if err != nil {
		t.Fatal("empty file not created:", err)
	}
	if stat.Size() != 0 {
		t.Fatalf("empty file size want 0, got %d", stat.Size())
	}
}

func TestMerge(t *testing.T) {
	beforeTest()
	bucket := useBucket(t, "sw-merge", true)

	// 准备两个源文件
	for _, name := range []string{"merge_src1.txt", "merge_src2.txt"} {
		file, err := bucket.OpenFile(name)
		if err != nil {
			t.Fatal("fail on open file", err)
		}
		if _, err := file.Write([]byte("hello seaweedfs " + name + "\n")); err != nil {
			t.Fatal("fail on write", err)
		}
		if err := file.Close(); err != nil {
			t.Fatal("fail on close", err)
		}
	}

	waitForObject(t, bucket, "merge_src1.txt", 10*time.Second)
	waitForObject(t, bucket, "merge_src2.txt", 10*time.Second)

	stat, err := bucket.MergeFile(context.Background(),
		fio.MergeOption{Path: "merge_dst.txt"},
		fio.MergeOption{Path: "merge_src1.txt"},
		fio.MergeOption{Path: "merge_src2.txt"},
	)
	if err != nil {
		t.Fatal("fail on merge", err)
	}
	if stat == nil {
		t.Fatal("merge returned nil stat")
	}
	t.Log("merged size:", stat.Size())
}

func TestList(t *testing.T) {
	beforeTest()
	bucket := useBucket(t, "sw-list", true)

	// 写一个文件确保非空
	file, _ := bucket.OpenFile("marker.txt")
	file.Write([]byte("marker"))
	file.Close()
	waitForObject(t, bucket, "marker.txt", 10*time.Second)

	list, err := bucket.List("/")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(len(list), " files found:")
	for i, v := range list {
		t.Log(i, ": ", v.Name(), "size=", v.Size())
	}
}

func TestClusterInfo(t *testing.T) {
	beforeTest()
	// Tier 1 未实现 ClusterInfo，应返回 ErrNotSupported
	_, err := fio.FileSystem.ClusterInfo(context.Background())
	if !errors.Is(err, seaweedfs.ErrNotSupported) {
		t.Fatalf("ClusterInfo want ErrNotSupported, got err=%v", err)
	}
}

// TestCloseIdempotent 验证重复 Close 不会触发错误或重复上传。
func TestCloseIdempotent(t *testing.T) {
	beforeTest()
	bucket := useBucket(t, "sw-close", true)

	file, err := bucket.OpenFile("idempotent.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte("repeat close test")); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal("first close:", err)
	}
	// 重复 Close 应安全返回 nil
	if err := file.Close(); err != nil {
		t.Fatal("second close should be safe, got:", err)
	}

	waitForObject(t, bucket, "idempotent.txt", 10*time.Second)
}

// writeLocalTestFile 生成本地测试文件（指定大小），返回路径。
func writeLocalTestFile(t *testing.T, size int) string {
	t.Helper()
	f, err := os.CreateTemp("", "sw-src-*.bin")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	buf := make([]byte, 4096)
	for i := range buf {
		buf[i] = byte(i)
	}
	remaining := size
	for remaining > 0 {
		n := len(buf)
		if remaining < n {
			n = remaining
		}
		if _, err := f.Write(buf[:n]); err != nil {
			t.Fatal(err)
		}
		remaining -= n
	}
	return f.Name()
}
