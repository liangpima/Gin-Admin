package common

const (
	StatusEnabled  = 1
	StatusDisabled = 0

	MenuTypeDir    = 0
	MenuTypeMenu   = 1
	MenuTypeButton = 2

	SuperAdminID = 1

	ContextKeyTenantID = "tenant_id"
	ContextKeyUserID   = "user_id"
	ContextKeyUsername = "username"
	ContextKeyRoles    = "roles"
	ContextKeyDeptID   = "dept_id"

	HeaderTenantID = "X-Tenant-Id"
	HeaderUserID   = "X-User-Id"
)
