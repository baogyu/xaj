package main

import (
	"fmt"

	"demo2/Calibration" // 使用模块路径而不是相对路径
)

func main() {
	// 创建 SCEUA 优化器实例
	sceua := Calibration.NewSCEUA()

	// 设置工作目录路径
	sceua.SetFilePath("/Users/baogy/goProject/owner/demo2/datas/") // 替换为你的实际工作目录路径

	// 可以根据需要设置其他参数
	// 例如：设置参数名称、初始值、上下限等

	fmt.Println("开始SCE-UA优化...")

	// 执行优化
	sceua.Optimize()

	fmt.Println("优化完成!")
}
