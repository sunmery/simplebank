package pkg

import (
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"simple_bank/constants"
)

func PgErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		fmt.Printf("postgres sql err message is '%s' \n", pgErr.Message)
		fmt.Printf("postgres sql err code is '%s' \n", pgErr.Code)

		switch pgErr.Code {
		case constants.ForeignKeyViolation:
			return ""
		case constants.UniqueViolation:
			return ""
		}
	}
	return ""
}
