package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func testApp(t *testing.T) *app {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err = db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		t.Fatal(err)
	}
	sqlBytes, err := os.ReadFile("../../migrations/001_init.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(string(sqlBytes)); err != nil {
		t.Fatal(err)
	}
	seed(db)
	hash, _ := bcrypt.GenerateFromPassword([]byte("Senha-de-teste-321!"), bcrypt.MinCost)
	if _, err = db.Exec("INSERT INTO users(name,email,password_hash,role) VALUES('Admin','admin@example.test',?,'ADMIN')", string(hash)); err != nil {
		t.Fatal(err)
	}
	return &app{db: db, secret: "secret-for-tests-with-more-than-thirty-two-bytes", issuer: "mare-viva", audience: "mare-viva-web"}
}
func TestPublicSpeciesEndpoint(t *testing.T) {
	a := testApp(t)
	w := httptest.NewRecorder()
	a.species(w, httptest.NewRequest("GET", "/api/species", nil))
	if w.Code != 200 {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	var body struct {
		Data []any `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || len(body.Data) < 8 {
		t.Fatalf("expected seeded species, got %d (%v)", len(body.Data), err)
	}
}
func TestLoginValidAndInvalid(t *testing.T) {
	a := testApp(t)
	bad := httptest.NewRecorder()
	a.login(bad, httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"email":"admin@example.test","password":"wrong"}`)))
	if bad.Code != 401 {
		t.Fatalf("invalid login status %d", bad.Code)
	}
	ok := httptest.NewRecorder()
	a.login(ok, httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"email":"admin@example.test","password":"Senha-de-teste-321!"}`)))
	if ok.Code != 200 {
		t.Fatalf("valid login status %d: %s", ok.Code, ok.Body)
	}
	if len(ok.Result().Cookies()) < 3 {
		t.Fatalf("expected access, refresh and csrf cookies")
	}
}
func TestJWTExpiredAndInvalid(t *testing.T) {
	a := testApp(t)
	next := func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }
	h := a.authenticated(next)
	for _, value := range []string{"not-a-jwt", func() string {
		c := claims{Role: "ADMIN", RegisteredClaims: jwt.RegisteredClaims{Issuer: a.issuer, Subject: "1", Audience: []string{a.audience}, ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute))}}
		tok, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString([]byte(a.secret))
		return tok
	}()} {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "mare_access", Value: value})
		w := httptest.NewRecorder()
		h(w, req)
		if w.Code != 401 {
			t.Errorf("expected 401 for invalid/expired JWT; got %d", w.Code)
		}
	}
}
func TestRefreshRotatesAndLogoutRevokes(t *testing.T) {
	a := testApp(t)
	login := httptest.NewRecorder()
	a.login(login, httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"email":"admin@example.test","password":"Senha-de-teste-321!"}`)))
	old := ""
	for _, c := range login.Result().Cookies() {
		if c.Name == "mare_refresh" {
			old = c.Value
		}
	}
	if old == "" {
		t.Fatal("refresh cookie missing")
	}
	req := httptest.NewRequest("POST", "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "mare_refresh", Value: old})
	w := httptest.NewRecorder()
	a.refresh(w, req)
	if w.Code != 200 {
		t.Fatalf("refresh status %d: %s", w.Code, w.Body)
	}
	newToken := ""
	for _, c := range w.Result().Cookies() {
		if c.Name == "mare_refresh" {
			newToken = c.Value
		}
	}
	if newToken == "" || newToken == old {
		t.Fatal("refresh token was not rotated")
	}
	req2 := httptest.NewRequest("POST", "/api/auth/logout", nil)
	req2.AddCookie(&http.Cookie{Name: "mare_refresh", Value: newToken})
	req2.AddCookie(&http.Cookie{Name: "mare_csrf", Value: "csrf"})
	req2.Header.Set("X-CSRF-Token", "csrf")
	out := httptest.NewRecorder()
	a.logout(out, req2)
	if out.Code != 200 {
		t.Fatalf("logout status %d", out.Code)
	}
	check := httptest.NewRecorder()
	refreshReq := httptest.NewRequest("POST", "/api/auth/refresh", nil)
	refreshReq.AddCookie(&http.Cookie{Name: "mare_refresh", Value: newToken})
	a.refresh(check, refreshReq)
	if check.Code != 401 {
		t.Fatalf("expected revoked refresh to fail, got %d", check.Code)
	}
}
func TestProtectedRouteAndRoleAuthorization(t *testing.T) {
	a := testApp(t)
	protected := a.authorized("ADMIN")(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	unauth := httptest.NewRecorder()
	protected(unauth, httptest.NewRequest("GET", "/", nil))
	if unauth.Code != 401 {
		t.Fatalf("expected 401, got %d", unauth.Code)
	}
	_, _ = a.db.Exec("INSERT INTO users(name,email,password_hash,role) VALUES('Editor','editor@example.test','unused','EDITOR')")
	token, _, err := a.issue(user{ID: 2, Role: "EDITOR"})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "mare_access", Value: token})
	denied := httptest.NewRecorder()
	protected(denied, req)
	if denied.Code != 403 {
		t.Fatalf("expected 403, got %d", denied.Code)
	}
}
func TestObservationPendingUntilModerated(t *testing.T) {
	a := testApp(t)
	req := httptest.NewRequest("POST", "/api/observations", strings.NewReader(`{"state":"PE","municipality":"Recife","date":"2026-01-02","category":"AVISTAMENTO","species":"","description":"Observação feita da praia, mantendo distância."}`))
	w := httptest.NewRecorder()
	a.createObservation(w, req)
	if w.Code != 201 {
		t.Fatalf("create status %d: %s", w.Code, w.Body)
	}
	var count int
	_ = a.db.QueryRow("SELECT COUNT(*) FROM observations WHERE status='PENDING'").Scan(&count)
	if count != 1 {
		t.Fatalf("expected pending report, got %d", count)
	}
	pub := httptest.NewRecorder()
	a.publicObservations(pub, httptest.NewRequest("GET", "/api/observations", nil))
	if strings.Contains(pub.Body.String(), "Recife") {
		t.Fatal("pending report leaked to public API")
	}
}
func TestModerationAndSimulator(t *testing.T) {
	a := testApp(t)
	_, _ = a.db.Exec("INSERT INTO observations(state,municipality,occurred_on,category,description) VALUES('PE','Recife','2026-01-01','POLUICAO','Relato de teste suficientemente longo')")
	token, _, _ := a.issue(user{ID: 1, Role: "ADMIN"})
	req := httptest.NewRequest("PATCH", "/api/admin/observations/1/status", strings.NewReader(`{"status":"APPROVED"}`))
	// Exercise the handler with the real path value and server-side role middleware.
	req = httptest.NewRequest("PATCH", "/api/admin/observations/1/status", strings.NewReader(`{"status":"APPROVED"}`))
	req.SetPathValue("id", "1")
	req.AddCookie(&http.Cookie{Name: "mare_access", Value: token})
	req.AddCookie(&http.Cookie{Name: "mare_csrf", Value: "csrf"})
	req.Header.Set("X-CSRF-Token", "csrf")
	w := httptest.NewRecorder()
	a.authorized("ADMIN")(a.moderate)(w, req)
	if w.Code != 200 {
		t.Fatalf("moderation status %d: %s", w.Code, w.Body)
	}
	pub := httptest.NewRecorder()
	a.publicObservations(pub, httptest.NewRequest("GET", "/api/observations", nil))
	if !strings.Contains(pub.Body.String(), "Recife") {
		t.Fatal("approved citizen report did not appear publicly")
	}
	sim := httptest.NewRecorder()
	a.simulate(sim, httptest.NewRequest("POST", "/api/ecosystem/simulate", strings.NewReader(`{"sharks":40}`)))
	if sim.Code != 200 || !strings.Contains(sim.Body.String(), `"scientific_forecast":false`) {
		t.Fatalf("simulator response %d: %s", sim.Code, sim.Body)
	}
}

func TestSpeciesCRUD(t *testing.T) {
	a := testApp(t)
	token, _, _ := a.issue(user{ID: 1, Role: "ADMIN"})
	call := func(method, path, payload string, handler http.HandlerFunc) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(payload))
		req.AddCookie(&http.Cookie{Name: "mare_access", Value: token})
		req.AddCookie(&http.Cookie{Name: "mare_csrf", Value: "csrf"})
		req.Header.Set("X-CSRF-Token", "csrf")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		return w
	}
	created := call("POST", "/api/admin/species", `{"common_name":"Tubarão de teste","scientific_name":"Prionace testus","published":true}`, a.authorized("ADMIN")(a.saveSpecies))
	if created.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", created.Code, created.Body)
	}
	var idResponse struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &idResponse); err != nil || idResponse.ID == 0 {
		t.Fatalf("invalid create response: %s", created.Body)
	}
	updateReq := httptest.NewRequest("PUT", "/api/admin/species/"+strconv.FormatInt(idResponse.ID, 10), strings.NewReader(`{"common_name":"Tubarão teste atualizado","scientific_name":"Prionace testus","published":true}`))
	updateReq.SetPathValue("id", strconv.FormatInt(idResponse.ID, 10))
	updateReq.AddCookie(&http.Cookie{Name: "mare_access", Value: token})
	updateReq.AddCookie(&http.Cookie{Name: "mare_csrf", Value: "csrf"})
	updateReq.Header.Set("X-CSRF-Token", "csrf")
	updated := httptest.NewRecorder()
	a.authorized("ADMIN")(a.saveSpecies)(updated, updateReq)
	if updated.Code != 200 {
		t.Fatalf("update: %d %s", updated.Code, updated.Body)
	}
	archReq := httptest.NewRequest("DELETE", "/api/admin/species/"+strconv.FormatInt(idResponse.ID, 10), nil)
	archReq.SetPathValue("id", strconv.FormatInt(idResponse.ID, 10))
	archReq.AddCookie(&http.Cookie{Name: "mare_access", Value: token})
	archReq.AddCookie(&http.Cookie{Name: "mare_csrf", Value: "csrf"})
	archReq.Header.Set("X-CSRF-Token", "csrf")
	archived := httptest.NewRecorder()
	a.authorized("ADMIN")(a.deleteSpecies)(archived, archReq)
	if archived.Code != 200 {
		t.Fatalf("archive: %d %s", archived.Code, archived.Body)
	}
}
