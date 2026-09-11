// crawler harness scheduler：按周期触发 Workflow（V1 使用简单 ticker，M6 接入 cron/时区/幂等）。
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/app"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/config"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "配置文件路径")
	workflow := flag.String("workflow", "daily_semiconductor_intelligence", "要调度的 Workflow key")
	every := flag.Duration("every", 0, "触发周期（如 24h / 1h / 10m）；0 表示不自动触发")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	a, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("装配 Harness 失败: %v", err)
	}

	if *every <= 0 {
		log.Printf("Scheduler 空闲：使用 -every 24h 启用周期触发（workflow=%s）", *workflow)
		<-ctx.Done()
		return
	}

	log.Printf("Scheduler 已启动：每 %s 触发 %s", *every, *workflow)
	ticker := time.NewTicker(*every)
	defer ticker.Stop()

	runOnce := func() {
		run, err := a.RunWorkflow(ctx, *workflow, "cron")
		if err != nil {
			log.Printf("触发失败: %v", err)
			return
		}
		log.Printf("触发成功：workflow_run=%s status=%s", run.ID, run.Status)
	}

	runOnce()
	for {
		select {
		case <-ctx.Done():
			log.Printf("Scheduler 退出")
			return
		case <-ticker.C:
			runOnce()
		}
	}
}
