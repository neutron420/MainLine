package interceptor

import (
	"context"
	"fmt"
)

func UserIDFromContext(ctx context.Context) (string, error) {
	v := ctx.Value(UserIDKey)
	if v == nil {
		return "", fmt.Errorf("user not authenticated")
	}
	return fmt.Sprintf("%v", v), nil
}

func UserEmailFromContext(ctx context.Context) (string, error) {
	v := ctx.Value(UserEmailKey)
	if v == nil {
		return "", fmt.Errorf("user not authenticated")
	}
	return fmt.Sprintf("%v", v), nil
}

func UserRoleFromContext(ctx context.Context) (string, error) {
	v := ctx.Value(UserRoleKey)
	if v == nil {
		return "", fmt.Errorf("user not authenticated")
	}
	return fmt.Sprintf("%v", v), nil
}
