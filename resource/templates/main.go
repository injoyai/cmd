package main

import "fmt"

// BuildDate 编译时间,由 build.sh 注入
// -ldflags="-s -w -X main.BuildDate=$(date +%Y-%m-%d)"
var BuildDate = "unknown"

func main() {
	fmt.Printf("starting, build: %s\n", BuildDate)
}
