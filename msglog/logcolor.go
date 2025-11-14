package msglog

// LogColorDefine 使用 github.com/bobwong89757/golog 的 cellnet 配色方案
// 定义了不同日志类型的颜色规则，用于美化日志输出
// 格式为 JSON，包含规则列表，每个规则指定文本匹配模式和颜色
const LogColorDefine = `
{
	"Rule":[
		{"Text":"panic:","Color":"Red"},
		{"Text":"[DB]","Color":"Green"},
		{"Text":"#http.listen","Color":"Blue"},
		{"Text":"#http.recv","Color":"Blue"},
		{"Text":"#http.send","Color":"Purple"},

		{"Text":"#tcp.listen","Color":"Blue"},
		{"Text":"#tcp.accepted","Color":"Blue"},
		{"Text":"#tcp.closed","Color":"Blue"},
		{"Text":"#tcp.recv","Color":"Blue"},
		{"Text":"#tcp.send","Color":"Purple"},
		{"Text":"#tcp.connected","Color":"Blue"},

		{"Text":"#ws.listen","Color":"Blue"},
		{"Text":"#ws.accepted","Color":"Blue"},
		{"Text":"#ws.closed","Color":"Blue"},
		{"Text":"#ws.recv","Color":"Blue"},
		{"Text":"#ws.send","Color":"Purple"},
		{"Text":"#ws.connected","Color":"Blue"},

		{"Text":"#udp.listen","Color":"Blue"},
		{"Text":"#udp.recv","Color":"Blue"},
		{"Text":"#udp.send","Color":"Purple"},

		{"Text":"#rpc.recv","Color":"Blue"},
		{"Text":"#rpc.send","Color":"Purple"},

		{"Text":"#relay.recv","Color":"Blue"},
		{"Text":"#relay.send","Color":"Purple"},

		{"Text":"#agent.recv","Color":"Blue"},
		{"Text":"#agent.send","Color":"Purple"}
	]
}
`
