package main

import (
	"fmt"
	"log"

	"InfoHub-agent/internal/ai"
	"InfoHub-agent/internal/crawler"
	"InfoHub-agent/internal/delivery"
	"InfoHub-agent/internal/service"
)

func main() {
	// 初始化 MVP 主流程组件，并输出 Markdown 日报。
	pipeline := service.NewPipeline(
		crawler.NewDemoCrawler(),
		ai.NewMockProcessor(),
	)

	items, err := pipeline.Run()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(delivery.RenderMarkdown(items))
}
