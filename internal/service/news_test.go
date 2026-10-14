package service

import (
	"context"
	"testing"

	"github.com/Qesy/QesyWeb/internal/model"
)

func TestCreateValidation(t *testing.T) {
	svc := &NewsService{}
	_, err := svc.Create(context.Background(), model.NewsForm{})
	if err == nil || err.Error() != "标题不能为空" {
		t.Fatalf("error = %v", err)
	}
}
