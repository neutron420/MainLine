package interceptor

import (
	"context"
	"reflect"
	"strings"

	"github.com/schemahub/backend/internal/audit/domain"
	"google.golang.org/grpc"
)

// skipAuditMethod returns true for RPCs that must not be audit-logged:
// auth endpoints (credential flows) and pure read/validation methods.
func skipAuditMethod(fullMethod string) bool {
	if !strings.HasPrefix(fullMethod, "/schemahub.") {
		return true
	}
	if strings.Contains(fullMethod, "AuthService/") {
		return true
	}
	method := fullMethod[strings.LastIndex(fullMethod, "/")+1:]
	for _, p := range []string{"Get", "List", "Tail", "Watch", "Validate", "DryRun"} {
		if strings.HasPrefix(method, p) {
			return true
		}
	}
	return false
}

// auditResourceType maps a method name to a coarse resource type.
func auditResourceType(method string) string {
	switch {
	case strings.HasPrefix(method, "CreateConnection"), strings.HasPrefix(method, "GetConnection"),
		strings.HasPrefix(method, "ListConnections"), strings.HasPrefix(method, "UpdateConnection"),
		strings.HasPrefix(method, "DeleteConnection"), strings.HasPrefix(method, "TestConnection"):
		return "connection"
	case strings.HasPrefix(method, "InviteMember"), strings.HasPrefix(method, "AddMember"),
		strings.HasPrefix(method, "RemoveMember"), strings.HasPrefix(method, "UpdateMemberRole"),
		strings.HasPrefix(method, "ListMembers"), strings.HasPrefix(method, "AcceptInvitation"):
		return "member"
	case strings.HasPrefix(method, "Project"), strings.Contains(method, "Project"):
		return "project"
	case strings.Contains(method, "Migration"), strings.Contains(method, "Rollback"):
		return "migration"
	case strings.Contains(method, "Schema"):
		return "schema"
	case strings.Contains(method, "Drift"):
		return "drift"
	default:
		return "system"
	}
}

func findFieldString(v any, names ...string) string {
	if v == nil {
		return ""
	}
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return ""
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return ""
	}
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.Type.Kind() != reflect.String {
			continue
		}
		for _, name := range names {
			if f.Name == name {
				s := rv.Field(i).String()
				if s != "" {
					return s
				}
			}
		}
	}
	return ""
}

func buildAuditEntry(ctx context.Context, fullMethod string, req, resp any) *domain.AuditEntry {
	parts := strings.Split(strings.TrimPrefix(fullMethod, "/"), "/")
	method := "unknown"
	if len(parts) >= 2 {
		method = parts[len(parts)-1]
	}
	service := "system"
	if len(parts) >= 2 {
		service = parts[len(parts)-2]
	}
	_ = service

	entry := &domain.AuditEntry{
		EventType:    method,
		Action:       strings.ToLower(method),
		ResourceType: auditResourceType(method),
		Metadata:     map[string]string{},
	}
	if userID, err := UserIDFromContext(ctx); err == nil {
		entry.ActorID = userID
	}
	if email, err := UserEmailFromContext(ctx); err == nil {
		entry.ActorEmail = email
	}

	if projectID := findFieldString(req, "ProjectId", "project_id"); projectID != "" {
		entry.Metadata["project_id"] = projectID
		entry.ResourceID = projectID
	}
	if resourceID := findFieldString(req, "Id", "ConnectionId", "MigrationId"); resourceID != "" {
		if entry.ResourceID == "" || entry.ResourceType == "connection" || entry.ResourceType == "migration" {
			entry.ResourceID = resourceID
		}
	}
	if resp != nil {
		if entry.ResourceID == "" {
			entry.ResourceID = findFieldString(resp, "Id", "ProjectId", "ConnectionId", "MigrationId", "RunId")
		}
		// prefer the concrete created resource id for creates
		for _, f := range []string{"Id", "ConnectionId", "MigrationId", "RunId", "InvitationId"} {
			if v := findFieldString(resp, f); v != "" {
				if f == "Id" && entry.ResourceType == "project" || f != "Id" {
					entry.ResourceID = v
					break
				}
			}
		}
	}
	return entry
}

// AuditInterceptor records every successful mutating RPC as an audit entry.
// Recording failures are logged via slog but never fail the RPC itself.
func AuditInterceptor(record func(ctx context.Context, entry *domain.AuditEntry) error, logf func(format string, args ...any)) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		if err != nil || record == nil || skipAuditMethod(info.FullMethod) {
			return resp, err
		}
		entry := buildAuditEntry(ctx, info.FullMethod, req, resp)
		if err := record(ctx, entry); err != nil && logf != nil {
			logf("audit record failed", "method", info.FullMethod, "error", err)
		}
		return resp, nil
	}
}
