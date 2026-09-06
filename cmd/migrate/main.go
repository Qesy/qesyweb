package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Qesy/QesyWeb/config"
	"github.com/Qesy/QesyWeb/internal/infra/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	if !cfg.MySQL.Enabled {
		panic("请在 .env 中设置 MYSQL_ENABLED=true")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := database.Open(ctx, cfg.MySQL)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	entries, err := os.ReadDir("migrations")
	if err != nil {
		panic(err)
	}
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)
	for _, name := range files {
		data, err := os.ReadFile(filepath.Join("migrations", name))
		if err != nil {
			panic(err)
		}
		for _, statement := range strings.Split(string(data), ";") {
			statement = strings.TrimSpace(statement)
			if statement == "" {
				continue
			}
			if _, err := db.ExecContext(ctx, statement); err != nil {
				panic(fmt.Errorf("执行迁移 %s 失败: %w", name, err))
			}
		}
	}
	fmt.Println("数据库迁移完成")
}
