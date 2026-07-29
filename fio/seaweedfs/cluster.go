package seaweedfs

import (
	"context"

	"github.com/NightmareZero/nzgoutil/fio"
)

// ClusterInfo implements fio.IFileSystem.
// TODO(Tier 2): SeaweedFS 没有 madmin 等价物，集群容量/桶数/对象数需要
// 对接 master server (`:9333/cluster/status`) 或 filer (`:8888/cluster/status`)
// 的原生 HTTP API，并扩展 fio.SeaweedFS 配置增加 MasterURL / FilerURL 字段。
// 当前为空壳实现，明确返回 ErrNotSupported，避免调用方误判零值为有效数据。
func (s *SeaweedFSFileSystem) ClusterInfo(_ context.Context) (fio.ClusterInfo, error) {
	return fio.ClusterInfo{}, ErrNotSupported
}
