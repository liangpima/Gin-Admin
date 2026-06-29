package model

type DashboardStats struct {
	UserCount   int64 `json:"userCount"`
	RoleCount   int64 `json:"roleCount"`
	MenuCount   int64 `json:"menuCount"`
	DeptCount   int64 `json:"deptCount"`
	PostCount   int64 `json:"postCount"`
	ConfigCount int64 `json:"configCount"`
	LogCount    int64 `json:"logCount"`
}
