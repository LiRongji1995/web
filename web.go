package web

import (
	"crypto/tls"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"
)

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

// ContentType 设置 HTTP 响应的 Content-Type 头。
// 例如，ctx.ContentType("json") 将 Content-Type 设置为 "application/json"。
// 如果提供的值包含斜杠（/），它将被直接用作 Content-Type。
// 返回值为设置的 Content-Type，如果未找到则返回空字符串。
func (ctx *Context) ContentType(val string) string {
	var ctype string
	if strings.Contains(val, "/") {
		ctype = val
	} else {
		if !strings.HasPrefix(val, ".") {
			val = "." + val
		}
		ctype = mime.TypeByExtension(val)
	}
	if ctype != "" {
		ctx.Header().Set("Content-Type", ctype)
	}
	return ctype
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

// SetCookie 向响应中添加一个 Cookie 头。
func (ctx *Context) SetCookie(cookie *http.Cookie) {
	ctx.SetHeader("Set-Cookie", cookie.String(), false)
}

// 小优化：缓存上下文类型，而不是反复调用 reflect.TypeOf。
var contextType reflect.Type

var defaultStaticDirs []string

func init() {
	contextType = reflect.TypeOf(Context{})
	// 查找可执行文件的位置。
	wd, _ := os.Getwd()
	arg0 := path.Clean(os.Args[0])
	var exeFile string
	if strings.HasPrefix(arg0, "/") {
		exeFile = arg0
	} else {
		// TODO: 为了提高健壮性，搜索 $PATH 中的每个目录。
		exeFile = path.Join(wd, arg0)
	}
	parent, _ := path.Split(exeFile)
	defaultStaticDirs = append(defaultStaticDirs, path.Join(parent, "static"))
	defaultStaticDirs = append(defaultStaticDirs, path.Join(wd, "static"))
	return
}

// Process 调用主服务器的路由系统。
func Process(c http.ResponseWriter, req *http.Request) {
	mainServer.Process(c, req)
}

// Run 启动 Web 应用程序并为主服务器处理 HTTP 请求。
func Run(addr string) {
	mainServer.Run(addr)
}

// RunTls 启动 Web 应用程序并为主服务器处理 HTTPS 请求。
func RunTls(add string, config *tls.Config) {
	mainServer.RunTls(add, config)
}

// RunScgi 启动 Web 应用程序并为主服务器处理 SCGI 请求。
func RunScgi(addr string) {
	mainServer.Runscgi(addr)
}

// RunFcgi 启动 Web 应用程序并为主服务器处理 FastCGI 请求。
func RunFcgi(addr string) {
	mainServer.RunFcgi(addr)
}

// Close 停止主服务器。
func Close() {
	mainServer.Close()
}

// Get 为主服务器的 'GET' HTTP 方法添加一个处理器。
func Get(route string, handler interface{}) {
	mainServer.Get(route, handler)
}

// Post 为主服务器的 'POST' HTTP 方法添加一个处理器。
func Post(route string, handler interface{}) {
	mainServer.addRoute(route, "Post", handler)
}

// Put 为主服务器的 'PUT' HTTP 方法添加一个处理器。
func Put(route string, handler interface{}) {
	mainServer.addRoute(route, "Put", handler)
}

// Delete 为主服务器的 'DELETE' HTTP 方法添加一个处理器。
func Delete(route string, handler interface{}) {
	mainServer.addRoute(route, "Delete", handler)
}

// Match 为主服务器的任意 HTTP 方法添加一个处理器。
func Match(method string, route string, handler interface{}) {
	mainServer.addRoute(route, method, handler)
}

// Handle 添加一个自定义的 http.Handler。在以 FCGI 或 SCGI 模式运行时将不起作用。
func Handle(route string, method string, httpHandler http.Handler) {
	mainServer.Handle(route, method, httpHandler)
}

// WebSocket 添加一个 WebSocket 的处理器。仅适用于 Web 服务器模式。
// 在以 FCGI 或 SCGI 模式运行时将不起作用。
func WebSocket(route string, httpHandler websocket.Handler) {
	mainServer.WebSocket(route, httpHandler)
}

// SetLogger 为主服务器设置日志记录器。
func SetLogger(logger *log.Logger) {
	mainServer.Logger = logger
}

// Config 是主服务器的配置。
var Config = &ServerConfig{
	RecoverPanic: true,
	ColorOutPut:  true,
}

var mainServer = NewServer()
