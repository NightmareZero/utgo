package fio

type MinIO struct {
	Endpoint string
	Access   string
	Secret   string
	Secure   bool
}

// SeaweedFS 配置 SeaweedFS S3 网关后端。
// 与 MinIO 并列，由 fioinit.InitSeaweedFS 显式选择。
// 当前仅包含 S3 网关所需字段；Tier 2 实现 ClusterInfo 时需扩展 MasterURL / FilerURL。
type SeaweedFS struct {
	Endpoint string // SeaweedFS S3 网关地址，如 "localhost:8333"
	Access   string
	Secret   string
	Secure   bool
}
