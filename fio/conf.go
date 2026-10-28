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
	Secure   bool // FilerURL 为 SeaweedFS Filer HTTP 地址（含 scheme，如 "http://localhost:8888"），
	// 用于追加写（走 Filer ?op=append）。为空则追加功能降级为 read-modify-write。
	FilerURL string
}
