# geektime-downloader

geektime-downloader 支持下载以下极客时间网站资源。

**极客时间**
- [x] 专栏(PDF/Markdown/音频)
- [x] 视频课
- [x] 每日一课
- [x] 大厂案例
- [x] 训练营视频
- [ ] 线下大会

**企业版极客时间**
- [ ] 体系课
- [ ] 每日一课
- [ ] 大厂案例
- [ ] 生态课
- [x] 训练营视频

部分资源暂未支持下载，欢迎PR。


[![go report card](https://goreportcard.com/badge/github.com/nicoxiang/geektime-downloader "go report card")](https://goreportcard.com/report/github.com/nicoxiang/geektime-downloader)
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)

## 快速开始

### 环境要求
- Chrome/Chromium 浏览器（用于PDF生成）
- Go 1.16+ （源码编译时需要）

### 安装方式

#### 方式1：下载预编译版本（推荐）
从 [Releases页面](https://github.com/nicoxiang/geektime-downloader/releases) 下载对应平台的可执行文件。

#### 方式2：从源码安装
```bash
# Go 1.16+
go install github.com/nicoxiang/geektime-downloader@latest

# Go version < 1.16
go get -u github.com/nicoxiang/geektime-downloader@latest
```

#### 方式3：从源码编译
```bash
# 克隆项目
git clone https://github.com/nicoxiang/geektime-downloader.git
cd geektime-downloader

# 使用构建脚本（推荐）
chmod +x build.sh
./build.sh              # 编译所有平台
./build.sh -p linux-amd64   # 编译指定平台

# 或手动编译当前平台
go build -o geektime-downloader
```

### 基本使用

#### 交互式模式（默认）
```bash
# 启动程序，按提示操作
geektime-downloader

# 或直接提供认证信息
geektime-downloader --gcid "your_gcid" --gcess "your_gcess"
```

#### 命令行模式（批量下载）
```bash
# 下载单个课程
geektime-downloader --gcid "xxx" --gcess "yyy" --course-ids 100056701 --product-type normal

# 批量下载多个课程
geektime-downloader --gcid "xxx" --gcess "yyy" --course-ids 100056701,100056702 --product-type normal
```

#### 配置文件模式（推荐）
```bash
# 创建配置文件
cp courses-example.yaml my-courses.yaml
# 编辑配置文件，添加您的课程信息

# 执行下载
geektime-downloader --config my-courses.yaml
```

## 命令行参数

```bash
geektime-downloader [flags]

Flags:
      --article-ids string      指定下载的文章ID，支持逗号分隔，例: 1,2,3
      --comments int            是否下载评论(0不下载,1下载首页评论,2下载所有评论) (default 1)
      --config string           配置文件路径，支持YAML格式的批量下载配置
      --course-ids string       课程ID列表，支持逗号分隔，例: 100056701,100056702
      --download-all            是否下载课程的所有内容 (default true)
      --enterprise              是否下载企业版极客时间资源
  -f, --folder string           专栏和视频课的下载目标位置
      --gcess string            极客时间 cookie 值 gcess
      --gcid string             极客时间 cookie 值 gcid
  -h, --help                    help for geektime-downloader
      --interval int            下载资源的间隔时间, 单位为秒 (default 1)
      --log-level string        日志记录级别(debug, info, warn, error, none) (default "info")
      --non-interactive         非交互模式标志（自动检测）
      --output int              专栏的输出内容(1pdf,2markdown,4audio)可自由组合 (default 1)
      --print-pdf-timeout int   Chrome生成PDF的超时时间, 单位为秒 (default 60)
      --print-pdf-wait int      Chrome生成PDF前的等待页面加载时间, 单位为秒 (default 5)
      --product-type string     产品类型: normal(普通课程), daily(每日一课), openclass(公开课), qconplus(大厂案例), university(训练营), other(其他)
  -q, --quality string          下载视频清晰度(ld标清,sd高清,hd超清) (default "sd")

更多详细使用说明请参考 [使用指南](doc/使用指南.md)。
```

## 文档

项目文档已整理到 `doc/` 目录下：

- **[使用指南](doc/使用指南.md)** - 完整的使用说明，包含所有功能和使用模式
- **[开发者指南](doc/开发者指南.md)** - 项目架构、开发指南和技术实现

### 多平台构建

使用 `build.sh` 脚本可以一键构建多个平台的可执行文件：

```bash
# 查看帮助
./build.sh --help

# 构建所有支持的平台
./build.sh

# 构建特定平台
./build.sh -p linux-amd64
./build.sh -p windows-amd64
./build.sh -p darwin-arm64

# 构建结果在 dist/ 目录
ls dist/
```

支持的平台：Windows (amd64/386/arm64), Linux (amd64/386/arm64/arm), macOS (amd64/arm64)

## Note

### 文件下载目标位置

文件下载目标位置可以通过 help 查看。默认情况下 Windows 位于 %USERPROFILE%/geektime-downloader 下；Unix, 包括 macOS, 位于 $HOME/geektime-downloader 下

### 如何查看课程 ID?

**普通课程：**

打开极客时间[课程列表页](https://time.geekbang.org/resource)，选择你想要查看的课程，在新打开的课程详情 Tab 页，查看 URL 最后的数字，例如下面的链接中 100056701 就是课程 ID：

```
https://time.geekbang.org/column/intro/100056701
```

**训练营课程：**

打开极客时间[训练营课程列表页](https://u.geekbang.org/schedule)，选择你想要查看的课程，在新打开的课程详情 Tab 页，查看 URL ```lesson/```后的数字，例如下面的链接中 419 就是课程 ID：

```
https://u.geekbang.org/lesson/419?article=535616
```

**每日一课课程：**

选择你想要下载的视频，查看 URL ```dailylesson/detail/```后的数字，例如下面的链接中 100122405 就是课程 ID：

```
https://time.geekbang.org/dailylesson/detail/100122405
```

**大厂案例课程：**

选择你想要下载的视频，查看 URL ```qconplus/detail/```后的数字，例如下面的链接中 100110494 就是课程 ID：

```
https://time.geekbang.org/qconplus/detail/100110494
```

**公开课课程：**

选择你想要下载的视频，查看 URL ```opencourse/intro/``` 或 ```opencourse/videointro/```后的数字，例如下面的链接中 100546701 就是课程 ID：

```
https://time.geekbang.org/opencourse/videointro/100546701
```

**其他：**

打开极客时间[我的课程-其他](https://time.geekbang.org/dashboard/course)，选择你想要查看的课程，在新打开的课程详情 Tab 页，查看 URL ```course/intro/``` 最后的数字，例如下面的链接中 100551201 就是课程 ID：

```
https://time.geekbang.org/course/intro/100551201
```

**企业版训练营：**

选择你想要查看的课程，查看 URL ```mall/product/```后的数字，例如下面的链接中 100618109 就是课程 ID：

```
https://b.geekbang.org/mall/product/100618109
```

### 为什么我下载的PDF是空白页?
首先下载课程请保证VPN已关闭。在此前提下如果仍然出现空白页情况，说明后台Chrome网页加载速度较慢，可以尝试加大--print-pdf-wait参数，保证页面完全加载完成后再开始生成PDF。

### 为什么我下载PDF一直提示超时?
首先下载课程请保证VPN已关闭。在此前提下如果下载持续出现超时，有可能是因为课程章节图片等内容较多，生成速度慢，比如课程《AI 绘画核心技术与实战》中的部分章节，可以尝试加大--print-pdf-timeout参数，并耐心等待。

### 如何下载专栏的 Markdown 格式和文章音频?

默认情况下载专栏的输出内容只有 PDF，可以通过 --output 参数按需选择是否需要下载 Markdown 格式和文章音频。比如 --output 3 就是下载 PDF 和 Markdown；--output 6 就是下载 Markdown 和音频；--output 7 就是下载所有。

Markdown 格式虽然显示效果上不及 PDF，但优势为可以显示完整的代码块（PDF 代码块在水平方向太长时会有缺失）并保留了原文中的超链接。

现在部分新课程的专栏文章中会包含视频，如课程《Kubernetes 入门实战课》等，目前程序会自动下载文章所包含的视频，视频目录在文章所在目录的子目录 videos 下，此类文章PDF的下载会耗费更多时间，请耐心等待。

### 退出程序和继续下载

Ctrl + C 退出程序。如果选择“下载所有”后中断程序，可重新进入程序继续下载。
