package main

import (
	"fmt"
	"github.com/sagoo-cloud/nexframe/i18n"
)

func main() {
	// 初始化全局翻译实例
	err := i18n.InitGlobal("zh")
	if err != nil {
		fmt.Printf("初始化失败: %v\n", err)
		return
	}

	fmt.Println(i18n.T("hello"))                            // 输出：你好
	fmt.Println(i18n.T("world"))                            // 输出：世界
	fmt.Println(i18n.FormatTranslation("welcome", "Alice")) // 输出：欢迎, Alice!

	// 切换语言
	err = i18n.SetLang("en")
	if err != nil {
		fmt.Printf("切换语言失败: %v\n", err)
		return
	}

	fmt.Println(i18n.T("hello")) // 输出：Hello
	fmt.Println(i18n.T("world")) // 输出：World

	// 添加新的翻译
	i18n.AddTranslation("goodbye", "再见")
	fmt.Println(i18n.T("goodbye")) // 输出：再见
}
