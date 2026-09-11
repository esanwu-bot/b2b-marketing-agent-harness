// Package observability 提供结构化日志与运行指标的基础设施。
package observability

import (
	"log/slog"
	"os"
)

// New 创建结构化日志器。
func New() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}
