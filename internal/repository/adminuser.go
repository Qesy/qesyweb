package repository

import (
	"context"

	"github.com/Qesy/QesyWeb/internal/model"
	"github.com/Qesy/qesysql"
)

type AdminUserRepository struct{ db *qesysql.DB }

func NewAdminUserRepository(db *qesysql.DB) *AdminUserRepository { return &AdminUserRepository{db: db} }
func (r *AdminUserRepository) FindByUsername(ctx context.Context, username string) (model.AdminUser, error) {
	var user model.AdminUser
	err := r.db.Table("admin_users").SetField("id", "username", "password_hash", "email", "status").SetWhere(qesysql.Eq{"username": username}).ExecSelectOne(ctx, &user)
	return user, err
}
