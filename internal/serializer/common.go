package serializer

// Response 基础序列化器
type Response struct {
	Code  int         `json:"code"`
	Data  interface{} `json:"data,omitempty"`
	Msg   string      `json:"msg"`
	Error string      `json:"error,omitempty"`
}

// TrackedErrorResponse 有追踪信息的错误响应
type TrackedErrorResponse struct {
	Response
	TrackID string `json:"track_id"`
}

// 错误代码常量
const (
	CodeSuccess         = 0
	CodeError           = 500
	CodeUnauthorized    = 401
	CodeForbidden       = 403
	CodeNotFound        = 404
	CodeBadRequest      = 400
	CodeTooManyRequests = 429
	CodeValidationError = 422
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

// DBErr 数据库错误
func DBErr(msg string, err error) Response {
	return Err(CodeError, msg, err)
}

// ParamErr 参数错误
func ParamErr(msg string, err error) Response {
	return Err(CodeBadRequest, msg, err)
}
