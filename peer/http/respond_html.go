package http

import (
	"bytes"
	"github.com/bobwong89757/cellnet"
	"github.com/bobwong89757/cellnet/log"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// getExt 获取文件扩展名
// s: 文件路径或文件名
// 返回文件扩展名（包含点号），如果没有扩展名返回空字符串
func getExt(s string) string {
	if strings.Index(s, ".") == -1 {
		return ""
	}
	return "." + strings.Join(strings.Split(s, ".")[1:], ".")
}

// SetTemplateDir 设置模板文件目录
// dir: 模板文件所在的目录路径
// 用于加载 HTML 模板文件
func (self *httpAcceptor) SetTemplateDir(dir string) {

	self.templateDir = dir
}

// SetTemplateDelims 设置模板分隔符
// delimsLeft: 左分隔符，默认 "{{"
// delimsRight: 右分隔符，默认 "}}"
// 用于解决默认 {{ }} 与其他模板引擎冲突的问题
func (self *httpAcceptor) SetTemplateDelims(delimsLeft, delimsRight string) {
	self.delimsLeft = delimsLeft
	self.delimsRight = delimsRight
}

// SetTemplateExtensions 设置模板文件扩展名
// exts: 扩展名列表，默认包含 ".tpl" 和 ".html"
// 只有匹配扩展名的文件才会被识别为模板文件
func (self *httpAcceptor) SetTemplateExtensions(exts []string) {
	self.templateExts = exts
}

// SetTemplateFunc 设置模板函数
// f: 模板函数映射列表
// 用于在模板中调用自定义函数
func (self *httpAcceptor) SetTemplateFunc(f []template.FuncMap) {
	self.templateFuncs = f
}

// Compile 编译模板
// 遍历模板目录，加载所有匹配扩展名的模板文件
// 应用分隔符和模板函数
// 返回编译后的模板对象
func (self *httpAcceptor) Compile() *template.Template {

	// 设置默认模板目录
	if self.templateDir == "" {
		self.templateDir = "."
	}

	// 设置默认扩展名
	if len(self.templateExts) == 0 {
		self.templateExts = []string{".tpl", ".html"}
	}

	// 创建模板对象
	t := template.New(self.templateDir)

	// 设置分隔符
	t.Delims(self.delimsLeft, self.delimsRight)
	// parse an initial template in case we don't have any
	//template.Must(t.Parse("Martini"))

	// 遍历模板目录，加载所有模板文件
	filepath.Walk(self.templateDir, func(path string, info os.FileInfo, err error) error {
		// 获取相对路径
		r, err := filepath.Rel(self.templateDir, path)
		if err != nil {
			return err
		}

		// 获取扩展名
		ext := getExt(r)

		// 检查是否匹配配置的扩展名
		for _, extension := range self.templateExts {
			if ext == extension {

				// 读取模板文件
				buf, err := ioutil.ReadFile(path)
				if err != nil {
					panic(err)
				}

				// 生成模板名称（去掉扩展名）
				name := r[0 : len(r)-len(ext)]
				tmpl := t.New(filepath.ToSlash(name))

				// 添加模板函数
				for _, funcs := range self.templateFuncs {
					tmpl.Funcs(funcs)
				}

				// 解析模板，如果失败则 panic（我们不希望静默启动失败）
				template.Must(tmpl.Parse(string(buf)))
				break
			}
		}

		return nil
	})

	return t
}

// HTMLRespond HTML 响应处理器
// 实现 RespondProc 接口，用于返回 HTML 模板渲染的响应
type HTMLRespond struct {
	// StatusCode HTTP 状态码
	StatusCode int

	// PageTemplate 页面模板名称
	// 用于查找对应的模板文件
	PageTemplate string

	// TemplateModel 模板数据模型
	// 传递给模板进行渲染
	TemplateModel interface{}
}

// WriteRespond 写入响应（实现 RespondProc 接口）
// ses: HTTP Session
// 使用模板渲染 HTML 并写入 HTTP 响应
// 返回处理错误
func (self *HTMLRespond) WriteRespond(ses *httpSession) error {

	peerInfo := ses.Peer().(cellnet.PeerProperty)

	// 记录发送日志
	log.GetLog().Debugf("#http.send(%s) '%s' %s | [%d] HTML %s",
		peerInfo.Name(),
		ses.req.Method,
		ses.req.URL.Path,
		self.StatusCode,
		self.PageTemplate)

	// 创建缓冲区
	buf := make([]byte, 64)

	bb := bytes.NewBuffer(buf)
	bb.Reset()

	// 执行模板渲染
	err := ses.t.ExecuteTemplate(bb, self.PageTemplate, self.TemplateModel)

	if err != nil {
		return err
	}

	// 模板渲染成功，写入响应
	ses.resp.Header().Set("Content-Type", "text/html")
	ses.resp.WriteHeader(self.StatusCode)
	io.Copy(ses.resp, bb)

	return nil
}

// TextRespond 文本响应处理器
// 实现 RespondProc 接口，用于返回纯文本响应
type TextRespond struct {
	// StatusCode HTTP 状态码
	StatusCode int

	// Text 要返回的文本内容
	Text string
}

// WriteRespond 写入响应（实现 RespondProc 接口）
// ses: HTTP Session
// 将文本内容写入 HTTP 响应
// 返回处理错误
func (self *TextRespond) WriteRespond(ses *httpSession) error {

	peerInfo := ses.Peer().(cellnet.PeerProperty)

	// 记录发送日志
	log.GetLog().Debugf("#http.send(%s) '%s' %s | [%d] HTML '%s'",
		peerInfo.Name(),
		ses.req.Method,
		ses.req.URL.Path,
		self.StatusCode,
		self.Text)

	// 设置响应头
	ses.resp.Header().Set("Content-Type", "text/html;charset=utf-8")
	ses.resp.WriteHeader(self.StatusCode)
	// 写入文本内容
	ses.resp.Write([]byte(self.Text))

	return nil
}
