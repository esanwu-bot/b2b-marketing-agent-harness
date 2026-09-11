package domain

import "time"

// Document 是知识库中的一篇文档（对应 knowledge_documents）。
type Document struct {
	ID          string         `json:"id"`
	WorkspaceID string         `json:"workspace_id,omitempty"`
	SourceType  string         `json:"source_type"` // crawl | rss | website | manual | linkedin | upload | api
	SourceID    string         `json:"source_id,omitempty"`
	Title       string         `json:"title"`
	URL         string         `json:"url,omitempty"`
	Language    string         `json:"language,omitempty"`
	Summary     string         `json:"summary,omitempty"`
	Content     string         `json:"content"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Hash        string         `json:"hash,omitempty"`
	FetchedAt   time.Time      `json:"fetched_at"`
}

// Chunk 是文档分块（对应 knowledge_chunks），供检索。
type Chunk struct {
	ID         string         `json:"id"`
	DocumentID string         `json:"document_id"`
	Seq        int            `json:"seq"`
	Content    string         `json:"content"`
	Tokens     int            `json:"token_count"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// SearchHit 是检索命中。
type SearchHit struct {
	DocumentID string  `json:"document_id"`
	Title      string  `json:"title"`
	URL        string  `json:"url,omitempty"`
	Snippet    string  `json:"snippet"`
	Score      float64 `json:"score"`
}

// Entity 是知识实体（product/technology/company/person/application/topic）。
type Entity struct {
	ID          string         `json:"id"`
	WorkspaceID string         `json:"workspace_id,omitempty"`
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Canonical   string         `json:"canonical_name"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}
