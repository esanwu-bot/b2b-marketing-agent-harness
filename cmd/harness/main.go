// crawler harness demo：离线跑通半导体营销参考流水线并打印报告。
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/app"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/config"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "配置文件路径")
	flag.Parse()

	args := flag.Args()
	cmd := "demo"
	if len(args) > 0 {
		cmd = args[0]
	}

	ctx := context.Background()
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	a, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("装配 Harness 失败: %v", err)
	}

	switch cmd {
	case "demo":
		if err := runDemo(ctx, a); err != nil {
			log.Fatalf("demo 失败: %v", err)
		}
	default:
		fmt.Printf("未知命令 %q，可用：demo\n", cmd)
		os.Exit(1)
	}
}

func runDemo(ctx context.Context, a *app.App) error {
	const wfKey = "daily_semiconductor_intelligence"

	fmt.Println("============================================================")
	fmt.Println("  B2B Marketing Agent Harness · 离线演示")
	fmt.Printf("  workspace=%s  store=%s  llm=%s  auto_approve=%v\n",
		a.Cfg.Workspace.Name, a.Store.Name(), a.LLM.Name(), a.Policy.AutoApprove())
	fmt.Println("============================================================")

	fmt.Println("\n[1] 执行 Workflow:", wfKey)
	run, err := a.RunWorkflow(ctx, wfKey, "demo")
	if err != nil {
		return err
	}
	fmt.Printf("    workflow_run=%s status=%s\n", run.ID, run.Status)
	for _, st := range run.Steps {
		fmt.Printf("    - step %d [%s] %s => %s\n", st.Seq, st.Kind, st.Ref, st.Status)
	}

	fmt.Println("\n[2] Agent Runs")
	runs, _ := a.Store.ListRuns(ctx, 20)
	printed := map[string]bool{}
	for i := len(runs) - 1; i >= 0; i-- {
		r := runs[i]
		if printed[r.ID] {
			continue
		}
		printed[r.ID] = true
		fmt.Printf("    - %s agent=%-10s status=%-9s steps=%d tokens=%d\n", r.ID, r.AgentID, r.Status, len(r.Steps), r.PromptTokens)
	}

	fmt.Println("\n[3] 知识库 / 选题 / 内容 / 发布")
	docs, _ := a.Store.CountDocuments(ctx)
	topics, _ := a.Store.ListTopics(ctx, 10)
	content, _ := a.Store.ListContent(ctx, 10)
	pubs, _ := a.Store.ListPublications(ctx, 10)
	approvals, _ := a.Store.ListApprovals(ctx, "", 10)
	events, _ := a.Store.ListEvents(ctx, 10)
	fmt.Printf("    文档=%d  选题=%d  内容=%d  发布=%d  审批=%d  事件=%d\n", docs, len(topics), len(content), len(pubs), len(approvals), len(events))
	for _, t := range topics {
		fmt.Printf("    选题 [%s] %s (score=%.0f)\n", t.Category, t.Title, t.Score)
	}
	for _, c := range content {
		fmt.Printf("    内容 [%s] %s\n", c.Status, c.Title)
		fmt.Printf("         %s\n", firstLine(c.Body))
	}
	for _, p := range pubs {
		fmt.Printf("    发布 [%s] %s %s\n", p.Status, p.ExternalID, p.ExternalURL)
	}

	fmt.Println("\n[4] 指标")
	m, _ := a.Store.ListMetrics(ctx, "", 10)
	_ = m
	fmt.Println("    内容发布指标已写入（content_metrics / publications）")

	fmt.Println("\n[DONE] 离线流水线执行完成。")
	return nil
}

func firstLine(s string) string {
	for i, r := range s {
		if r == '\n' {
			return s[:i]
		}
	}
	if len(s) > 80 {
		return s[:80]
	}
	return s
}
