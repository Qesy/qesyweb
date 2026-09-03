package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Qesy/QesyWeb/config"
	"github.com/Qesy/QesyWeb/internal/cache"
	"github.com/Qesy/QesyWeb/internal/database"
	"github.com/Qesy/QesyWeb/internal/news/model"
	"github.com/Qesy/QesyWeb/internal/news/repository"
	"github.com/Qesy/qesysql"
)

func TestNewsServiceIntegrationWithRedis(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.MySQL.Enabled || !cfg.Redis.Enabled {
		t.Skip("需要 MYSQL_ENABLED=true 且 REDIS_ENABLED=true")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	db, err := database.Open(ctx, cfg.MySQL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	redisCache, err := cache.Open(ctx, cfg.Redis)
	if err != nil {
		t.Fatal(err)
	}
	defer redisCache.Close()
	svc := New(repository.New(db), redisCache)
	status := fmt.Sprintf("integration-%d", time.Now().UnixNano())
	input := model.Form{Title: "缓存集成测试", Slug: status, Content: "cache", Status: "draft"}
	id, err := svc.Create(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.Table("news").SetWhere(qesysql.Eq{"id": id}).ExecDelete(context.Background())
		_ = redisCache.DeletePrefix(context.Background(), cache.NewsListPrefix)
		_ = redisCache.DeletePrefix(context.Background(), cache.NewsDetailPrefix)
	}()
	first, err := svc.Find(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.Find(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID {
		t.Fatalf("cache result mismatch: %#v %#v", first, second)
	}
	page, err := svc.List(ctx, 1, 20, "draft")
	if err != nil {
		t.Fatal(err)
	}
	if page.Total < 1 {
		t.Fatalf("page total = %d", page.Total)
	}
	if err := svc.Update(ctx, id, model.Form{Title: "缓存更新", Slug: status, Content: "changed", Status: "draft"}); err != nil {
		t.Fatal(err)
	}
	updated, err := svc.Find(ctx, id)
	if err != nil || updated.Title != "缓存更新" {
		t.Fatalf("updated = %#v err=%v", updated, err)
	}
}
