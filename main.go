package main

import (
	"fmt"

	"demo2/Calibration" // 使用模块路径而不是相对路径
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("程序发生错误: %v\n", r)
		}
	}()

	sceua := Calibration.NewSCEUA()

	workPath := "/Users/baogy/goProject/owner/demo2/datas/"
	fmt.Printf("设置工作目录: %s\n", workPath)
	sceua.SetFilePath(workPath)

	fmt.Println("开始SCE-UA优化...")
	sceua.Optimize()
	fmt.Println("优化完成!")
}
