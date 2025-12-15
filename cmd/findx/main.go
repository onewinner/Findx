package main

import (
	"fmt"
	"log"
	"os"

	"Findx/internal/config"
	"Findx/internal/scanner"

	"github.com/urfave/cli/v2"
)

func main() {
	// 获取应用信息
	appName, appUsage, appVersion := config.GetAppInfo()

	app := &cli.App{
		Name:                 appName,
		Usage:                appUsage,
		Version:              appVersion,
		Description:          config.GetDescription(),
		UsageText:            config.GetUsageText(),
		EnableBashCompletion: true,
		Flags:                config.GetFlags(),
		Before: func(c *cli.Context) error {
			// 打印Banner
			config.PrintBanner()
			return nil
		},
		Action: func(c *cli.Context) error {
			// 解析配置
			cfg, err := config.ParseConfig(c)
			if err != nil {
				return fmt.Errorf("解析配置失败: %w", err)
			}

			// 验证配置
			if err := cfg.Validate(); err != nil {
				return fmt.Errorf("配置验证失败: %w", err)
			}

			// 打印配置信息
			cfg.PrintConfig()

			// 创建并运行扫描器
			s := scanner.NewScanner(cfg)
			if err := s.Run(); err != nil {
				return fmt.Errorf("扫描失败: %w", err)
			}

			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "examples",
				Aliases: []string{"ex"},
				Usage:   "显示使用示例 / Show usage examples",
				Action: func(c *cli.Context) error {
					fmt.Println(config.GetExamples())
					return nil
				},
			},
		},
	}

	// 自定义帮助模板
	cli.AppHelpTemplate = `名称 / NAME:
   {{.Name}} - {{.Usage}}

用法 / USAGE:
   {{.UsageText}}

版本 / VERSION:
   {{.Version}}

全局选项 / GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
命令 / COMMANDS:
   {{range .Commands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
   {{end}}
运行 'findx examples' 查看更多示例 / Run 'findx examples' for more examples
`

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
