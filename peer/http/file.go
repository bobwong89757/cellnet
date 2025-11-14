package http

import (
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

// SetFileServe 设置静态文件服务
// dir: 虚拟路径，如 "/static"
// root: 文件系统根目录，如 "/var/www"
// 用于提供静态文件服务
func (self *httpAcceptor) SetFileServe(dir string, root string) {

	self.httpDir = dir
	self.httpRoot = root
}

// GetDir 获取文件服务目录
// 如果 httpDir 是绝对路径，直接使用
// 否则将其与 httpRoot 组合
// 返回 http.Dir 类型，用于文件服务
func (self *httpAcceptor) GetDir() http.Dir {

	if filepath.IsAbs(self.httpDir) {
		return http.Dir(self.httpDir)
	} else {
		return http.Dir(filepath.Join(self.httpRoot, self.httpDir))
	}

	//workDir, _ := os.Getwd()
	//log.GetLog().Debugf("Http serve file: %s (%s)", self.dir, workDir)
}

// ServeFile 提供文件服务
// res: HTTP 响应写入器
// req: HTTP 请求对象
// dir: 文件目录
// 支持 GET 和 HEAD 方法，自动处理目录索引（index.html）
// 返回错误和是否已处理
func (self *httpAcceptor) ServeFile(res http.ResponseWriter, req *http.Request, dir http.Dir) (error, bool) {
	// 只支持 GET 和 HEAD 方法
	if req.Method != "GET" && req.Method != "HEAD" {
		return nil, false
	}

	file := req.URL.Path

	// 打开文件
	f, err := dir.Open(file)
	if err != nil {
		return errNotFound, false
	}
	defer f.Close()

	// 获取文件信息
	fi, err := f.Stat()
	if err != nil {
		return errNotFound, false
	}

	// 如果是目录，尝试提供 index.html
	if fi.IsDir() {
		// 如果路径末尾没有斜杠，重定向到带斜杠的路径
		if !strings.HasSuffix(req.URL.Path, "/") {
			dest := url.URL{
				Path:     req.URL.Path + "/",
				RawQuery: req.URL.RawQuery,
				Fragment: req.URL.Fragment,
			}
			http.Redirect(res, req, dest.String(), http.StatusFound)
			return nil, false
		}

		// 尝试打开 index.html
		file = path.Join(file, "index.html")
		f, err = dir.Open(file)
		if err != nil {
			return errNotFound, false
		}
		defer f.Close()

		fi, err = f.Stat()
		if err != nil || fi.IsDir() {
			return errNotFound, false
		}
	}

	// 使用 HTTP 标准库提供文件内容
	http.ServeContent(res, req, file, fi.ModTime(), f)

	return nil, true
}

// ServeFileWithDir 使用配置的目录提供文件服务
// res: HTTP 响应写入器
// req: HTTP 请求对象
// 获取配置的文件目录，然后调用 ServeFile 提供文件服务
// 返回消息、错误和是否已处理
func (self *httpAcceptor) ServeFileWithDir(res http.ResponseWriter, req *http.Request) (msg interface{}, err error, handled bool) {

	dir := self.GetDir()

	// 如果目录未配置，返回未找到错误
	if dir == "" {
		return nil, errNotFound, false
	}

	// 提供文件服务
	err, handled = self.ServeFile(res, req, dir)

	return
}
