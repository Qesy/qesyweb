package repository

import (
	"context"
	"database/sql"

	"github.com/Qesy/QesyWeb/internal/modules/news/model"
	"github.com/Qesy/qesysql"
)

const newsTable = "news"

type Repository struct{ db *qesysql.DB }

func New(db *qesysql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) List(ctx context.Context, page, pageSize int, status string) ([]model.News, error) {
	var items []model.News
	query := r.db.Table(newsTable).SetField("id", "title", "slug", "summary", "content", "status", "published_at", "created_at", "updated_at").SetSort("id DESC").SetLimit(pageSize).SetOffset((page - 1) * pageSize)
	if status != "" {
		query.SetWhere(qesysql.Eq{"status": status})
	}
	return items, query.ExecSelect(ctx, &items)
}
func (r *Repository) Count(ctx context.Context, status string) (int, error) {
	// QesySQL currently focuses on row CRUD; count is the documented sqlx escape hatch.
	var total int
	if status == "" {
		err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM news")
		return total, err
	}
	err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM news WHERE status = ?", status)
	return total, err
}
func (r *Repository) Find(ctx context.Context, id int64) (model.News, error) {
	var item model.News
	err := r.db.Table(newsTable).SetField("id", "title", "slug", "summary", "content", "status", "published_at", "created_at", "updated_at").SetWhere(qesysql.Eq{"id": id}).ExecSelectOne(ctx, &item)
	return item, err
}
func (r *Repository) FindPublished(ctx context.Context, id int64) (model.News, error) {
	var item model.News
	err := r.db.Table(newsTable).SetField("id", "title", "slug", "summary", "content", "status", "published_at", "created_at", "updated_at").SetWhere(qesysql.Eq{"id": id, "status": "published"}).ExecSelectOne(ctx, &item)
	return item, err
}
func (r *Repository) Create(ctx context.Context, input model.Form) (int64, error) {
	result, err := r.db.Table(newsTable).SetInsert(qesysql.Values{"title": input.Title, "slug": input.Slug, "summary": input.Summary, "content": input.Content, "status": input.Status}).ExecInsert(ctx)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
func (r *Repository) Update(ctx context.Context, id int64, input model.Form) error {
	_, err := r.db.Table(newsTable).SetUpdate(qesysql.Values{"title": input.Title, "slug": input.Slug, "summary": input.Summary, "content": input.Content, "status": input.Status}).SetWhere(qesysql.Eq{"id": id}).ExecUpdate(ctx)
	return err
}
func (r *Repository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Table(newsTable).SetWhere(qesysql.Eq{"id": id}).ExecDelete(ctx)
	return err
}
func (r *Repository) CreateWithAudit(ctx context.Context, input model.Form) error {
	return r.db.InTx(ctx, nil, func(tx *qesysql.Tx) error {
		result, err := tx.Table(newsTable).SetInsert(qesysql.Values{"title": input.Title, "slug": input.Slug, "summary": input.Summary, "content": input.Content, "status": input.Status}).ExecInsert(ctx)
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		_, err = tx.Table("news_audit_logs").SetInsert(qesysql.Values{"news_id": id, "action": "create"}).ExecInsert(ctx)
		return err
	})
}
func IsNotFound(err error) bool { return err == sql.ErrNoRows }
