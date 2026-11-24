package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func adminRequest(a *app, handler http.HandlerFunc, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	parts := strings.Split(path, "/")
	req.SetPathValue("id", parts[len(parts)-1])
	req = contextWithUser(req, user{ID: 1, Name: "Admin", Email: "admin@example.test", Role: "ADMIN"})
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

func TestAdministrativeCollectionsCRUD(t *testing.T) {
	a := testApp(t)
	cases := []struct {
		name, body string
	}{
		{"indicators", `{"title":"Teste","value":1.5,"unit":"%","year":2026,"scope":"Brasil","taxonomic_group":"Tubarões"}`},
		{"institutions", `{"name":"Instituição teste","state":"PE","city":"Recife"}`},
		{"sources", `{"title":"Fonte teste","year":2026,"url":"https://example.org/source"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := a.adminCollection(tc.name)
			created := adminRequest(a, h, http.MethodPost, "/api/admin/"+tc.name, tc.body)
			if created.Code != http.StatusCreated {
				t.Fatalf("create returned %d: %s", created.Code, created.Body)
			}
			listed := adminRequest(a, h, http.MethodGet, "/api/admin/"+tc.name, "")
			if listed.Code != http.StatusOK || !strings.Contains(listed.Body.String(), "Teste") && !strings.Contains(listed.Body.String(), "Instituição teste") && !strings.Contains(listed.Body.String(), "Fonte teste") {
				t.Fatalf("record missing from list: %d %s", listed.Code, listed.Body)
			}
		})
	}
	quiz := adminRequest(a, a.adminQuiz, http.MethodPost, "/api/admin/quiz", `{"prompt":"Pergunta?","explanation":"Explicação","source_url":"https://example.org","published":true,"options":[{"label":"Certo","correct":true},{"label":"Errado","correct":false}]}`)
	if quiz.Code != http.StatusOK {
		t.Fatalf("quiz save returned %d: %s", quiz.Code, quiz.Body)
	}
	users := adminRequest(a, a.adminUsers, http.MethodDelete, "/api/admin/users/1", "")
	if users.Code != http.StatusConflict {
		t.Fatalf("last administrator deactivation returned %d", users.Code)
	}
}
