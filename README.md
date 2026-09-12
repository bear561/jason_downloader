# jason-downloader

用 Go 从零手写一个**命令行多线程分段下载器**——一个以学习 Go 语言为目的的教学项目。

纯标准库实现，零第三方依赖（`net/http`、`os`、`io`、`flag`、`sync`、`encoding/json`、`context`）。

## 当前进度

| Step | 内容 | 状态 |
| --- | --- | --- |
| 0 | 初始化项目（go.mod） | ✅ |
| 1 | 最小单连接版：HTTP → 流式落盘 | ✅ |
| 2 | 探测文件大小 + 分段 + 并发下载（goroutine/WaitGroup/Range） | ⬜ 计划中 |
| 3 | 用 channel 汇总下载进度 | ⬜ 计划中 |
| 4 | 实时进度条（sync/atomic + Ticker + `\r`） | ⬜ 计划中 |
| 5 | 断点续传 + Ctrl+C 优雅退出（json/context/signal） | ⬜ 计划中 |
| 6 | 批量 URL 队列 + 路径处理 | ⬜ 计划中 |
| 7 | httptest 表驱动测试 + README 收尾 | ⬜ 计划中 |

> 上面的 ⬜ 一行一行的填满，就是整个项目从入门到工程化的过程。

## 快速开始

```bash
go mod init jason-downloader   # （仅首次）
go run . -u <URL> -o a.txt     # 下载到当前目录 a.txt
```

### 需要的 Go 版本

Go 1.24+（`go.mod` 里声明为 `go 1.24.0`）。

## 命令行参数

| 参数 | 含义 | 状态 |
| --- | --- | --- |
| `-u` | 要下载的 URL | ✅ |
| `-o` | 保存的文件名（缺省时从 URL 推断） | ✅ |
| `-c` | 并发分段数（1 = 单连接，默认 4） | ⬜ Step 2 |
| `-resume` | 尝试从上次断点续传 | ⬜ Step 5 |
| `-list` | 批量：URL 列表文件，每行一个 | ⬜ Step 6 |
| `-dir` | 保存目录 | ⬜ Step 6 |

## 目标架构

```
main.go                        CLI 入口：flag 解析、任务调度、退出码
internal/downloader/           核心：分段模型 / 探测 / 单段下载 / 编排合并 / 断点续传
internal/progress/             原子计数 + 终端进度条渲染
internal/queue/                URL 列表解析、文件名推断
```

核心并发模型：**固定分片，一 goroutine 一段**——探测文件大小与 Range 支持情况后，把文件切成 `-c` 段，每段一个 goroutine 并发下载到独立的 `.partN` 文件，用 `sync.WaitGroup` 收口，最后按序合并。服务器不支持 Range 时自动降级为单连接。

## 学习路线图（每步对应的 Go 特性）

| Step | 你会学到 |
| --- | --- |
| 1 | `net/http`、`defer` 资源管理、`io.Copy`、error 值哲学、flag 与退出码 |
| 2 | 包与 `internal` 约束、HTTP Range / 206、goroutine、`sync.WaitGroup` |
| 3 | channel（方向、关闭约定、死锁预防） |
| 4 | `sync/atomic`、`time.Ticker`、终端 `\r` 刷新 |
| 5 | `encoding/json`、context 取消、`signal.NotifyContext` 优雅退出、原子写文件 |
| 6 | 批量调度、错误隔离、文件名清洗（Windows 路径陷阱） |
| 7 | `httptest` + `http.ServeContent` 本地 Range 服务器、表驱动测试 |

## 验证铁律

并发绝不能改变内容。任何 Step 完成后：

```bash
go run . -u <URL> -c 1 -o x1.bin
go run . -u <URL> -c 8 -o x8.bin
md5sum x1.bin x8.bin    # 两个 MD5 必须完全一致
```