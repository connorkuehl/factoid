// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.2

package sqlite

import (
	"database/sql"
)

type Fact struct {
	ID        int64
	CreatedAt sql.NullTime
	UpdatedAt sql.NullTime
	DeletedAt sql.NullTime
	Content   string
	Source    sql.NullString
}
