package service

import (
	"context"
	"testing"

	"github.com/Qesy/QesyWeb/internal/news/model"
)

func TestCreateValidation(t *testing.T) {
	svc := &Service{}
	_, err := svc.Create(context.Background(), model.Form{})
	if err == nil || err.Error() != "标题不能为空" {
		t.Fatalf("error = %v", err)
	}
}
