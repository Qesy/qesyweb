package service

import (
	"context"
	"testing"

	newsdata "github.com/Qesy/QesyWeb/internal/data/news"
)

func TestCreateValidation(t *testing.T) {
	svc := &Service{}
	_, err := svc.Create(context.Background(), newsdata.Form{})
	if err == nil || err.Error() != "标题不能为空" {
		t.Fatalf("error = %v", err)
	}
}
