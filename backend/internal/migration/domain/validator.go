package domain

import (
	"fmt"
	"regexp"
	"strings"
)

var disallowedStatements = []string{
	"DROP DATABASE",
	"DROP TABLESPACE",
	"ALTER SYSTEM",
	"DROP EXTENSION",
}

type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

type SQLValidator struct{}

func NewSQLValidator() *SQLValidator {
	return &SQLValidator{}
}

type statementInfo struct {
	raw    string
	verb   string
	object string
	name   string
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

func (v *SQLValidator) ValidateWithWarnings(upSQL, downSQL string) *ValidationResult {
	result := &ValidationResult{}

	if upSQL == "" {
		result.Errors = append(result.Errors, "up_sql is required")
	}

	errs, warns := v.validateSQLWithWarnings(upSQL, "up")
	result.Errors = append(result.Errors, errs...)
	result.Warnings = append(result.Warnings, warns...)

	if downSQL != "" {
		errs, warns = v.validateSQLWithWarnings(downSQL, "down")
		result.Errors = append(result.Errors, errs...)
		result.Warnings = append(result.Warnings, warns...)
	}

	result.Valid = len(result.Errors) == 0
	return result
}

func (v *SQLValidator) validateSQL(sql string, label string) []string {
	errs, _ := v.validateSQLWithWarnings(sql, label)
	return errs
}

func (v *SQLValidator) validateSQLWithWarnings(sql string, label string) ([]string, []string) {
	var errors []string
	var warnings []string

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

		info := parseStatement(trimmed)
		if err := v.validateStatement(info, i+1, label); err != "" {
			errors = append(errors, err)
		}
		if warn := v.warnStatement(info, i+1, label); warn != "" {
			warnings = append(warnings, warn)
		}
	}

	return errors, warnings
}

var ddlVerbs = []string{"CREATE", "ALTER", "DROP", "TRUNCATE", "RENAME", "COMMENT"}
var dmlVerbs = []string{"SELECT", "INSERT", "UPDATE", "DELETE"}
var txnVerbs = []string{"BEGIN", "COMMIT", "ROLLBACK"}
var otherVerbs = []string{"GRANT", "REVOKE", "SET", "ANALYZE", "VACUUM", "REINDEX"}

func parseStatement(sql string) statementInfo {
	upper := strings.TrimSpace(strings.ToUpper(sql))
	parts := strings.Fields(upper)

	info := statementInfo{raw: sql}

	if len(parts) == 0 {
		return info
	}

	info.verb = parts[0]

	if len(parts) >= 3 && (parts[0] == "CREATE" || parts[0] == "DROP" || parts[0] == "ALTER") {
		info.object = parts[1]
		info.name = parts[2]
		if info.object == "TABLE" || info.object == "INDEX" || info.object == "VIEW" ||
			info.object == "SEQUENCE" || info.object == "FUNCTION" || info.object == "TRIGGER" ||
			info.object == "TYPE" || info.object == "SCHEMA" || info.object == "EXTENSION" ||
			info.object == "COLUMN" || info.object == "CONSTRAINT" {
			if info.object == "COLUMN" && len(parts) >= 4 {
				info.name = parts[2] + " " + parts[3]
			}
		} else if strings.Contains(upper, "IF NOT EXISTS") {
			for i, p := range parts {
				if p == "TABLE" || p == "INDEX" || p == "VIEW" || p == "SCHEMA" || p == "SEQUENCE" {
					info.object = p
					if i+1 < len(parts) {
						info.name = parts[i+1]
					}
					break
				}
			}
		}
	}

	return info
}

func (v *SQLValidator) validateStatement(info statementInfo, stmtNum int, label string) string {
	if info.verb == "" {
		return ""
	}

	allVerbs := append(append(append(ddlVerbs, dmlVerbs...), txnVerbs...), otherVerbs...)
	recognized := false
	for _, v := range allVerbs {
		if info.verb == v {
			recognized = true
			break
		}
	}

	if !recognized {
		return fmt.Sprintf("%s statement %d: unrecognized SQL command %q", label, stmtNum, info.verb)
	}

	if info.verb == "DROP" && info.object == "TABLE" && !strings.Contains(strings.ToUpper(info.raw), "IF EXISTS") {
		return fmt.Sprintf("%s statement %d: DROP TABLE without IF EXISTS is dangerous", label, stmtNum)
	}

	if !strings.HasSuffix(strings.TrimRight(info.raw, " \n\r\t"), ";") {
		return fmt.Sprintf("%s statement %d: must end with semicolon", label, stmtNum)
	}

	return ""
}

var dropTableWarning = regexp.MustCompile(`(?i)\bDROP\s+TABLE\b`)

func (v *SQLValidator) warnStatement(info statementInfo, stmtNum int, label string) string {
	if info.verb == "ALTER" {
		if strings.Contains(strings.ToUpper(info.raw), "DROP") &&
			(strings.Contains(strings.ToUpper(info.raw), "COLUMN") ||
				strings.Contains(strings.ToUpper(info.raw), "CONSTRAINT")) {
			return fmt.Sprintf("%s statement %d: ALTER DROP is destructive — verify downstream impact", label, stmtNum)
		}
	}

	if info.verb == "DROP" && info.object != "TABLE" && info.object != "" {
		return fmt.Sprintf("%s statement %d: DROP %s is potentially destructive", label, stmtNum, info.object)
	}

	return ""
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
