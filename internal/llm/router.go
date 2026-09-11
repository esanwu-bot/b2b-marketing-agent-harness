package llm

import (
	"context"
	"fmt"
	"sync"
)

// Router 按名称路由到不同的 Provider，并保留默认 Provider。
type Router struct {
	mu        sync.RWMutex
	providers map[string]Provider
	def       Provider
}

// NewRouter 创建路由器。
func NewRouter(def Provider) *Router {
	r := &Router{providers: map[string]Provider{}}
	if def != nil {
		r.def = def
		r.providers[def.Name()] = def
	}
	return r
}

// Register 注册一个 Provider。
func (r *Router) Register(p Provider) {
	if p == nil {
		return
	}
	r.mu.Lock()
	r.providers[p.Name()] = p
	if r.def == nil {
		r.def = p
	}
	r.mu.Unlock()
}

// Default 返回默认 Provider。
func (r *Router) Default() Provider { return r.def }

// Get 按名称取 Provider，找不到返回默认。
func (r *Router) Get(name string) Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if p, ok := r.providers[name]; ok {
		return p
	}
	return r.def
}

// Chat 用默认 Provider 调用。
func (r *Router) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if r.def == nil {
		return nil, fmt.Errorf("未配置任何 LLM Provider")
	}
	return r.def.Chat(ctx, req)
}
