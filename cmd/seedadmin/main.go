package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Qesy/QesyWeb/config"
	"github.com/Qesy/QesyWeb/internal/infra/database"
	"github.com/Qesy/qesysql"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	if !cfg.MySQL.Enabled {
		panic("set MYSQL_ENABLED=true in .env first")
	}
	username := prompt("Admin username: ")
	password := promptPassword("Admin password: ")
	email := prompt("Admin email: ")
	if username == "" || password == "" || email == "" {
		panic("username, password, and email are required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := database.Open(ctx, cfg.MySQL)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	var existing struct {
		ID int64 `db:"id"`
	}
	err = db.Table("admin_users").SetField("id").SetWhere(qesysql.Eq{"username": username}).ExecSelectOne(ctx, &existing)
	if err == sql.ErrNoRows {
		_, err = db.Table("admin_users").SetInsert(qesysql.Values{"username": username, "password_hash": string(hash), "email": email, "status": "active"}).ExecInsert(ctx)
	} else if err == nil {
		_, err = db.Table("admin_users").SetUpdate(qesysql.Values{"password_hash": string(hash), "email": email, "status": "active"}).SetWhere(qesysql.Eq{"id": existing.ID}).ExecUpdate(ctx)
	}
	if err != nil {
		panic(err)
	}
	fmt.Println("Administrator account saved:", username)
}

func prompt(label string) string { fmt.Print(label); return strings.TrimSpace(readLine()) }
func promptPassword(label string) string {
	fmt.Print(label)
	data, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err == nil {
		return strings.TrimSpace(string(data))
	}
	return strings.TrimSpace(readLine())
}
func readLine() string { line, _ := bufio.NewReader(os.Stdin).ReadString('\n'); return line }
