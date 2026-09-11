// crawler harness worker：消费 Task 并驱动 Agent（V1 提供基础循环，队列将在 M2/M3 接入 Redis Streams）。
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
	workflow := flag.String("workflow", "", "启动后立即执行一次的 Workflow key")
	once := flag.Bool("once", false, "执行一次后退出")
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

	if *workflow != "" {
		run, err := a.RunWorkflow(ctx, *workflow, "worker")
		if err != nil {
			log.Fatalf("执行 workflow 失败: %v", err)
		}
		log.Printf("workflow=%s status=%s", run.WorkflowKey, run.Status)
		if *once {
			return
		}
	}

	log.Printf("Worker 已启动（store=%s）。等待任务中…（Ctrl+C 退出）", a.Store.Name())
	<-ctx.Done()
	log.Printf("Worker 退出")
	time.Sleep(50 * time.Millisecond)
}
