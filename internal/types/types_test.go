package types

import "testing"

func TestRole_IsWriter(t *testing.T) {
	cases := []struct {
		role     Role
		isWriter bool
	}{
		{RoleFrontend, true},
		{RoleBackend, true},
		{RolePM, false},
		{RoleQA, false},
		{RoleCoordinator, false},
	}
	for _, c := range cases {
		if c.role.IsWriter() != c.isWriter {
			t.Errorf("Role(%s).IsWriter() = %v, want %v", c.role, c.role.IsWriter(), c.isWriter)
		}
	}
}
