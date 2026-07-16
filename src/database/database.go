package database

import (
	"database/sql"

	gsDatabase "github.com/gerp93/gameshell-framework/database"
)

func query(sqlString string, params ...any) (*sql.Rows, error) {
	return gsDatabase.Query(sqlString, params...)
}

func execute(sqlString string, params ...any) error {
	return gsDatabase.Execute(sqlString, params...)
}
