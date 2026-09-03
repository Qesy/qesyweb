package cache

import "fmt"

// Redis Key 集中放在这里，便于统一版本、命名空间和失效前缀。
const (
	NewsListPrefix   = "news:v1:list:"
	NewsDetailPrefix = "news:v1:detail:"
	NewsPublicDetailPrefix = "news:v1:public:detail:"
	SessionPrefix    = "session:v1:"
)

func NewsList(page, pageSize int, status string) string {
	return fmt.Sprintf("%s%d:%d:%s", NewsListPrefix, page, pageSize, status)
}

func NewsDetail(id int64) string {
	return fmt.Sprintf("%s%d", NewsDetailPrefix, id)
}

func NewsPublicDetail(id int64) string {
	return fmt.Sprintf("%s%d", NewsPublicDetailPrefix, id)
}

func Session(id string) string {
	return SessionPrefix + id
}
