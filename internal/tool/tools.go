package tool

import (
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/llm"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/store"
)

// DefaultTools 组装内置的全部 Native Tool。
func DefaultTools(st store.Store, p llm.Provider) []domain.Tool {
	var out []domain.Tool
	out = append(out, NewCrawlerTools()...)
	out = append(out, NewContentTools(p)...)
	out = append(out, NewKnowledgeTools(st)...)
	out = append(out, NewLinkedInTools(st)...)
	out = append(out, NewNotificationTools()...)
	out = append(out, NewBrowserTools()...)
	out = append(out, NewMarketingTools(st, p)...)
	out = append(out, NewPipelineTools(st)...)
	return out
}

// BuildRegistry 构建包含全部内置 Tool 的注册表。
func BuildRegistry(st store.Store, p llm.Provider) *Registry {
	r := NewRegistry()
	r.RegisterAll(DefaultTools(st, p))
	return r
}
