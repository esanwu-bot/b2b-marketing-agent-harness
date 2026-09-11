package policy

import (
	"testing"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
)

func TestExternalRequiresApproval(t *testing.T) {
	e := New(false)
	d := e.Evaluate(domain.Action{Tool: "linkedin.create_post", Risk: domain.RiskExternal})
	if d.Outcome != domain.OutcomeRequireApproval {
		t.Fatalf("外部动作应需审批，得到 %s", d.Outcome)
	}
}

func TestExternalAutoApprove(t *testing.T) {
	e := New(true)
	d := e.Evaluate(domain.Action{Tool: "linkedin.create_post", Risk: domain.RiskExternal})
	if d.Outcome != domain.OutcomeAllow {
		t.Fatalf("自动审批模式下应放行，得到 %s", d.Outcome)
	}
}

func TestReadAllowedByDefault(t *testing.T) {
	e := New(false)
	d := e.Evaluate(domain.Action{Tool: "crawler.search", Risk: domain.RiskRead})
	if d.Outcome != domain.OutcomeAllow {
		t.Fatalf("只读动作应放行，得到 %s", d.Outcome)
	}
}

func TestExplicitDenyRule(t *testing.T) {
	e := New(false)
	e.SetRule("content.delete", domain.OutcomeDeny)
	d := e.Evaluate(domain.Action{Tool: "content.delete", Risk: domain.RiskWrite})
	if d.Outcome != domain.OutcomeDeny {
		t.Fatalf("显式拒绝规则应生效，得到 %s", d.Outcome)
	}
}
