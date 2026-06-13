package errors

import "net/http"

// AppError 业务错误
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string { return e.Message }

// 预定义错误码
var (
	ErrInvalidParam  = &AppError{Code: 40001, Message: "参数错误"}
	ErrUnauthorized  = &AppError{Code: 40101, Message: "未登录或登录已过期"}
	ErrForbidden     = &AppError{Code: 40301, Message: "无权限"}
	ErrNotFound      = &AppError{Code: 40401, Message: "资源不存在"}
	ErrConflict      = &AppError{Code: 40901, Message: "数据冲突"}
	ErrInternal      = &AppError{Code: 50001, Message: "服务器内部错误"}
)

// HTTPStatus 返回 HTTP 状态码
func HTTPStatus(err *AppError) int {
	switch {
	case err.Code < 40000:
		return http.StatusOK
	case err.Code < 41000:
		return http.StatusBadRequest
	case err.Code < 42000:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
