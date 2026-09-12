// Command jason-downloader —— 程序入口。
//
// Step 1：最小单连接下载器。
// 目标是把 "HTTP 请求 → 流式落盘" 这条最基础的链路跑通，
// 顺便认识 Go 里最高频的四个习惯性写法：
//   - error 是普通的值，用 if 显式检查，没有 try/catch
//   - 资源打开后立刻 defer 安排关闭，"谁打开谁负责"
//   - io.Copy 当搬运工
//   - flag 解析命令行参数，非法输入用退出码区分
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

func main() {
	// 1. 命令行参数。flag 包自动支持 -h / --help。
	u := flag.String("u", "", "要下载的 URL")
	o := flag.String("o", "", "保存的文件名（缺省时从 URL 推断）")
	flag.Parse()

	if *u == "" {
		fmt.Fprintln(os.Stderr, "请用 -u 指定要下载的 URL")
		flag.Usage()
		os.Exit(2) // 2 = 用法错误（惯例），1 = 运行错误
	}

	// 取名：优先用户指定，其次从 URL 路径最后一段推断。
	out := *o
	if out == "" {
		out = path.Base(*u) // path 包按 '/' 切路径，和网址的语义一致
		if out == "" || out == "." || out == "/" {
			out = "index.html"
		}
	}

	// 2. 发请求。拿到 *http.Response 后，第一件事就是 defer 关 Body——
	// "资源一打开就安排上关闭"，函数无论走哪条分支都不会泄漏连接。
	resp, err := http.Get(*u)
	if err != nil {
		fmt.Fprintln(os.Stderr, "请求失败:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "服务器返回异常状态码：%s\n", resp.Status)
		os.Exit(1)
	}

	// 3. 建本地文件。os.Create 会创建或清空目标文件。
	f, err := os.Create(out)
	if err != nil {
		fmt.Fprintln(os.Stderr, "创建文件失败:", err)
		os.Exit(1)
	}
	defer f.Close() // 同样的惯例：创建后立即安排关闭

	// 4. 搬运。io.Copy(dst, src) 反复把 src 读到 EOF、边读边写，
	//    返回搬运的字节数和第一个错误。这是标准库最高频的"水泵"。
	n, err := io.Copy(f, resp.Body)
	if err != nil {
		fmt.Fprintln(os.Stderr, "下载中断:", err)
		os.Exit(1)
	}

	// 5. 汇报。Go 没有异常机制：错误就是普通返回值，
	//    "检查 → 处理 → 传递" 全靠显式 if，出错路径和成功路径同样清晰。
	fmt.Printf("完成：%s（%.1f MB）\n", out, float64(n)/(1024*1024))
}