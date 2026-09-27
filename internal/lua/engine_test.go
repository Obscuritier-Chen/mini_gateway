package lua

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRuleHandlesQueryAndAdminAccess(t *testing.T) {
	engine, err := New("../../rules/rule.lua")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name   string
		path   string
		token  string
		status int
	}{
		{"admin without token", "/admin/settings", "", http.StatusForbidden},
		{"admin with token and debug", "/admin/settings?debug=true", "Bearer secret123", http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.token != "" {
				req.Header.Set("Authorization", tc.token)
			}
			status, _, err := engine.ExecuteRule(req, "test-request")
			if err != nil {
				t.Fatal(err)
			}
			if status != tc.status {
				t.Fatalf("status = %d, want %d", status, tc.status)
			}
		})
	}
}
