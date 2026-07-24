package domain

import (
	"fmt"
	"strings"
)

var disallowedStatements = []string{
	"DROP DATABASE",
	"DROP TABLESPACE",
	"ALTER SYSTEM",
	"DROP EXTENSION",
}

type SQLValidator struct{}

func NewSQLValidator() *SQLValidator {
	return &SQLValidator{}
}

func (v *SQLValidator) Validate(upSQL, downSQL string) (bool, []string) {
	var errors []string

	if upSQL == "" {
		errors = append(errors, "up_sql is required")
	}

	errs := v.validateSQL(upSQL, "up")
	errors = append(errors, errs...)

	if downSQL != "" {
		errs = v.validateSQL(downSQL, "down")
		errors = append(errors, errs...)
	}

	return len(errors) == 0, errors
}

func (v *SQLValidator) validateSQL(sql string, label string) []string {
	var errors []string
	upper := strings.ToUpper(strings.TrimSpace(sql))

	for _, banned := range disallowedStatements {
		if strings.Contains(upper, banned) {
			errors = append(errors, fmt.Sprintf("%s: statement %q is not allowed", label, banned))
		}
	}

	stmts := splitStatements(sql)
	for i, stmt := range stmts {
		trimmed := strings.TrimSpace(stmt)
		if trimmed == "" {
			continue
		}
		if err := v.basicParse(trimmed); err != nil {
			errors = append(errors, fmt.Sprintf("%s statement %d: %v", label, i+1, err))
		}
	}

	return errors
}

func (v *SQLValidator) basicParse(sql string) error {
	upper := strings.ToUpper(sql)

	knownCommands := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "ALTER", "DROP",
		"TRUNCATE", "GRANT", "REVOKE", "BEGIN", "COMMIT", "ROLLBACK",
		"SET", "COMMENT", "RENAME", "ANALYZE", "VACUUM", "REINDEX",
	}

	hasKnown := false
	for _, cmd := range knownCommands {
		if strings.HasPrefix(upper, cmd) {
			hasKnown = true
			break
		}
	}

	if !hasKnown {
		return fmt.Errorf("unrecognized SQL command")
	}

	if !strings.HasSuffix(strings.TrimRight(sql, " \n\r\t"), ";") {
		return fmt.Errorf("statement must end with semicolon")
	}

	return nil
}

func splitStatements(sql string) []string {
	var stmts []string
	current := strings.Builder{}
	depth := 0
	inString := false
	stringChar := byte(0)

	for i := 0; i < len(sql); i++ {
		ch := sql[i]

		if inString {
			current.WriteByte(ch)
			if ch == stringChar && (i+1 >= len(sql) || sql[i+1] != stringChar) {
				inString = false
			}
			continue
		}

		switch ch {
		case '\'', '"':
			inString = true
			stringChar = ch
			current.WriteByte(ch)
		case '(':
			depth++
			current.WriteByte(ch)
		case ')':
			depth--
			current.WriteByte(ch)
		case ';':
			if depth == 0 {
				current.WriteByte(ch)
				stmts = append(stmts, current.String())
				current.Reset()
			} else {
				current.WriteByte(ch)
			}
		case '-':
			if i+1 < len(sql) && sql[i+1] == '-' {
				for i < len(sql) && sql[i] != '\n' {
					i++
				}
				continue
			}
			current.WriteByte(ch)
		default:
			current.WriteByte(ch)
		}
	}

	remaining := strings.TrimSpace(current.String())
	if remaining != "" {
		stmts = append(stmts, remaining)
	}

	return stmts
}
