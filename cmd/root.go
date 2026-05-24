package cmd

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/nicoxiang/geektime-downloader/internal/config"
	"github.com/nicoxiang/geektime-downloader/internal/course"
	"github.com/nicoxiang/geektime-downloader/internal/fsm"
	"github.com/nicoxiang/geektime-downloader/internal/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
	"github.com/nicoxiang/geektime-downloader/internal/ui"
	"github.com/spf13/cobra"
	"golang.org/x/net/html"
	"gopkg.in/yaml.v3"
)

var (
	geektimeClient *geektime.Client
	cfg            config.AppConfig

	// 非交互式模式参数
	courseIDs      string
	productType    string
	downloadAll    bool
	nonInteractive bool
	articleIDs     string
	configFile     string
	listCourses    bool
)

// 配置文件相关数据结构
type CourseConfig struct {
	ID          int    `yaml:"id"`
	Type        string `yaml:"type"`
	DownloadAll bool   `yaml:"download_all"`
	ArticleIDs  []int  `yaml:"article_ids,omitempty"`

	// 可选的单个课程配置覆盖
	Quality         *string `yaml:"quality,omitempty"`
	Output          *int    `yaml:"output,omitempty"`
	Comments        *int    `yaml:"comments,omitempty"`
	Folder          *string `yaml:"folder,omitempty"`
	Interval        *int    `yaml:"interval,omitempty"`
	Concurrency     *int    `yaml:"concurrency,omitempty"`
	PrintPDFWait    *int    `yaml:"print_pdf_wait,omitempty"`
	PrintPDFTimeout *int    `yaml:"print_pdf_timeout,omitempty"`
	Enterprise      *bool   `yaml:"enterprise,omitempty"`
}

type GlobalConfig struct {
	GCID  string `yaml:"gcid,omitempty"`
	GCESS string `yaml:"gcess,omitempty"`

	Folder  string `yaml:"folder,omitempty"`
	Quality string `yaml:"quality,omitempty"`
	Output  int    `yaml:"output,omitempty"`

	Comments    int  `yaml:"comments,omitempty"`
	Interval    int  `yaml:"interval,omitempty"`
	Concurrency int  `yaml:"concurrency,omitempty"`

	PrintPDFWait    int `yaml:"print_pdf_wait,omitempty"`
	PrintPDFTimeout int `yaml:"print_pdf_timeout,omitempty"`

	Enterprise bool `yaml:"enterprise,omitempty"`
}

type BatchConfig struct {
	Global          *GlobalConfig  `yaml:"global,omitempty"`
	Courses         []CourseConfig `yaml:"courses,omitempty"`
	AdvancedCourses []CourseConfig `yaml:"advanced_courses,omitempty"`
}

func init() {
	userHomeDir, _ := os.UserHomeDir()
	defaultDownloadFolder := filepath.Join(userHomeDir, config.GeektimeDownloaderFolder)

	rootCmd.Flags().StringVar(&cfg.Gcid, "gcid", "", "极客时间 cookie 值 gcid")
	rootCmd.Flags().StringVar(&cfg.Gcess, "gcess", "", "极客时间 cookie 值 gcess")
	rootCmd.Flags().StringVarP(&cfg.DownloadFolder, "folder", "f", defaultDownloadFolder, "专栏和视频课的下载目标位置")
	rootCmd.Flags().StringVarP(&cfg.Quality, "quality", "q", "sd", "下载视频清晰度(ld标清,sd高清,hd超清)")
	rootCmd.Flags().IntVar(&cfg.DownloadComments, "comments", 1, "是否下载评论(0不下载,1下载首页评论,2下载所有评论)")
	rootCmd.Flags().IntVar(&cfg.ColumnOutputType, "output", 1, "专栏的输出内容(1pdf,2markdown,4audio)可自由组合")
	rootCmd.Flags().IntVar(&cfg.PrintPDFWaitSeconds, "print-pdf-wait", 5, "Chrome生成PDF前的等待页面加载时间, 单位为秒, 默认5秒")
	rootCmd.Flags().IntVar(&cfg.PrintPDFTimeoutSeconds, "print-pdf-timeout", 60, "Chrome生成PDF的超时时间, 单位为秒, 默认60秒")
	rootCmd.Flags().IntVar(&cfg.Interval, "interval", 1, "下载资源的间隔时间, 单位为秒, 默认1秒")
	rootCmd.Flags().BoolVar(&cfg.IsEnterprise, "enterprise", false, "是否下载企业版极客时间资源")
	rootCmd.Flags().StringVar(&cfg.LogLevel, "log-level", "info", "日志记录级别(debug, info, warn, error, none)")

	// 新增非交互式模式参数
	rootCmd.Flags().StringVar(&courseIDs, "course-ids", "", "课程ID列表，支持逗号分隔，例: 100056701,100056702")
	rootCmd.Flags().StringVar(&productType, "product-type", "", "产品类型: normal(普通课程), daily(每日一课), openclass(公开课), qconplus(大厂案例), university(训练营), other(其他)")
	rootCmd.Flags().BoolVar(&downloadAll, "download-all", true, "是否下载课程的所有内容")
	rootCmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "非交互模式标志（自动检测）")
	rootCmd.Flags().StringVar(&articleIDs, "article-ids", "", "指定下载的文章ID，支持逗号分隔，例: 1,2,3 （仅在download-all=false时生效）")
	rootCmd.Flags().StringVar(&configFile, "config", "", "配置文件路径，支持YAML格式的批量下载配置")
	rootCmd.Flags().BoolVar(&listCourses, "list", false, "列出已订阅的所有课程")

	rootCmd.MarkFlagsRequiredTogether("gcid", "gcess")
	rootCmd.MarkFlagsMutuallyExclusive("course-ids", "config")
}

var rootCmd = &cobra.Command{
	Use:          "geektime-downloader",
	Short:        "Geektime-downloader is used to download geek time lessons",
	SilenceUsage: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		logger.Init(cfg.LogLevel)
		return config.ValidateConfig(&cfg)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		readCookies := config.ReadCookiesFromInput(&cfg)
		geektimeClient = geektime.NewClient(readCookies)

		// 如果指定了--list，列出已订阅的课程
		if listCourses {
			return runListCourses(cmd.Context())
		}

		// 如果指定了配置文件，需要先读取配置文件来获取认证信息
		if configFile != "" {
			return runBatchDownloadFromConfig(cmd.Context(), configFile)
		}

		// 判断是否进入非交互式模式
		if shouldRunNonInteractive() {
			return runNonInteractiveMode(cmd.Context())
		}

		runner := fsm.NewFSMRunner(cmd.Context(), &cfg, geektimeClient)
		return runner.Run()
	},
}

// ================== 非交互式模式相关函数 ==================

// shouldRunNonInteractive 判断是否进入非交互式模式
func shouldRunNonInteractive() bool {
	if nonInteractive {
		return true
	}
	if courseIDs != "" {
		return true
	}
	return false
}

// runNonInteractiveMode 运行非交互式模式
func runNonInteractiveMode(ctx context.Context) error {
	fmt.Println("进入非交互式模式...")

	if courseIDs == "" {
		return fmt.Errorf("非交互模式下必须指定 --course-ids 或 --config 参数")
	}

	if productType == "" {
		return fmt.Errorf("非交互模式下必须指定 --product-type 参数")
	}

	courseIDList, err := parseCourseIDs(courseIDs)
	if err != nil {
		return fmt.Errorf("解析课程ID失败: %v", err)
	}

	productTypeOption, err := getProductTypeByString(productType)
	if err != nil {
		return fmt.Errorf("不支持的产品类型: %v", err)
	}

	var articleIDList []int
	if !downloadAll && articleIDs != "" {
		articleIDList, err = parseArticleIDs(articleIDs)
		if err != nil {
			return fmt.Errorf("解析文章ID失败: %v", err)
		}
	}

	concurrency := int(math.Ceil(float64(runtime.NumCPU()) / 2.0))
	downloader := course.NewCourseDownloader(ctx, &cfg, geektimeClient, nil)

	fmt.Printf("将下载 %d 个课程...\n", len(courseIDList))
	for i, courseID := range courseIDList {
		fmt.Printf("\n[%d/%d] 正在处理课程 ID: %d\n", i+1, len(courseIDList), courseID)

		courseConfig := CourseConfig{
			ID:          courseID,
			Type:        productType,
			DownloadAll: downloadAll,
			ArticleIDs:  articleIDList,
		}

		if err := downloadCourseByConfig(ctx, courseConfig, productTypeOption, downloader, concurrency); err != nil {
			fmt.Printf("课程 %d 下载失败: %v\n", courseID, err)
			continue
		}

		fmt.Printf("课程 %d 下载完成\n", courseID)

		if i < len(courseIDList)-1 {
			time.Sleep(time.Duration(cfg.Interval) * time.Second)
		}
	}

	fmt.Println("\n所有课程处理完成!")
	return nil
}

// runBatchDownloadFromConfig 从配置文件批量下载
func runBatchDownloadFromConfig(ctx context.Context, configPath string) error {
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	var batchConfig BatchConfig
	if err := yaml.Unmarshal(configData, &batchConfig); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 合并所有课程配置
	allCourses := append(batchConfig.Courses, batchConfig.AdvancedCourses...)

	fmt.Printf("从配置文件中读取到 %d 个课程配置...\n", len(allCourses))

	concurrency := int(math.Ceil(float64(runtime.NumCPU()) / 2.0))
	downloader := course.NewCourseDownloader(ctx, &cfg, geektimeClient, nil)

	for i, courseConfig := range allCourses {
		fmt.Printf("\n[%d/%d] 正在处理课程: ID=%d, Type=%s\n",
			i+1, len(allCourses), courseConfig.ID, courseConfig.Type)

		productTypeOption, err := getProductTypeByString(courseConfig.Type)
		if err != nil {
			fmt.Printf("课程 %d 不支持的产品类型 %s: %v\n", courseConfig.ID, courseConfig.Type, err)
			continue
		}

		// 应用全局配置 + 课程级别覆盖到当前 downloader cfg
		applyConfigForCourse(&cfg, batchConfig.Global, courseConfig, concurrency)

		// 重新创建 downloader 以使用新配置
		downloader = course.NewCourseDownloader(ctx, &cfg, geektimeClient, nil)

		if err := downloadCourseByConfig(ctx, courseConfig, productTypeOption, downloader, concurrency); err != nil {
			fmt.Printf("课程 %d 下载失败: %v\n", courseConfig.ID, err)
			continue
		}

		fmt.Printf("课程 %d 下载完成\n", courseConfig.ID)

		if i < len(allCourses)-1 {
			time.Sleep(time.Duration(cfg.Interval) * time.Second)
		}
	}

	fmt.Println("\n所有课程处理完成!")
	return nil
}

// downloadCourseByConfig 根据配置下载课程
func downloadCourseByConfig(ctx context.Context, courseConfig CourseConfig, productTypeOption ui.ProductTypeSelectOption, downloader *course.CourseDownloader, _ int) error {
	if !productTypeOption.NeedSelectArticle {
		// 处理每日一课、大厂案例等类型
		return downloadSingleProduct(ctx, courseConfig.ID, productTypeOption, downloader)
	}

	// 加载课程信息
	var c geektime.Course
	var err error

	if courseConfig.Type == "university" {
		c, err = geektimeClient.UniversityClassInfo(courseConfig.ID)
	} else if cfg.IsEnterprise {
		c, err = geektimeClient.EnterpriseCourseInfo(courseConfig.ID)
	} else {
		c, err = geektimeClient.CourseInfo(courseConfig.ID)
	}

	if err != nil {
		return fmt.Errorf("加载课程信息失败: %v", err)
	}

	if !c.Access {
		return fmt.Errorf("尚未购买该课程")
	}

	if courseConfig.DownloadAll {
		return downloader.DownloadAll(c, productTypeOption)
	} else {
		return downloadSpecificArticles(ctx, c, productTypeOption, downloader, courseConfig.ArticleIDs)
	}
}

// downloadSingleProduct 下载单个产品（每日一课、大厂案例等）
func downloadSingleProduct(ctx context.Context, productID int, productTypeOption ui.ProductTypeSelectOption, downloader *course.CourseDownloader) error {
	productInfo, err := geektimeClient.ProductInfo(productID)
	if err != nil {
		return fmt.Errorf("获取产品信息失败: %v", err)
	}

	if productInfo.Data.Info.Extra.Sub.AccessMask == 0 {
		return fmt.Errorf("尚未购买该课程")
	}

	return downloader.DownloadSingleVideoProduct(productInfo.Data.Info.Title,
		productInfo.Data.Info.Article.ID,
		productTypeOption.SourceType)
}

// downloadSpecificArticles 下载指定文章
func downloadSpecificArticles(ctx context.Context, c geektime.Course, productTypeOption ui.ProductTypeSelectOption, downloader *course.CourseDownloader, articleIDList []int) error {
	if len(articleIDList) == 0 {
		return fmt.Errorf("未指定要下载的文章ID")
	}

	articleMap := make(map[int]geektime.Article)
	for _, article := range c.Articles {
		articleMap[article.AID] = article
	}

	fmt.Printf("正在下载指定的 %d 个文章/视频...\n", len(articleIDList))

	for i, articleID := range articleIDList {
		article, exists := articleMap[articleID]
		if !exists {
			fmt.Printf("警告: 文章ID %d 不存在，跳过\n", articleID)
			continue
		}

		fmt.Printf("[%d/%d] 下载: %s\n", i+1, len(articleIDList), article.Title)

		if err := downloader.DownloadArticle(c, productTypeOption, article, false); err != nil {
			return err
		}

		if i < len(articleIDList)-1 {
			time.Sleep(time.Duration(cfg.Interval) * time.Second)
		}
	}

	return nil
}

// applyConfigForCourse 应用全局配置和课程级别配置到 AppConfig
func applyConfigForCourse(appCfg *config.AppConfig, global *GlobalConfig, courseConfig CourseConfig, defaultConcurrency int) {
	// 应用全局配置
	if global != nil {
		if global.Folder != "" {
			appCfg.DownloadFolder = global.Folder
		}
		if global.Quality != "" {
			appCfg.Quality = global.Quality
		}
		if global.Output != 0 {
			appCfg.ColumnOutputType = global.Output
		}
		if global.Comments != 0 {
			appCfg.DownloadComments = global.Comments
		}
		if global.Interval != 0 {
			appCfg.Interval = global.Interval
		}
		if global.PrintPDFWait != 0 {
			appCfg.PrintPDFWaitSeconds = global.PrintPDFWait
		}
		if global.PrintPDFTimeout != 0 {
			appCfg.PrintPDFTimeoutSeconds = global.PrintPDFTimeout
		}
		if global.Enterprise {
			appCfg.IsEnterprise = true
		}
	}

	// 应用课程级别覆盖
	if courseConfig.Quality != nil {
		appCfg.Quality = *courseConfig.Quality
	}
	if courseConfig.Output != nil {
		appCfg.ColumnOutputType = *courseConfig.Output
	}
	if courseConfig.Comments != nil {
		appCfg.DownloadComments = *courseConfig.Comments
	}
	if courseConfig.Folder != nil {
		appCfg.DownloadFolder = *courseConfig.Folder
	}
	if courseConfig.Interval != nil {
		appCfg.Interval = *courseConfig.Interval
	}
	if courseConfig.PrintPDFWait != nil {
		appCfg.PrintPDFWaitSeconds = *courseConfig.PrintPDFWait
	}
	if courseConfig.PrintPDFTimeout != nil {
		appCfg.PrintPDFTimeoutSeconds = *courseConfig.PrintPDFTimeout
	}
	if courseConfig.Enterprise != nil {
		appCfg.IsEnterprise = *courseConfig.Enterprise
	}
}

// parseCourseIDs 解析课程ID列表
func parseCourseIDs(courseIDsStr string) ([]int, error) {
	idStrs := strings.Split(strings.TrimSpace(courseIDsStr), ",")
	var ids []int

	for _, idStr := range idStrs {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid course ID: %s", idStr)
		}

		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("no valid course IDs provided")
	}

	return ids, nil
}

// parseArticleIDs 解析文章ID列表
func parseArticleIDs(articleIDsStr string) ([]int, error) {
	idStrs := strings.Split(strings.TrimSpace(articleIDsStr), ",")
	var ids []int

	for _, idStr := range idStrs {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid article ID: %s", idStr)
		}

		ids = append(ids, id)
	}

	return ids, nil
}

// getProductTypeByString 根据字符串获取产品类型
func getProductTypeByString(typeStr string) (ui.ProductTypeSelectOption, error) {
	typeMap := map[string]int{
		"normal":     0, // 普通课程
		"daily":      1, // 每日一课
		"openclass":  2, // 公开课
		"qconplus":   3, // 大厂案例
		"university": 4, // 训练营
		"other":      5, // 其他
	}

	index, exists := typeMap[typeStr]
	if !exists {
		return ui.ProductTypeSelectOption{}, fmt.Errorf("unsupported product type: %s. Supported types: normal, daily, openclass, qconplus, university, other", typeStr)
	}

	// 根据 index 构建 ProductTypeSelectOption
	options := buildProductTypeOptions(cfg.IsEnterprise)
	if index >= len(options) {
		return ui.ProductTypeSelectOption{}, fmt.Errorf("product type index out of range: %d", index)
	}

	return options[index], nil
}

func buildProductTypeOptions(isEnterprise bool) []ui.ProductTypeSelectOption {
	if isEnterprise {
		return []ui.ProductTypeSelectOption{
			{Index: 0, Text: "训练营", SourceType: 5, AcceptProductTypes: []string{"c44"}, NeedSelectArticle: true, IsEnterpriseMode: true},
		}
	}
	return []ui.ProductTypeSelectOption{
		{Index: 0, Text: "普通课程", SourceType: 1, AcceptProductTypes: []string{"c1", "c3"}, NeedSelectArticle: true, IsEnterpriseMode: false},
		{Index: 1, Text: "每日一课", SourceType: 2, AcceptProductTypes: []string{"d"}, NeedSelectArticle: false, IsEnterpriseMode: false},
		{Index: 2, Text: "公开课", SourceType: 1, AcceptProductTypes: []string{"p35", "p29", "p30"}, NeedSelectArticle: true, IsEnterpriseMode: false},
		{Index: 3, Text: "大厂案例", SourceType: 4, AcceptProductTypes: []string{"q"}, NeedSelectArticle: false, IsEnterpriseMode: false},
		{Index: 4, Text: "训练营", SourceType: 5, AcceptProductTypes: []string{""}, NeedSelectArticle: true, IsEnterpriseMode: false},
		{Index: 5, Text: "其他", SourceType: 1, AcceptProductTypes: []string{"x", "c6"}, NeedSelectArticle: true, IsEnterpriseMode: false},
	}
}

// runListCourses 列出已订阅的所有课程
func runListCourses(ctx context.Context) error {
	products, err := geektimeClient.MyProducts()
	if err != nil {
		return fmt.Errorf("获取课程列表失败: %v", err)
	}

	var columns []geektime.MyProduct
	var videos []geektime.MyProduct
	var others []geektime.MyProduct

	for _, p := range products {
		switch p.Type {
		case "c1":
			columns = append(columns, p)
		case "c3", "dls":
			videos = append(videos, p)
		default:
			others = append(others, p)
		}
	}

	sortMyProductsByID(columns)
	sortMyProductsByID(videos)
	sortMyProductsByID(others)

	fmt.Printf("\n%s\n", strings.Repeat("=", 100))
	fmt.Printf("%-100s (共 %d 个)\n", "专  栏", len(columns))
	fmt.Printf("%s\n", strings.Repeat("=", 100))
	fmt.Printf("%-14s %-6s %-10s %-12s %s\n", "课程ID", "类型", "完结", "进度", "课程标题")
	fmt.Printf("%s\n", strings.Repeat("-", 100))
	for _, c := range columns {
		fin := "否"
		if c.IsFinish {
			fin = "是"
		}
		unit := c.Unit
		if unit == "" {
			unit = "-"
		}
		fmt.Printf("%-14d %-6s %-10s %-12s %s\n", c.ID, c.Type, fin, unit, c.Title)
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 100))
	fmt.Printf("%-100s (共 %d 个)\n", "视频课", len(videos))
	fmt.Printf("%s\n", strings.Repeat("=", 100))
	fmt.Printf("%-14s %-6s %-10s %-12s %s\n", "课程ID", "类型", "完结", "进度", "课程标题")
	fmt.Printf("%s\n", strings.Repeat("-", 100))
	for _, v := range videos {
		fin := "否"
		if v.IsFinish {
			fin = "是"
		}
		unit := v.Unit
		if unit == "" {
			unit = "-"
		}
		fmt.Printf("%-14d %-6s %-10s %-12s %s\n", v.ID, v.Type, fin, unit, v.Title)
	}

	if len(others) > 0 {
		fmt.Printf("\n%s\n", strings.Repeat("=", 100))
		fmt.Printf("%-100s (共 %d 个)\n", "其他 (训练营/公开课/大会等)", len(others))
		fmt.Printf("%s\n", strings.Repeat("=", 100))
		fmt.Printf("%-14s %-6s %-10s %-6s %s\n", "课程ID", "类型", "完结", "进度", "课程标题")
		fmt.Printf("%s\n", strings.Repeat("-", 100))
		for _, o := range others {
			fin := "否"
			if o.IsFinish {
				fin = "是"
			}
			unit := o.Unit
			if unit == "" {
				unit = "-"
			}
			fmt.Printf("%-14d %-6s %-10s %-6s %s\n", o.ID, o.Type, fin, unit, o.Title)
		}
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 100))
	fmt.Printf("总计: 专栏 %d 个, 视频 %d 个, 其他 %d 个, 共 %d 个\n", len(columns), len(videos), len(others), len(products))
	fmt.Printf("%s\n", strings.Repeat("=", 100))

	return nil
}

// sortMyProductsByID 按 ID 升序排序
func sortMyProductsByID(products []geektime.MyProduct) {
	for i := 0; i < len(products); i++ {
		for j := i + 1; j < len(products); j++ {
			if products[i].ID > products[j].ID {
				products[i], products[j] = products[j], products[i]
			}
		}
	}
}

// Sometime video exist in article content, see issue #104
// <p>
// <video poster="https://static001.geekbang.org/resource/image/6a/f7/6ada085b44eddf37506b25ad188541f7.jpg" preload="none" controls="">
// <source src="https://media001.geekbang.org/customerTrans/fe4a99b62946f2c31c2095c167b26f9c/30d99c0d-16d14089303-0000-0000-01d-dbacd.mp4" type="video/mp4">
// <source src="https://media001.geekbang.org/2ce11b32e3e740ff9580185d8c972303/a01ad13390fe4afe8856df5fb5d284a2-f2f547049c69fa0d4502ab36d42ea2fa-sd.m3u8" type="application/x-mpegURL">
// <source src="https://media001.geekbang.org/2ce11b32e3e740ff9580185d8c972303/a01ad13390fe4afe8856df5fb5d284a2-2528b0077e78173fd8892de4d7b8c96d-hd.m3u8" type="application/x-mpegURL"></video>
// </p>
func getVideoURLFromArticleContent(content string) (hasVideo bool, videoURL string) {
	if !strings.Contains(content, "<video") || !strings.Contains(content, "<source") {
		return false, ""
	}
	doc, err := html.Parse(strings.NewReader(content))
	if err != nil {
		return false, ""
	}
	hasVideo, videoURL = false, ""
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "video" {
			hasVideo = true
		}
		if n.Type == html.ElementNode && n.Data == "source" {
			for _, a := range n.Attr {
				if a.Key == "src" && hasVideo && strings.HasSuffix(a.Val, ".mp4") {
					videoURL = a.Val
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return hasVideo, videoURL
}

// Execute ...
func Execute() {
	ctx := context.Background()

	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
