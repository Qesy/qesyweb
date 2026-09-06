package adminuser

import (
	"context"

	"github.com/Qesy/qesysql"
)

type Repository struct{ db *qesysql.DB }

func NewRepository(db *qesysql.DB) *Repository { return &Repository{db: db} }
func (r *Repository) FindByUsername(ctx context.Context, username string) (AdminUser, error) {
	var user AdminUser
	err := r.db.Table("admin_users").SetField("id", "username", "password_hash", "email", "status").SetWhere(qesysql.Eq{"username": username}).ExecSelectOne(ctx, &user)
	return user, err
}
