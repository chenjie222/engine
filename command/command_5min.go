package command

import (
	"fmt"
	"gitee.com/quant1x/engine/tools"
	cmder "github.com/spf13/cobra"
)

const (
	min5Command     = "5min"
	min5Description = "下载5分钟K线数据"
)

var (
	// Cmd5Min 5分钟K线下载命令
	Cmd5Min *cmder.Command = nil
	
	// flag5MinCode 指定单个股票代码
	flag5MinCode = cmdFlag[string]{
		Name:  "code",
		Usage: "指定股票代码, 如: sz300773",
		Value: "",
	}
	
	// flag5MinSector 指定板块
	flag5MinSector = cmdFlag[string]{
		Name:  "sector",
		Usage: "指定板块, 如: sh600, sz300, sh688 等",
		Value: "",
	}
	
	// flag5MinAll 下载所有股票
	flag5MinAll = cmdFlag[bool]{
		Name:  "all",
		Usage: "下载所有股票的5分钟K线",
		Value: false,
	}
	
	// flag5MinPath 显示存储路径
	flag5MinPath = cmdFlag[bool]{
		Name:  "path",
		Usage: "显示5分钟K线数据存储路径",
		Value: false,
	}
	
	// flag5MinStart 开始日期
	flag5MinStart = cmdFlag[string]{
		Name:  "start",
		Usage: "开始日期, 格式YYYYMMDD, 如: --start=20230101 (默认20200101)",
		Value: "",
	}
	
	// flag5MinUpdate 增量更新模式
	flag5MinUpdate = cmdFlag[bool]{
		Name:  "update",
		Usage: "增量更新模式, 只下载新增数据",
		Value: false,
	}
	
	// flag5MinStats 显示统计信息
	flag5MinStats = cmdFlag[bool]{
		Name:  "stats",
		Usage: "显示指定股票的5分钟K线统计信息",
		Value: false,
	}
)

// InitCmd5Min 初始化5分钟K线命令
// 在 initSubCommands 中被调用
func InitCmd5Min() {
	Cmd5Min = &cmder.Command{
		Use:     min5Command,
		Example: Application + " " + min5Command + " --all",
		Short:   min5Description,
		Long:    "下载通达信5分钟K线数据, 支持单个股票、板块或全部股票\n\n使用示例:\n  engine 5min --code=sz300773                    # 下载单个股票(默认从2020-01-01)\n  engine 5min --code=sz300773 --start=20230101   # 下载指定日期\n  engine 5min --code=sz300773 --update           # 增量更新\n  engine 5min --sector=sh600 --start=20240101    # 下载指定板块\n  engine 5min --all                              # 下载所有股票\n  engine 5min --all --update                     # 增量更新全部\n  engine 5min --path                             # 显示存储路径\n  engine 5min --code=sz300773 --stats            # 查看统计信息",
		Run: func(cmd *cmder.Command, args []string) {
			handle5MinDownload()
		},
	}
	
	// 注册命令参数
	commandInit(Cmd5Min, &flag5MinCode)
	commandInit(Cmd5Min, &flag5MinSector)
	commandInit(Cmd5Min, &flag5MinAll)
	commandInit(Cmd5Min, &flag5MinPath)
	commandInit(Cmd5Min, &flag5MinStart)
	commandInit(Cmd5Min, &flag5MinUpdate)
	commandInit(Cmd5Min, &flag5MinStats)
}

func handle5MinDownload() {
	// 1. 显示路径
	if flag5MinPath.Value {
		tools.Show5MinKLinePath()
		return
	}
	
	// 2. 显示统计信息
	if flag5MinStats.Value {
		if len(flag5MinCode.Value) > 0 {
			count, first, last := tools.Get5MinKLineStats(flag5MinCode.Value)
			if count > 0 {
				fmt.Printf("股票 %s 5分钟K线统计:\n", flag5MinCode.Value)
				fmt.Printf("  记录数: %d 条\n", count)
				fmt.Printf("  起始时间: %s\n", first)
				fmt.Printf("  结束时间: %s\n", last)
			} else {
				fmt.Printf("股票 %s 暂无5分钟K线数据\n", flag5MinCode.Value)
			}
		} else {
			fmt.Println("错误: --stats 参数需要配合 --code 使用")
			fmt.Println("  示例: ./engine 5min --code=sz300773 --stats")
		}
		return
	}
	
	// 验证日期格式
	if len(flag5MinStart.Value) > 0 {
		if !tools.ValidateDateFormat(flag5MinStart.Value) {
			fmt.Printf("错误: 日期格式不正确: %s\n", flag5MinStart.Value)
			fmt.Println("  正确格式: YYYYMMDD, 如: 20230101")
			return
		}
	}
	
	// 3. 下载单个股票
	if len(flag5MinCode.Value) > 0 {
		tools.Download5MinKLineForCode(flag5MinCode.Value, flag5MinStart.Value, flag5MinUpdate.Value, nil)
		return
	}
	
	// 4. 下载指定板块
	if len(flag5MinSector.Value) > 0 {
		tools.Download5MinKLineForSector(flag5MinSector.Value, flag5MinStart.Value, flag5MinUpdate.Value)
		return
	}
	
	// 5. 下载所有股票
	if flag5MinAll.Value {
		tools.Download5MinKLineForAll(flag5MinStart.Value, flag5MinUpdate.Value)
		return
	}
	
	// 6. 没有指定任何参数, 显示帮助
	fmt.Println("错误: 需要指定以下参数之一:")
	fmt.Println("  --code=sz300773          下载指定股票")
	fmt.Println("  --sector=sh600           下载指定板块")
	fmt.Println("  --all                    下载所有股票")
	fmt.Println("  --path                   显示存储路径")
	fmt.Println()
	fmt.Println("可选参数:")
	fmt.Println("  --start=20230101         设置开始日期(默认20200101)")
	fmt.Println("  --update                 增量更新模式")
	fmt.Println("  --stats                  显示统计信息(需配合--code)")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  ./engine 5min --code=sz300773 ")
	fmt.Println("  ./engine 5min --code=sz300773 --start=20230101")
	fmt.Println("  ./engine 5min --code=sz300773 --update")
	fmt.Println("  ./engine 5min --sector=sh600 --start=20240101")
	fmt.Println("  ./engine 5min --all --update")
	fmt.Println()
	_ = Cmd5Min.Usage()
}
