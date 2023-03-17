package authn

import "database/sql"

type IUserStore interface {
	GetUserById(db *sql.DB, id string) (User, error)
	GetUserByAuth(db *sql.DB, auth UserAuth) (User, error)
}
