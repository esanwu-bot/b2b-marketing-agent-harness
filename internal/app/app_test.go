package app_test

import (
	"context"
	"testing"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/app"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/config"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
)

// TestReferencePipeline 验证「半导体营销参考流水线」端到端跑通。
func TestReferencePipeline(t *testing.T) {
	ctx := context.Background()
	a, err := app.New(ctx, config.Default())
	if err != nil {
		t.Fatalf("装配失败: %v", err)
	}

	run, err := a.RunWorkflow(ctx, "daily_semiconductor_intelligence", "test")
	if err != nil {
		t.Fatalf("执行 workflow 失败: %v", err)
	}
	if run.Status != domain.WfSucceeded {
		t.Fatalf("期望 workflow succeeded，得到 %s (error=%s)", run.Status, run.Error)
	}
	if len(run.Steps) != 8 {
		t.Fatalf("期望 8 个步骤，得到 %d", len(run.Steps))
	}

	docs, _ := a.Store.CountDocuments(ctx)
	if docs == 0 {
		t.Fatalf("研究阶段未写入知识库")
	}
	topics, _ := a.Store.ListTopics(ctx, 10)
	if len(topics) == 0 {
		t.Fatalf("趋势阶段未产出选题")
	}
	content, _ := a.Store.ListContent(ctx, 10)
	if len(content) == 0 {
		t.Fatalf("内容阶段未产出草稿")
	}
	pubs, _ := a.Store.ListPublications(ctx, 10)
	if len(pubs) == 0 {
		t.Fatalf("发布阶段未产生 publication")
	}
	runs, _ := a.Store.ListRuns(ctx, 20)
	if len(runs) < 6 {
		t.Fatalf("期望至少 6 个 agent run，得到 %d", len(runs))
	}
}
