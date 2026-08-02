//go:build tools

// Package tools tracks development tool dependencies so their versions are
// pinned in go.mod. It is never imported or compiled into the server binary.
package tools

import (
	_ "github.com/air-verse/air"
	_ "github.com/bufbuild/buf/cmd/buf"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
)
