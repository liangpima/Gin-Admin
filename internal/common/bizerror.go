package common

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// BizError 业务错误：由调用方的输入或业务前置条件不满足导致，
// 与「数据库挂了」「签名失败」这类系统错误有本质区别。
//
// 背景：此前所有 Controller 把 service 返回的错误一律映射为 500，
// 于是「用户名已存在」「订单已关闭」这类本该是 400 的提示也变成了 500。
// 结果是前端只能靠文案区分错误类型，网关和监控也无法按状态码告警。
//
// 用法：service 层用 NewBizError / NewNotFoundError 显式标记，
// Controller 统一用 common.FailWith(c, err) 输出。
type BizError struct {
	Code int
	Msg  string
}

func (e *BizError) Error() string { return e.Msg }

// NewBizError 业务校验失败或不满足前置条件，对应 400
func NewBizError(msg string) error {
	return &BizError{Code: CodeBadRequest, Msg: msg}
}

// NewBizErrorf 带格式化的业务错误，对应 400
func NewBizErrorf(format string, args ...interface{}) error {
	return &BizError{Code: CodeBadRequest, Msg: fmt.Sprintf(format, args...)}
}

// NewNotFoundError 操作的目标资源不存在，对应 404。
// 与 400 的区别：400 是「请求本身有问题」，404 是「请求没问题但引用了不存在的东西」。
func NewNotFoundError(msg string) error {
	return &BizError{Code: CodeNotFound, Msg: msg}
}

// NotFoundOrErr 把「记录不存在」归一为 404 业务错误，其余错误原样返回。
//
// GORM 查询不到记录时返回的是 gorm.ErrRecordNotFound，它本身看不出是什么资源没找到，
// 直接透出会变成 500（且对外暴露的是空泛的内部错误）。各 Service 的单条查询应统一
// 用这里转成带语义的 404，Controller 侧只需 common.FailWith 即可。
func NotFoundOrErr(err error, msg string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NewNotFoundError(msg)
	}
	return err
}

// AsBizError 取出错误链上的业务错误；ok 为 false 表示这是系统错误
func AsBizError(err error) (*BizError, bool) {
	var be *BizError
	if err == nil {
		return nil, false
	}
	if !errors.As(err, &be) {
		return nil, false
	}
	return be, true
}

// IsBizError 判断是否为业务错误（供需要区分处理的地方使用）
func IsBizError(err error) bool {
	_, ok := AsBizError(err)
	return ok
}
