package db

import "database/sql"

func NullInt64Value(n sql.NullInt64, defaultVal int64) int64 {
	if n.Valid {
		return n.Int64
	}
	return defaultVal
}

func NullStringValue(n sql.NullString, defaultVal string) string {
	if n.Valid {
		return n.String
	}
	return defaultVal
}

func NullInt64ToBool(n sql.NullInt64) bool {
	return n.Valid && n.Int64 == 1
}
