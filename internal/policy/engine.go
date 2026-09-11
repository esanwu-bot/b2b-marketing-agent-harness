// Package policy 是策略引擎：决定某个动作是「直接执行 / 需人工审批 / 拒绝」。
package policy

import (
	"sync"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
)

// Engine 按规则裁决动作。
type Engine struct {
	mu          sync.RWMutex
	rules       map[string]domain.DecisionOutcome // key: 精确工具名
	autoApprove bool
}

// New 创建策略引擎。autoApprove=true 时，本应审批的动作自动通过（演示/低风险环境）。
func New(autoApprove bool) *Engine {
	return &Engine{rules: map[string]domain.DecisionOutcome{}, autoApprove: autoApprove}
}

// SetRule 设置某工具的裁决规则。
func (e *Engine) SetRule(tool string, out domain.DecisionOutcome) {
	e.mu.Lock()
	e.rules[tool] = out
	e.mu.Unlock()
}

// AutoApprove 返回是否自动审批。
func (e *Engine) AutoApprove() bool { return e.autoApprove }

// Evaluate 裁决一个动作。
func (e *Engine) Evaluate(a domain.Action) domain.PolicyDecision {
	e.mu.RLock()
	rule, hasRule := e.rules[a.Tool]
	e.mu.RUnlock()

	// 1) 显式规则优先
	if hasRule {
		return e.decideFromOutcome(rule, "rule:"+a.Tool, a)
	}
	// 2) 默认按风险等级
	switch a.Risk {
	case domain.RiskExternal:
		return e.decideFromOutcome(domain.OutcomeRequireApproval, "risk:external", a)
	case domain.RiskWrite:
		return e.decideFromOutcome(domain.OutcomeAllow, "risk:write", a)
	default:
		return e.decideFromOutcome(domain.OutcomeAllow, "risk:read", a)
	}
}

func (e *Engine) decideFromOutcome(out domain.DecisionOutcome, rule string, a domain.Action) domain.PolicyDecision {
	reason := "策略判定"
	if out == domain.OutcomeRequireApproval && e.autoApprove {
		return domain.PolicyDecision{Outcome: domain.OutcomeAllow, Rule: rule + "+auto_approve", Reason: "自动审批已开启", Risk: a.Risk}
	}
	switch out {
	case domain.OutcomeRequireApproval:
		reason = "该动作风险较高，需人工审批"
	case domain.OutcomeDeny:
		reason = "该动作被策略拒绝"
	default:
		reason = "允许执行"
	}
	return domain.PolicyDecision{Outcome: out, Rule: rule, Reason: reason, Risk: a.Risk}
}
