package domain

import (
	"strings"
	"testing"
)

func TestGenerateSlug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "simple", in: "My Service", want: "my-service"},
		{name: "mixed case", in: "SchemaHub Core", want: "schemahub-core"},
		{name: "special chars", in: "User's DB (v2)!", want: "user-s-db-v2"},
		{name: "leading trailing spaces", in: "  padded  ", want: "padded"},
		{name: "multiple spaces", in: "a  b   c", want: "a-b-c"},
		{name: "already slug", in: "my-service", want: "my-service"},
		{name: "unicode", in: "कॉफ़ी data", want: "data"},
		{name: "numbers", in: "123 project", want: "123-project"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateSlug(tt.in)
			if got != tt.want {
				t.Errorf("GenerateSlug(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestGenerateSlug_AllEmptiesFallsBack(t *testing.T) {
	slug := GenerateSlug("!!!")
	if !strings.HasPrefix(slug, "project-") {
		t.Errorf("GenerateSlug(!!!)= %q, want project- prefix fallback", slug)
	}
}

func TestGenerateSlug_LongNameTruncated(t *testing.T) {
	slug := GenerateSlug(strings.Repeat("a", 500))
	if len(slug) > 200 {
		t.Errorf("slug length = %d, want <= 200", len(slug))
	}
}

func TestProjectRole_Permissions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		role ProjectRole
		want int
	}{
		{RoleOwner, 100},
		{RoleAdmin, 80},
		{RoleMember, 50},
		{RoleViewer, 10},
		{ProjectRole("unknown"), 0},
	}

	for _, tt := range tests {
		if got := tt.role.Permissions(); got != tt.want {
			t.Errorf("Permissions(%s) = %d, want %d", tt.role, got, tt.want)
		}
	}
}

func TestProjectRole_Checks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		role           ProjectRole
		canManage      bool
		canWrite       bool
		canConnections bool
	}{
		{RoleOwner, true, true, true},
		{RoleAdmin, true, true, true},
		{RoleMember, false, true, false},
		{RoleViewer, false, false, false},
	}

	for _, tt := range tests {
		if got := tt.role.CanManageMembers(); got != tt.canManage {
			t.Errorf("CanManageMembers(%s) = %v, want %v", tt.role, got, tt.canManage)
		}
		if got := tt.role.CanWrite(); got != tt.canWrite {
			t.Errorf("CanWrite(%s) = %v, want %v", tt.role, got, tt.canWrite)
		}
		if got := tt.role.CanManageConnections(); got != tt.canConnections {
			t.Errorf("CanManageConnections(%s) = %v, want %v", tt.role, got, tt.canConnections)
		}
	}
}

func TestValidateVisibility(t *testing.T) {
	t.Parallel()

	for _, valid := range []string{"private", "team", "public"} {
		if _, err := ValidateVisibility(valid); err != nil {
			t.Errorf("ValidateVisibility(%q) error = %v, want nil", valid, err)
		}
	}

	if _, err := ValidateVisibility("secret"); err == nil {
		t.Error("ValidateVisibility(secret) = nil error, want error")
	}
}

func TestValidateRole(t *testing.T) {
	t.Parallel()

	for _, valid := range []string{"owner", "admin", "member", "viewer"} {
		if _, err := ValidateRole(valid); err != nil {
			t.Errorf("ValidateRole(%q) error = %v, want nil", valid, err)
		}
	}

	if _, err := ValidateRole("superuser"); err == nil {
		t.Error("ValidateRole(superuser) = nil error, want error")
	}
}
