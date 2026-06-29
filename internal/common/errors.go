package common

const (
	CodeSuccess       = 0
	CodeBadRequest    = 400
	CodeUnauthorized  = 401
	CodeForbidden     = 403
	CodeNotFound      = 404
	CodeInternalError = 500

	CodeUserNotFound      = 1001
	CodeUserDisabled      = 1002
	CodePasswordError     = 1003
	CodeTokenExpired      = 1004
	CodeTokenInvalid      = 1005
	CodeCaptchaError      = 1006
	CodeUserAlreadyExists = 1007

	CodeRoleNotFound      = 2001
	CodeRoleAlreadyExists = 2002
	CodeMenuNotFound      = 2003
	CodeDeptNotFound      = 2004

	CodeFileTooLarge   = 3001
	CodeFileNotAllowed = 3002
	CodeUploadFailed   = 3003
)

var ErrorCodeMessages = map[int]string{
	CodeSuccess:            "成功",
	CodeBadRequest:         "请求参数错误",
	CodeUnauthorized:       "未登录或Token已过期",
	CodeForbidden:          "没有权限",
	CodeNotFound:           "资源不存在",
	CodeInternalError:      "服务器内部错误",
	CodeUserNotFound:       "用户不存在",
	CodeUserDisabled:       "用户已被禁用",
	CodePasswordError:      "密码错误",
	CodeTokenExpired:       "Token已过期",
	CodeTokenInvalid:       "Token无效",
	CodeCaptchaError:       "验证码错误",
	CodeUserAlreadyExists:  "用户已存在",
	CodeRoleNotFound:       "角色不存在",
	CodeRoleAlreadyExists:  "角色已存在",
	CodeMenuNotFound:       "菜单不存在",
	CodeDeptNotFound:       "部门不存在",
	CodeFileTooLarge:       "文件太大",
	CodeFileNotAllowed:     "文件类型不允许",
	CodeUploadFailed:       "上传失败",
}
