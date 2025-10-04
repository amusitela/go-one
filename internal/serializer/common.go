package serializer

// 统一的HTTP响应格式
type Response struct {
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

// 响应码常量
const (
	CodeSuccess         = 0   // 成功
	CodeBadRequest      = 400 // 请求参数错误
	CodeUnauthorized    = 401 // 未授权
	CodeForbidden       = 403 // 禁止访问
	CodeNotFound        = 404 // 资源不存在
	CodeTooManyRequests = 429 // 请求过于频繁
	CodeError           = 500 // 服务器错误
)

// Success 成功响应
func Success(msg string, data interface{}) Response {
	return Response{
		Code: CodeSuccess,
		Msg:  msg,
		Data: data,
	}
}

// Err 错误响应
func Err(code int, msg string, err error) Response {
	res := Response{
		Code: code,
		Msg:  msg,
	}
	if err != nil {
		res.Error = err.Error()
	}
	return res
}

// ParamErr 参数错误响应（400）
func ParamErr(msg string, err error) Response {
	return Err(CodeBadRequest, msg, err)
}

// DBErr 数据库错误响应（500）
func DBErr(msg string, err error) Response {
	return Err(CodeError, msg, err)
}
