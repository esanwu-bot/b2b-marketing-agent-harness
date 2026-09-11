// crawler harness API：REST + SSE 服务，供 Workbench 控制台消费。
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/app"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/config"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/transport/httpapi"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "配置文件路径")
	addr := flag.String("addr", "", "监听地址（覆盖配置）")
	flag.Parse()

	ctx := context.Background()
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	if *addr != "" {
		cfg.Server.Addr = *addr
	}

	a, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("装配 Harness 失败: %v", err)
	}

	srv := &http.Server{
		Addr:              cfg.Server.Addr,
		Handler:           httpapi.New(a).Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("B2B Marketing Agent Harness API 启动于 http://%s/api/v1/health", cfg.Server.Addr)
	log.Printf("store=%s llm=%s auto_approve=%v workspace=%s", a.Store.Name(), a.LLM.Name(), a.Policy.AutoApprove(), cfg.Workspace.Name)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("API 服务异常: %v", err)
	}
}
