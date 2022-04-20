package def

type Config struct {
	Path         string   `json:"path"`          // 项目所在的目录
	CoveragePath string   `json:"coverage_path"` // 覆盖文件路径
	DiffCommit   string   `json:"diff_commit"`   // 版本起始的commit
	DiffExclude  []string `json:"diff_exclude"`  // 需要排除的变更目录，不计算代码变更行数
	UnitExclude  []string `json:"unit_exclude"`  // 需要排除的单元测试覆盖率目录，不计算代码的单测覆盖率
	ShowDetail   bool     `json:"show_detail"`   // 是否需要展示未覆盖代码的明细
}
