package localfs_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/NightmareZero/nzgoutil/fio/localfs"
)

// newLocalFS 创建一个基于临时目录的 LocalFileSystem 用于测试。
func newLocalFS(t *testing.T) *localfs.LocalFileSystem {
	t.Helper()
	dir := t.TempDir()
	fs := &localfs.LocalFileSystem{BasePath: filepath.Join(dir, "files")}
	if err := os.MkdirAll(fs.BasePath, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	fs.Init()
	return fs
}

// TestClusterInfo 验证 localfs 实现 ClusterInfo 返回本地磁盘容量信息。
func TestClusterInfo(t *testing.T) {
	fs := newLocalFS(t)

	info, err := fs.ClusterInfo(context.Background())
	if err != nil {
		t.Fatalf("ClusterInfo failed: %v", err)
	}

	// BackendType 固定为 LocalFS
	if info.BackendType != "LocalFS" {
		t.Errorf("BackendType = %q, want %q", info.BackendType, "LocalFS")
	}

	// localfs 不填 Mode
	if info.Mode != "" {
		t.Errorf("Mode = %q, want empty", info.Mode)
	}

	// 总容量应当大于 0 (任何真实磁盘都不为 0)
	if info.TotalCapacity == 0 {
		t.Errorf("TotalCapacity should be > 0")
	}

	// 已用不应大于总容量
	if info.UsedCapacity > info.TotalCapacity {
		t.Errorf("UsedCapacity (%d) > TotalCapacity (%d)",
			info.UsedCapacity, info.TotalCapacity)
	}

	// FreeCapacity 应满足 total = used + free
	if info.TotalCapacity != info.UsedCapacity+info.FreeCapacity {
		t.Errorf("capacity mismatch: total=%d, used=%d, free=%d",
			info.TotalCapacity, info.UsedCapacity, info.FreeCapacity)
	}

	// localfs 不填桶/对象/版本数
	if info.BucketCount != 0 || info.ObjectCount != 0 || info.VersionCount != 0 {
		t.Errorf("counters should be 0 for localfs, got buckets=%d objects=%d versions=%d",
			info.BucketCount, info.ObjectCount, info.VersionCount)
	}

	// localfs Extra 必须为 nil
	if info.Extra != nil {
		t.Errorf("Extra should be nil for localfs, got %T", info.Extra)
	}

	t.Logf("ClusterInfo: backend=%s, total=%d, used=%d, free=%d",
		info.BackendType, info.TotalCapacity, info.UsedCapacity, info.FreeCapacity,
	)
}
