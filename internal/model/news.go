package model

import "time"

type News struct {
	ID          int64      `db:"id"`
	Title       string     `db:"title"`
	Slug        string     `db:"slug"`
	Summary     string     `db:"summary"`
	Content     string     `db:"content"`
	Status      string     `db:"status"`
	PublishedAt *time.Time `db:"published_at"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}
type NewsAuditLog struct {
	ID        int64     `db:"id"`
	NewsID    int64     `db:"news_id"`
	Action    string    `db:"action"`
	CreatedAt time.Time `db:"created_at"`
}
type NewsPage struct {
	Items                 []News
	Page, PageSize, Total int
}
type NewsForm struct {
	Title, Slug, Summary, Content, Status string
	Action, Error                         string
}
