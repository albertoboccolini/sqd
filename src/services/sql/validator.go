package sql

import (
	"strings"

	"github.com/overthinkinglabs/sqd/src/models/displayable_errors"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (sqlValidator *Validator) Validate(sql string) error {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return displayable_errors.NewInvalidQueryError("Query cannot be empty")
	}

	upperSql := strings.ToUpper(sql)

	validPrefixes := []string{"SELECT COUNT", "SELECT", "UPDATE", "DELETE"}
	hasValidPrefix := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(upperSql, prefix) {
			hasValidPrefix = true
			break
		}
	}

	if !hasValidPrefix {
		return displayable_errors.NewInvalidQueryError("Query must start with SELECT, UPDATE, or DELETE")
	}

	if strings.HasPrefix(upperSql, "SELECT") && !strings.Contains(upperSql, "FROM") {
		return displayable_errors.NewInvalidQueryError("SELECT query must contain FROM clause")
	}

	if strings.HasPrefix(upperSql, "UPDATE") && !strings.Contains(upperSql, "SET") {
		return displayable_errors.NewInvalidQueryError("UPDATE query must contain SET clause")
	}

	if strings.HasPrefix(upperSql, "DELETE") && !strings.Contains(upperSql, "FROM") {
		return displayable_errors.NewInvalidQueryError("DELETE query must contain FROM clause")
	}

	return nil
}
