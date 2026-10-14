package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Qesy/QesyWeb/config"
	"github.com/Qesy/QesyWeb/internal/infra/database"
	"github.com/Qesy/QesyWeb/internal/model"
	"github.com/Qesy/qesysql"
)

func TestNewsRepositoryIntegration(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.MySQL.Enabled {
		t.Skip("MYSQL_ENABLED 未开启")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	db, err := database.Open(ctx, cfg.MySQL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := NewNewsRepository(db)
	slug := fmt.Sprintf("integration-%d", time.Now().UnixNano())
	input := model.NewsForm{Title: "集成测试新闻", Slug: slug, Summary: "summary", Content: "content", Status: "draft"}
	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.Table("news_audit_logs").SetWhere(qesysql.Eq{"news_id": id}).ExecDelete(context.Background())
		_, _ = db.Table("news").SetWhere(qesysql.Eq{"id": id}).ExecDelete(context.Background())
	}()
	item, err := repo.Find(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if item.Title != input.Title || item.Slug != slug {
		t.Fatalf("item = %#v", item)
	}
	items, err := repo.List(ctx, 1, 10, "draft")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) == 0 {
		t.Fatal("created news missing from list")
	}
	if total, err := repo.Count(ctx, "draft"); err != nil || total < 1 {
		t.Fatalf("count=%d err=%v", total, err)
	}
	input.Title = "更新后的新闻"
	if err := repo.Update(ctx, id, input); err != nil {
		t.Fatal(err)
	}
	item, err = repo.Find(ctx, id)
	if err != nil || item.Title != input.Title {
		t.Fatalf("updated item = %#v err=%v", item, err)
	}
	if err := repo.Delete(ctx, id); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateWithAudit(ctx, model.NewsForm{Title: "事务测试", Slug: slug + "-tx", Content: "tx", Status: "draft"}); err != nil {
		t.Fatal(err)
	}
}
