package miniofs

import (
	"context"
	"fmt"
	"time"

	"github.com/NightmareZero/nzgoutil/fio"
	"github.com/minio/madmin-go/v2"
)

// MinioClusterExtra 包装 madmin.InfoMessage，作为 fio.ClusterInfo.Extra 的载体。
// 暴露给需要访问节点详情 / Pool / 磁盘明细 / EC 配置等 minio 专有字段的调用方。
type MinioClusterExtra struct {
	Info *madmin.InfoMessage
}

// String 返回 minio 集群专有信息的摘要 (节点数 / 在线磁盘 / EC 配置 / DeploymentID)
func (e *MinioClusterExtra) String() string {
	if e == nil || e.Info == nil {
		return "<nil>"
	}
	info := e.Info

	// 节点统计
	var onlineNodes, totalDrives, onlineDrives, offlineDrives int
	for _, srv := range info.Servers {
		if srv.State == "online" {
			onlineNodes++
		}
		for _, d := range srv.Disks {
			totalDrives++
			if d.State == "ok" {
				onlineDrives++
			} else {
				offlineDrives++
			}
		}
	}

	// 后端细节
	backend := info.Backend
	ecStr := "N/A"
	if info.BackendType() == madmin.Erasure {
		ecStr = fmt.Sprintf("%d+%d", backend.StandardSCParity, backend.RRSCParity)
	}

	return fmt.Sprintf("MinioClusterExtra{Nodes=%d/%d, Drives=%d online / %d offline, EC=%s, DeploymentID=%s, Region=%s}",
		onlineNodes, len(info.Servers),
		onlineDrives, offlineDrives,
		ecStr, info.DeploymentID, info.Region,
	)
}

// ClusterInfo implements fio.IFileSystem.
// 调用 madmin.ServerInfo 获取集群信息，聚合在线磁盘的物理容量填入通用字段，
// 原始 InfoMessage 通过 Extra 返回，便于调用方获取节点 / Pool / 磁盘明细。
func (m *MinIOFileSystem) ClusterInfo(ctx context.Context) (fio.ClusterInfo, error) {
	cctx, cf := context.WithTimeout(ctx, 10*time.Second)
	defer cf()

	info, err := m.mc.ServerInfo(cctx)
	if err != nil {
		return fio.ClusterInfo{}, fmt.Errorf("error on get server info: %w", err)
	}

	// 聚合在线磁盘 (State == "ok") 的物理容量
	// 注意: 每个节点只列自己的磁盘，不会跨节点重复
	var total, used uint64
	for _, srv := range info.Servers {
		for _, d := range srv.Disks {
			if d.State == "ok" {
				total += d.TotalSpace
				used += d.UsedSpace
			}
		}
	}

	var free uint64
	if total > used {
		free = total - used
	}

	return fio.ClusterInfo{
		BackendType:   string(info.Backend.Type),
		Mode:          info.Mode,
		TotalCapacity: total,
		UsedCapacity:  used,
		FreeCapacity:  free,
		BucketCount:   info.Buckets.Count,
		ObjectCount:   info.Objects.Count,
		VersionCount:  info.Versions.Count,
		Extra:         &MinioClusterExtra{Info: &info},
	}, nil
}
