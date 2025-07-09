package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("节点管理测速系统 - 面板服务")
	fmt.Println("版本: v1.0.0")
	fmt.Println("这是一个测试版本，用于验证GitHub Actions工作流")
	
	// 如果提供了参数，显示参数
	if len(os.Args) > 1 {
		fmt.Println("参数:", os.Args[1:])
	}
} 