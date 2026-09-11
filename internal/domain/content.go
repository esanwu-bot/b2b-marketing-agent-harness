package domain

import "time"

// ChannelKind 渠道类型。
type ChannelKind string

const (
	ChannelLinkedIn   ChannelKind = "linkedin"
	ChannelWebsite    ChannelKind = "website"
	ChannelYouTube    ChannelKind = "youtube"
	ChannelNewsletter ChannelKind = "newsletter"
	ChannelX          ChannelKind = "x"
	ChannelEmail      ChannelKind = "email"
)

// Topic 是内容选题（对应 topics 表）。
type Topic struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id,omitempty"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	Category    string    `json:"category,omitempty"`
	Keywords    []string  `json:"keywords,omitempty"`
	SourceRefs  []string  `json:"source_refs,omitempty"`
	Score       float64   `json:"score"`
	ClusterKey  string    `json:"cluster_key,omitempty"`
	Status      string    `json:"status"` // candidate | approved | rejected | used
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ContentStatus 内容状态。
type ContentStatus string

const (
	ContentDraft     ContentStatus = "draft"
	ContentInReview  ContentStatus = "in_review"
	ContentApproved  ContentStatus = "approved"
	ContentScheduled ContentStatus = "scheduled"
	ContentPublished ContentStatus = "published"
	ContentRejected  ContentStatus = "rejected"
	ContentFailed    ContentStatus = "failed"
)

// Content 是一篇内容（对应 content 表）。
type Content struct {
	ID            string         `json:"id"`
	WorkspaceID   string         `json:"workspace_id,omitempty"`
	TopicID       string         `json:"topic_id,omitempty"`
	ChannelKind   ChannelKind    `json:"channel_kind"`
	Format        string         `json:"format"` // post | article | video_script | newsletter
	Language      string         `json:"language"`
	Title         string         `json:"title"`
	Body          string         `json:"body"`
	Hashtags      []string       `json:"hashtags,omitempty"`
	Status        ContentStatus  `json:"status"`
	AuthorAgentID string         `json:"author_agent_id,omitempty"`
	CreatedBy     string         `json:"created_by,omitempty"`
	ScheduledAt   *time.Time     `json:"scheduled_at,omitempty"`
	PublishedAt   *time.Time     `json:"published_at,omitempty"`
	ExternalURL   string         `json:"external_url,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// Publication 是内容在某渠道的一次发布（对应 publications 表）。
type Publication struct {
	ID          string      `json:"id"`
	ContentID   string      `json:"content_id"`
	ChannelKind ChannelKind `json:"channel_kind"`
	Status      string      `json:"status"` // pending | scheduled | publishing | published | failed
	ExternalID  string      `json:"external_id,omitempty"`
	ExternalURL string      `json:"external_url,omitempty"`
	ScheduledAt *time.Time  `json:"scheduled_at,omitempty"`
	PublishedAt *time.Time  `json:"published_at,omitempty"`
	Error       string      `json:"error,omitempty"`
}

// Metric 是一次内容表现指标（对应 content_metrics 表）。
type Metric struct {
	PublicationID  string    `json:"publication_id"`
	MeasuredAt     time.Time `json:"measured_at"`
	Impressions    int       `json:"impressions"`
	Reach          int       `json:"reach"`
	Likes          int       `json:"likes"`
	Comments       int       `json:"comments"`
	Shares         int       `json:"shares"`
	Clicks         int       `json:"clicks"`
	EngagementRate float64   `json:"engagement_rate"`
	CTR            float64   `json:"ctr"`
}
