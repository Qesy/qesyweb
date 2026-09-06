package service

import (
	"context"
	"errors"
	"time"

	newsdata "github.com/Qesy/QesyWeb/internal/data/news"
	"github.com/Qesy/QesyWeb/internal/infra/cache"
)

type Service struct {
	repo  *newsdata.Repository
	cache *cache.Cache
}

func New(repo *newsdata.Repository, cache *cache.Cache) *Service {
	return &Service{repo: repo, cache: cache}
}
func (s *Service) List(ctx context.Context, page, pageSize int, status string) (newsdata.Page, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	key := cache.NewsList(page, pageSize, status)
	var result newsdata.Page
	if s.cache != nil && s.cache.GetJSON(ctx, key, &result) == nil {
		return result, nil
	}
	items, err := s.repo.List(ctx, page, pageSize, status)
	if err != nil {
		return result, err
	}
	total, err := s.repo.Count(ctx, status)
	if err != nil {
		return result, err
	}
	result = newsdata.Page{Items: items, Page: page, PageSize: pageSize, Total: total}
	if s.cache != nil {
		_ = s.cache.SetJSON(ctx, key, result, 30*time.Second)
	}
	return result, nil
}
func (s *Service) Find(ctx context.Context, id int64) (newsdata.News, error) {
	key := cache.NewsDetail(id)
	var item newsdata.News
	if s.cache != nil && s.cache.GetJSON(ctx, key, &item) == nil {
		return item, nil
	}
	item, err := s.repo.Find(ctx, id)
	if err != nil {
		return item, err
	}
	if s.cache != nil {
		_ = s.cache.SetJSON(ctx, key, item, 5*time.Minute)
	}
	return item, nil
}
func (s *Service) FindPublished(ctx context.Context, id int64) (newsdata.News, error) {
	key := cache.NewsPublicDetail(id)
	var item newsdata.News
	if s.cache != nil && s.cache.GetJSON(ctx, key, &item) == nil {
		return item, nil
	}
	item, err := s.repo.FindPublished(ctx, id)
	if err != nil {
		return item, err
	}
	if s.cache != nil {
		_ = s.cache.SetJSON(ctx, key, item, 5*time.Minute)
	}
	return item, nil
}
func (s *Service) Create(ctx context.Context, input newsdata.Form) (int64, error) {
	if err := validate(input); err != nil {
		return 0, err
	}
	id, err := s.repo.Create(ctx, input)
	s.invalidate(ctx, id)
	return id, err
}
func (s *Service) CreateWithAudit(ctx context.Context, input newsdata.Form) error {
	if err := validate(input); err != nil {
		return err
	}
	err := s.repo.CreateWithAudit(ctx, input)
	s.invalidate(ctx, 0)
	return err
}
func (s *Service) Update(ctx context.Context, id int64, input newsdata.Form) error {
	if err := validate(input); err != nil {
		return err
	}
	err := s.repo.Update(ctx, id, input)
	s.invalidate(ctx, id)
	return err
}
func (s *Service) Delete(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, id)
	s.invalidate(ctx, id)
	return err
}
func (s *Service) invalidate(ctx context.Context, id int64) {
	if s.cache == nil {
		return
	}
	if id > 0 {
		_ = s.cache.Delete(ctx, cache.NewsDetail(id))
		_ = s.cache.Delete(ctx, cache.NewsPublicDetail(id))
	}
	_ = s.cache.DeletePrefix(ctx, cache.NewsListPrefix)
}
func validate(input newsdata.Form) error {
	if input.Title == "" {
		return errors.New("标题不能为空")
	}
	if input.Slug == "" {
		return errors.New("Slug 不能为空")
	}
	if input.Status == "" {
		return errors.New("状态不能为空")
	}
	return nil
}
