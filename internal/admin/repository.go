package admin

import (
	"context"

	"github.com/Qesy/qesysql"
)

type User struct {
	ID           int64  `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
	Email        string `db:"email"`
	Status       string `db:"status"`
}
type Repository struct{ db *qesysql.DB }

func NewRepository(db *qesysql.DB) *Repository { return &Repository{db: db} }
func (r *Repository) FindByUsername(ctx context.Context, username string) (User, error) {
	var user User
	err := r.db.Table("admin_users").SetField("id", "username", "password_hash", "email", "status").SetWhere(qesysql.Eq{"username": username}).ExecSelectOne(ctx, &user)
	return user, err
}
