package model

// Response 统一 API 响应结构
// 所有 API 都使用此结构返回数据，保证接口一致性
type Response struct {
	Code    int         `json:"code"`    // 状态码: 0 表示成功，非 0 表示错误
	Message string      `json:"message"` // 状态消息
	Data    interface{} `json:"data"`    // 响应数据
}

// 预定义错误码
const (
	CodeSuccess          = 0
	CodeBadRequest       = 400
	CodeUnauthorized     = 401
	CodeForbidden        = 403
	CodeNotFound         = 404
	CodeInternalError    = 500
	CodeInvalidFileType  = 1001
	CodeFileTooLarge     = 1002
	CodeProcessingFailed = 1003
	CodeStorageFailed    = 1004
)

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) Response {
	return Response{
		Code:    CodeSuccess,
		Message: "ok",
		Data:    data,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, message string) Response {
	return Response{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}

// PaginatedList 分页列表响应
type PaginatedList struct {
	Items      interface{} `json:"items"`       // 数据列表
	Total      int64       `json:"total"`       // 总数
	Page       int         `json:"page"`        // 当前页
	PageSize   int         `json:"page_size"`   // 每页数量
	TotalPages int         `json:"total_pages"` // 总页数
}
