# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目性质

这是一个 **Go 教学项目**：用 Go 从零手写一个命令行多线程分段下载器（`jason-downloader`），核心目的是学习 Go 语言特点。用户是 Go 初学者，期望以 **分步教学的方式**推进——每实现一个 Step 就停下来讲解、让用户自己动手，而不是一次性写完所有代码。

沟通语言：中文。代码注释以中文为主，用于解释 Go 的语言特性与设计取舍。

**当前状态（重要）**：`go 1.24.0` 已安装，Git 仓库已初始化但**尚无任何提交**，仓库内**还没有任何 Go 代码**（`go.mod` 都尚未创建）。完整设计方案见已批准的计划文件：`C:\Users\22619\.claude\plans\go-go-role-go-golang-vast-possum.md`。

## 常用命令

代码就位后使用：

```bash
go build ./...          # 构建
go vet ./...            # 静态检查
go test ./...           # 全部测试
go test ./internal/downloader/ -run TestSplitSegments -v   # 单个测试
go run . -u <URL> -c 8  # 运行下载器
go run . -list urls.txt -dir downloads
```

技术约束：**纯标准库，零第三方依赖**（net/http、os、io、flag、sync/atomic、encoding/json、context）。测试用 `httptest` + `http.ServeContent` 起本地 Range 服务器，不引入外部依赖。

## 目标架构

核心并发模型：**固定分片，一 goroutine 一段**（`-c` 即分段数）。每段下载到独立 `.partN` 文件（断点续传精确到段的前提），用 `sync.WaitGroup` 收口。进度用 `sync/atomic` 计数器无锁累计，错误经带缓冲 channel 上报，取消经 `context` 传播。

```
main.go                        CLI 入口：flag 解析、单任务/批量调度、退出码
internal/downloader/           核心：task(分段模型) probe(探测) worker(单段下载) download(编排/合并) resume(断点状态)
internal/progress/             原子计数 Tracker + 终端 \r 进度条渲染
internal/queue/                urls.txt 解析、从 URL 推断文件名
```

关键决策（改动前先读）：
- 段文件与状态文件 `<输出名>.jdl.json` 是断点续传的载体；下载完成后必须清理。
- 服务器不支持/无视 `Range` 时自动降级为单连接（`ErrRangeIgnored` 触发整体重试）。
- 未知大小的文件不支持续传（无法定位断点）。
- `internal/` 包约束是本项目的刻意教学点之一。

## 分步实施路线（教学节奏）

1. Step 0：初始化项目（`go mod init`——**用户自己执行**）
2. Step 1：最小单连接版（net/http、defer、io.Copy、error 哲学）
3. Step 2：探测 + 分段 + 并发下载（goroutine、WaitGroup、Range）
4. Step 3：用 channel 汇总进度
5. Step 4：实时进度条（sync/atomic、time.Ticker、`\r` 刷新）
6. Step 5：断点续传 + 优雅退出（encoding/json、context、signal.NotifyContext）
7. Step 6：批量队列 + 错误隔离
8. Step 7：httptest 表驱动测试 + README

每步应可独立运行验证（下载后对比 MD5）。建议在每个 Step 完成后提交一个 git 提交点，方便随时回退。