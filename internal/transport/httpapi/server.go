// Package httpapi 提供 Harness 的 REST + SSE 接口（供 Workbench 控制台消费）。
package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/app"
	"github.com/esanwu-bot/b2b-marketing-agent-harness/internal/domain"
)

// Server HTTP 服务。
type Server struct {
	app   *app.App
	start time.Time
}

// New 创建服务。
func New(a *app.App) *Server {
	return &Server{app: a, start: time.Now()}
}

// Handler 构建路由。
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/health", s.health)
	mux.HandleFunc("GET /api/v1/metrics", s.metrics)

	mux.HandleFunc("GET /api/v1/agents", s.listAgents)
	mux.HandleFunc("POST /api/v1/agents/{id}/run", s.runAgent)
	mux.HandleFunc("GET /api/v1/tools", s.listTools)

	mux.HandleFunc("GET /api/v1/workflows", s.listWorkflows)
	mux.HandleFunc("POST /api/v1/workflows/{key}/run", s.runWorkflow)
	mux.HandleFunc("GET /api/v1/workflow-runs", s.listWorkflowRuns)
	mux.HandleFunc("GET /api/v1/workflow-runs/{id}", s.getWorkflowRun)

	mux.HandleFunc("GET /api/v1/runs", s.listRuns)
	mux.HandleFunc("GET /api/v1/runs/{id}", s.getRun)
	mux.HandleFunc("GET /api/v1/runs/{id}/stream", s.streamRun)

	mux.HandleFunc("GET /api/v1/tasks", s.listTasks)

	mux.HandleFunc("GET /api/v1/approvals", s.listApprovals)
	mux.HandleFunc("POST /api/v1/approvals/{id}/approve", s.approve)
	mux.HandleFunc("POST /api/v1/approvals/{id}/reject", s.reject)

	mux.HandleFunc("GET /api/v1/knowledge", s.listKnowledge)
	mux.HandleFunc("GET /api/v1/knowledge/search", s.searchKnowledge)
	mux.HandleFunc("GET /api/v1/topics", s.listTopics)
	mux.HandleFunc("GET /api/v1/content", s.listContent)
	mux.HandleFunc("GET /api/v1/publications", s.listPublications)
	mux.HandleFunc("GET /api/v1/events", s.listEvents)

	return withCORS(mux)
}

// ---------------- 基础 ----------------

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	if err := s.app.Store.Ping(r.Context()); err != nil {
		status = "degraded"
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":       status,
		"uptime_s":     int(time.Since(s.start).Seconds()),
		"store":        s.app.Store.Name(),
		"llm":          s.app.LLM.Name(),
		"workspace":    s.app.Cfg.Workspace.Name,
		"auto_approve": s.app.Policy.AutoApprove(),
		"time":         time.Now().Format(time.RFC3339),
	})
}

// ---------------- Agent / Tool ----------------

func (s *Server) listAgents(w http.ResponseWriter, r *http.Request) {
	specs, err := s.app.Store.ListAgents(r.Context())
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"agents": specs})
}

func (s *Server) runAgent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ag, ok := s.app.Agents.Get(id)
	if !ok {
		writeErr(w, 404, "agent 不存在: "+id)
		return
	}
	var req struct {
		Goal  string         `json:"goal"`
		Input map[string]any `json:"input"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	task := &domain.Task{AgentID: id, Type: "manual", Goal: req.Goal, Input: req.Input, Status: domain.TaskRunning}
	if err := s.app.Store.CreateTask(r.Context(), task); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	res, err := ag.Run(r.Context(), task)
	if err != nil {
		writeJSON(w, 200, map[string]any{"result": res, "error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"result": res})
}

func (s *Server) listTools(w http.ResponseWriter, r *http.Request) {
	infos, err := s.app.Store.ListTools(r.Context())
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"tools": infos})
}

// ---------------- Workflow ----------------

func (s *Server) listWorkflows(w http.ResponseWriter, r *http.Request) {
	defs, err := s.app.Store.ListWorkflows(r.Context())
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"workflows": defs})
}

func (s *Server) runWorkflow(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	run, err := s.app.RunWorkflow(r.Context(), key, "manual")
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"workflow_run": run})
}

func (s *Server) listWorkflowRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := s.app.Store.ListWorkflowRuns(r.Context(), queryInt(r, "limit", 20))
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"workflow_runs": runs})
}

func (s *Server) getWorkflowRun(w http.ResponseWriter, r *http.Request) {
	run, err := s.app.Store.GetWorkflowRun(r.Context(), r.PathValue("id"))
	if err != nil || run == nil {
		writeErr(w, 404, "workflow run 不存在")
		return
	}
	writeJSON(w, 200, map[string]any{"workflow_run": run})
}

// ---------------- Run / Task ----------------

func (s *Server) listRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := s.app.Store.ListRuns(r.Context(), queryInt(r, "limit", 30))
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"runs": runs})
}

func (s *Server) getRun(w http.ResponseWriter, r *http.Request) {
	run, err := s.app.Store.GetRun(r.Context(), r.PathValue("id"))
	if err != nil || run == nil {
		writeErr(w, 404, "run 不存在")
		return
	}
	writeJSON(w, 200, map[string]any{"run": run})
}

func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.app.Store.ListTasks(r.Context(), queryInt(r, "limit", 30))
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"tasks": tasks})
}

// ---------------- Approval ----------------

func (s *Server) listApprovals(w http.ResponseWriter, r *http.Request) {
	list, err := s.app.Store.ListApprovals(r.Context(), r.URL.Query().Get("status"), queryInt(r, "limit", 50))
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"approvals": list})
}

func (s *Server) approve(w http.ResponseWriter, r *http.Request) { s.decide(w, r, "approved") }
func (s *Server) reject(w http.ResponseWriter, r *http.Request)  { s.decide(w, r, "rejected") }

func (s *Server) decide(w http.ResponseWriter, r *http.Request, status string) {
	id := r.PathValue("id")
	apr, err := s.app.Store.GetApproval(r.Context(), id)
	if err != nil || apr == nil {
		writeErr(w, 404, "审批不存在")
		return
	}
	now := time.Now()
	apr.Status = status
	apr.DecidedBy = "admin"
	apr.DecidedAt = &now
	if err := s.app.Store.UpdateApproval(r.Context(), apr); err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	_ = s.app.Store.AppendEvent(r.Context(), domain.Event{Type: "approval." + status, AggregateType: "approval", AggregateID: id})
	// 审批完成后，恢复（通过）或取消（拒绝）关联的工作流运行。
	if apr.WorkflowRunID != "" {
		if status == "approved" {
			_, _ = s.app.ResumeWorkflow(r.Context(), apr.WorkflowRunID)
		} else {
			_, _ = s.app.CancelWorkflow(r.Context(), apr.WorkflowRunID, "审批被拒绝")
		}
	}
	writeJSON(w, 200, map[string]any{"approval": apr})
}

// ---------------- Knowledge / Content ----------------

func (s *Server) listKnowledge(w http.ResponseWriter, r *http.Request) {
	docs, err := s.app.Store.ListDocuments(r.Context(), queryInt(r, "limit", 50))
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	count, _ := s.app.Store.CountDocuments(r.Context())
	writeJSON(w, 200, map[string]any{"documents": docs, "total": count})
}

func (s *Server) searchKnowledge(w http.ResponseWriter, r *http.Request) {
	hits, err := s.app.Store.SearchDocuments(r.Context(), r.URL.Query().Get("q"), queryInt(r, "limit", 20))
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"hits": hits})
}

func (s *Server) listTopics(w http.ResponseWriter, r *http.Request) {
	topics, err := s.app.Store.ListTopics(r.Context(), queryInt(r, "limit", 50))
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"topics": topics})
}

func (s *Server) listContent(w http.ResponseWriter, r *http.Request) {
	list, err := s.app.Store.ListContent(r.Context(), queryInt(r, "limit", 50))
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"content": list})
}

func (s *Server) listPublications(w http.ResponseWriter, r *http.Request) {
	list, err := s.app.Store.ListPublications(r.Context(), queryInt(r, "limit", 50))
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"publications": list})
}

func (s *Server) listEvents(w http.ResponseWriter, r *http.Request) {
	list, err := s.app.Store.ListEvents(r.Context(), queryInt(r, "limit", 50))
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"events": list})
}

// ---------------- 汇总指标（Dashboard） ----------------

func (s *Server) metrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	runs, _ := s.app.Store.ListRuns(ctx, 500)
	tasks, _ := s.app.Store.ListTasks(ctx, 500)
	content, _ := s.app.Store.ListContent(ctx, 500)
	approvals, _ := s.app.Store.ListApprovals(ctx, "pending", 500)
	topics, _ := s.app.Store.ListTopics(ctx, 500)
	pubs, _ := s.app.Store.ListPublications(ctx, 500)
	docs, _ := s.app.Store.CountDocuments(ctx)

	active, tasksCompleted := 0, 0
	cost := 0.0
	perAgent := map[string]map[string]int{}
	for _, run := range runs {
		if run.Status == domain.RunRunning || run.Status == domain.RunAwaitingApproval {
			active++
		}
		cost += run.Cost
		pa := perAgent[run.AgentID]
		if pa == nil {
			pa = map[string]int{}
			perAgent[run.AgentID] = pa
		}
		pa["total"]++
		if run.Status == domain.RunSucceeded {
			pa["success"]++
		}
	}
	for _, t := range tasks {
		if t.Status == domain.TaskSucceeded {
			tasksCompleted++
		}
	}
	contentByStatus := map[string]int{}
	contentReady := 0
	for _, c := range content {
		contentByStatus[string(c.Status)]++
		if c.Status == domain.ContentApproved || c.Status == domain.ContentScheduled || c.Status == domain.ContentPublished {
			contentReady++
		}
	}
	agentPerf := map[string]float64{}
	for id, pa := range perAgent {
		if pa["total"] == 0 {
			continue
		}
		agentPerf[id] = float64(pa["success"]) / float64(pa["total"]) * 100
	}

	writeJSON(w, 200, map[string]any{
		"active_runs":       active,
		"total_runs":        len(runs),
		"tasks_completed":   tasksCompleted,
		"content_ready":     contentReady,
		"llm_cost":          cost,
		"documents":         docs,
		"topics":            len(topics),
		"publications":      len(pubs),
		"approvals_pending": len(approvals),
		"content_by_status": contentByStatus,
		"agent_performance": agentPerf,
		"workspace":         s.app.Cfg.Workspace.Name,
	})
}

// ---------------- SSE 实时 Trace ----------------

func (s *Server) streamRun(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeErr(w, 500, "streaming unsupported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	deadline := time.Now().Add(60 * time.Second)
	last := ""
	for time.Now().Before(deadline) {
		run, _ := s.app.Store.GetRun(r.Context(), id)
		if run == nil {
			fmt.Fprintf(w, "event: error\ndata: {\"error\":\"run not found\"}\n\n")
			flusher.Flush()
			return
		}
		payload, _ := json.Marshal(run)
		if string(payload) != last {
			fmt.Fprintf(w, "event: run\ndata: %s\n\n", payload)
			flusher.Flush()
			last = string(payload)
		}
		if run.Status == domain.RunSucceeded || run.Status == domain.RunFailed || run.Status == domain.RunCancelled {
			fmt.Fprintf(w, "event: done\ndata: {\"status\":\"%s\"}\n\n", run.Status)
			flusher.Flush()
			return
		}
		select {
		case <-r.Context().Done():
			return
		case <-time.After(time.Second):
		}
	}
}

// ---------------- 辅助 ----------------

func queryInt(r *http.Request, key string, def int) int {
	if v := r.URL.Query().Get(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"error": msg})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
