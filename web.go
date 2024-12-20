package web

import "net/http"

// Context 每个传入的 HTTP 请求都会创建一个 Context 对象，并将其作为可选的第一个参数传递给处理函数。
// 它提供了关于请求的信息，包括 http.Request 对象、GET 和 POST 参数，并作为响应的 Writer。
type Context struct {
	Request *http.Request
	Params  map[string]string
	Server  *Server
	http.ResponseWriter
}

// WriteString 将字符串数据写入响应对象。
func (ctx *Context) WriteString(content string) {
	_, err := ctx.ResponseWriter.Write([]byte(content))
	if err != nil {
		return
	}
}

// Abort 是一个辅助方法，用于发送 HTTP 头和可选的响应体。
// 它对于返回 4xx 或 5xx 错误非常有用。
// 一旦调用了它，处理函数中的任何返回值都将不会写入响应。
func (ctx *Context) Abort(status int, body string) {
	ctx.SetHeader("Content-Type", "text/html; charset=utf-8", true)
	ctx.ResponseWriter.WriteHeader(status)
	_, err := ctx.ResponseWriter.Write([]byte(body))
	if err != nil {
		return
	}
}

// Redirect 是一个用于 3xx 重定向的辅助方法。
func (ctx *Context) Redirect(status int, url_ string) {
	ctx.ResponseWriter.Header().Set("Location", url_)
	ctx.ResponseWriter.WriteHeader(status)
	_, err := ctx.ResponseWriter.Write([]byte("Redirecting to: " + url_))
	if err != nil {
		return
	}
}

// BadRequest 写入一个 400 HTTP 响应。
func (ctx *Context) BadRequest() {
	ctx.ResponseWriter.WriteHeader(400)
}

// NotModified 写入一个 304 HTTP 响应。
func (ctx *Context) NotModified() {
	ctx.ResponseWriter.WriteHeader(304)
}

// Unauthorized 写入一个 401 HTTP 响应。
func (ctx *Context) unauthorized() {
	ctx.ResponseWriter.WriteHeader(401)
}

// Forbidden 写入一个 403 HTTP 响应。
func (ctx *Context) Forbidden() {
	ctx.ResponseWriter.WriteHeader(403)
}

// NotFound 写入一个 404 HTTP 响应。
func (ctx *Context) NotFound(message string) {
	ctx.ResponseWriter.WriteHeader(404)
	_, err := ctx.ResponseWriter.Write([]byte(message))
	if err != nil {
		return
	}

}

// SetHeader 设置响应头。如果 `unique` 为 true，则当前该头部的值将被覆盖。
// 如果为 false，则会将新值追加到该头部。
func (ctx *Context) SetHeader(hdr string, val string, unique bool) {
	if unique {
		ctx.Header().Set(hdr, val)
	} else {
		ctx.Header().Add(hdr, val)
	}
}
