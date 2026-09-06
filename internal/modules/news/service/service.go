package service

import (
	"context"
	"errors"
	"time"

	"github.com/Qesy/QesyWeb/internal/infra/cache"
	"github.com/Qesy/QesyWeb/internal/modules/news/model"
	"github.com/Qesy/QesyWeb/internal/modules/news/repository"
)

type Service struct {
	repo  *repository.Repository
	cache *cache.Cache
}

func New(repo *repository.Repository, cache *cache.Cache) *Service {
	return &Service{repo: repo, cache: cache}
}
func (s *Service) List(ctx context.Context, page, pageSize int, status string) (model.Page, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	key := cache.NewsList(page, pageSize, status)
	var result model.Page
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
	result = model.Page{Items: items, Page: page, PageSize: pageSize, Total: total}
	if s.cache != nil {
		_ = s.cache.SetJSON(ctx, key, result, 30*time.Second)
	}
	return result, nil
}
func (s *Service) Find(ctx context.Context, id int64) (model.News, error) {
	key := cache.NewsDetail(id)
	var item model.News
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
func (s *Service) FindPublished(ctx context.Context, id int64) (model.News, error) {
	key := cache.NewsPublicDetail(id)
	var item model.News
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
func (s *Service) Create(ctx context.Context, input model.Form) (int64, error) {
	if err := validate(input); err != nil {
		return 0, err
	}
	id, err := s.repo.Create(ctx, input)
	s.invalidate(ctx, id)
	return id, err
}
func (s *Service) CreateWithAudit(ctx context.Context, input model.Form) error {
	if err := validate(input); err != nil {
		return err
	}
	err := s.repo.CreateWithAudit(ctx, input)
	s.invalidate(ctx, 0)
	return err
}
func (s *Service) Update(ctx context.Context, id int64, input model.Form) error {
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
func validate(input model.Form) error {
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
