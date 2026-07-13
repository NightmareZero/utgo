package miniofs_test

import (
	"context"
	"testing"

	"github.com/NightmareZero/nzgoutil/fio"
	miniofs "github.com/NightmareZero/nzgoutil/fio/minio"
)

// TestClusterInfo 验证 minio 实现 ClusterInfo 返回结构正确性。
// 注意: 本测试需要可用的 minio 实例 (beforeTest 配置)，未实际部署时请跳过。
func TestClusterInfo(t *testing.T) {
	beforeTest()

	info, err := fio.FileSystem.ClusterInfo(context.Background())
	if err != nil {
		t.Fatalf("ClusterInfo failed: %v", err)
	}

	// BackendType 应是非空字符串 (Erasure / FS / Gateway)
	if info.BackendType == "" {
		t.Errorf("BackendType should not be empty")
	}

	// 总容量应当大于 0
	if info.TotalCapacity == 0 {
		t.Logf("warning: TotalCapacity is 0 (cluster may have no online disks)")
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

	// Extra 不应为 nil，且类型应为 *miniofs.MinioClusterExtra
	if info.Extra == nil {
		t.Fatalf("Extra should not be nil for minio implementation")
	}
	extra, ok := info.Extra.(*miniofs.MinioClusterExtra)
	if !ok {
		t.Fatalf("Extra should be *miniofs.MinioClusterExtra, got %T", info.Extra)
	}
	if extra.Info == nil {
		t.Fatalf("MinioClusterExtra.Info should not be nil")
	}

	// 原始 InfoMessage.Servers 应当非空 (集群至少有一个节点)
	if len(extra.Info.Servers) == 0 {
		t.Errorf("InfoMessage.Servers should not be empty")
	}

	t.Logf("ClusterInfo: backend=%s, mode=%s, total=%d, used=%d, free=%d, buckets=%d, objects=%d, versions=%d, nodes=%d",
		info.BackendType, info.Mode,
		info.TotalCapacity, info.UsedCapacity, info.FreeCapacity,
		info.BucketCount, info.ObjectCount, info.VersionCount,
		len(extra.Info.Servers),
	)

	// 验证 String() 输出
	t.Logf("ClusterInfo.String() = %s", info.String())
	t.Logf("MinioClusterExtra.String() = %s", extra.String())
}
