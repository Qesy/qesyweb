package adminuser

type AdminUser struct {
	ID           int64  `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
	Email        string `db:"email"`
	Status       string `db:"status"`
}
