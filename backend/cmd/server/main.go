package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

type app struct {
	db                       *sql.DB
	secret, issuer, audience string
}
type user struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}
type claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

var loginAttempts = struct {
	sync.Mutex
	entries map[string][]time.Time
}{entries: map[string][]time.Time{}}

func main() {
	dbPath := env("DB_PATH", "./mare-viva.db")
	_ = os.MkdirAll(filepath.Dir(dbPath), 0755)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	if _, err = db.Exec("PRAGMA foreign_keys=ON; PRAGMA busy_timeout=5000"); err != nil {
		log.Fatal(err)
	}
	if err = migrate(db); err != nil {
		log.Fatal(err)
	}
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			log.Println("Migrations aplicadas")
		case "seed-science":
			if err = seedScience(db); err != nil {
				log.Fatal(err)
			}
			log.Println("Seed científico aplicado")
		case "seed-demo":
			if err = seedDemo(db); err != nil {
				log.Fatal(err)
			}
			log.Println("Seed demonstrativo aplicado")
		default:
			log.Fatalf("comando desconhecido: %s", os.Args[1])
		}
		return
	}
	if err = seedScience(db); err != nil {
		log.Fatal(err)
	}
	bootstrap(db)
	a := &app{db: db, secret: env("JWT_SECRET", ""), issuer: env("JWT_ISSUER", "mare-viva"), audience: env("JWT_AUDIENCE", "mare-viva-web")}
	if len(a.secret) < 32 {
		log.Fatal("JWT_SECRET precisa ter ao menos 32 caracteres")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) { jsonOut(w, 200, map[string]any{"status": "ok"}) })
	mux.HandleFunc("GET /api/species", a.species)
	mux.HandleFunc("GET /api/species/{id}", a.speciesDetail)
	mux.HandleFunc("GET /api/indicators", a.indicators)
	mux.HandleFunc("GET /api/institutions", a.institutions)
	mux.HandleFunc("GET /api/sources", a.sources)
	mux.HandleFunc("GET /api/observations", a.publicObservations)
	mux.HandleFunc("POST /api/observations", a.createObservation)
	mux.HandleFunc("POST /api/ecosystem/simulate", a.simulate)
	mux.HandleFunc("GET /api/quiz", a.quiz)
	mux.HandleFunc("POST /api/quiz/evaluate", a.evaluateQuiz)
	mux.HandleFunc("POST /api/auth/login", a.login)
	mux.HandleFunc("POST /api/auth/refresh", a.refresh)
	mux.HandleFunc("POST /api/auth/logout", a.logout)
	mux.HandleFunc("GET /api/auth/me", a.authenticated(a.me))
	mux.HandleFunc("GET /api/admin/dashboard", a.authorized("ADMIN", "EDITOR", "MODERATOR")(a.dashboard))
	mux.HandleFunc("GET /api/admin/species", a.authorized("ADMIN", "EDITOR")(a.species))
	mux.HandleFunc("POST /api/admin/species", a.authorized("ADMIN", "EDITOR")(a.saveSpecies))
	mux.HandleFunc("PUT /api/admin/species/{id}", a.authorized("ADMIN", "EDITOR")(a.saveSpecies))
	mux.HandleFunc("DELETE /api/admin/species/{id}", a.authorized("ADMIN", "EDITOR")(a.deleteSpecies))
	for _, resource := range []string{"indicators", "institutions", "sources"} {
		roles := []string{"ADMIN", "EDITOR"}
		mux.HandleFunc("GET /api/admin/"+resource, a.authorized(roles...)(a.adminCollection(resource)))
		mux.HandleFunc("POST /api/admin/"+resource, a.authorized(roles...)(a.adminCollection(resource)))
		mux.HandleFunc("PUT /api/admin/"+resource+"/{id}", a.authorized(roles...)(a.adminCollection(resource)))
		mux.HandleFunc("DELETE /api/admin/"+resource+"/{id}", a.authorized(roles...)(a.adminCollection(resource)))
	}
	mux.HandleFunc("GET /api/admin/quiz", a.authorized("ADMIN", "EDITOR")(a.adminQuiz))
	mux.HandleFunc("POST /api/admin/quiz", a.authorized("ADMIN", "EDITOR")(a.adminQuiz))
	mux.HandleFunc("PUT /api/admin/quiz/{id}", a.authorized("ADMIN", "EDITOR")(a.adminQuiz))
	mux.HandleFunc("DELETE /api/admin/quiz/{id}", a.authorized("ADMIN", "EDITOR")(a.adminQuiz))
	mux.HandleFunc("GET /api/admin/users", a.authorized("ADMIN")(a.adminUsers))
	mux.HandleFunc("POST /api/admin/users", a.authorized("ADMIN")(a.adminUsers))
	mux.HandleFunc("PUT /api/admin/users/{id}", a.authorized("ADMIN")(a.adminUsers))
	mux.HandleFunc("DELETE /api/admin/users/{id}", a.authorized("ADMIN")(a.adminUsers))
	mux.HandleFunc("GET /api/admin/observations", a.authorized("ADMIN", "MODERATOR")(a.adminObservations))
	mux.HandleFunc("GET /api/admin/observations/{id}", a.authorized("ADMIN", "MODERATOR")(a.observationDetail))
	mux.HandleFunc("PATCH /api/admin/observations/{id}/status", a.authorized("ADMIN", "MODERATOR")(a.moderate))
	mux.HandleFunc("GET /api/admin/audit-logs", a.authorized("ADMIN")(a.audit))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		if origin := os.Getenv("CORS_ORIGIN"); origin != "" && r.Header.Get("Origin") == origin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,X-CSRF-Token")
			w.WriteHeader(204)
			return
		}
		log.Printf("%s %s", r.Method, r.URL.Path)
		mux.ServeHTTP(w, r)
	})
	addr := env("ADDR", ":8080")
	log.Printf("Maré Viva API em %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func jsonOut(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func errOut(w http.ResponseWriter, status int, msg string) {
	jsonOut(w, status, map[string]string{"error": msg})
}
func readJSON(r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
	return json.NewDecoder(r.Body).Decode(v)
}
func bootstrap(db *sql.DB) {
	email, pw := os.Getenv("ADMIN_EMAIL"), os.Getenv("ADMIN_PASSWORD")
	if email == "" || pw == "" {
		return
	}
	var n int
	_ = db.QueryRow("SELECT COUNT(*) FROM users WHERE role='ADMIN' AND active=1").Scan(&n)
	if n == 0 {
		h, e := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
		if e != nil {
			log.Printf("bootstrap admin: %v", e)
			return
		}
		_, e = db.Exec("INSERT INTO users(name,email,password_hash,role) VALUES(?,?,?,'ADMIN')", env("ADMIN_NAME", "Administrador Maré Viva"), email, string(h))
		if e != nil {
			log.Printf("bootstrap admin: %v", e)
		} else {
			log.Printf("administrador inicial criado para %s", email)
		}
	}
}
func (a *app) species(w http.ResponseWriter, r *http.Request) {
	rows, e := a.db.Query("SELECT sp.id,common_name,scientific_name,image_url,description,physical_characteristics,size,diet,habitat,distribution,brazil_presence,northeast_presence,ecology,threats,brazil_status,iucn_status,COALESCE(national_assessment_year,0),COALESCE(iucn_assessment_year,0),source_note,COALESCE((SELECT url FROM scientific_sources src JOIN species_sources l ON l.source_id=src.id WHERE l.species_id=sp.id AND src.kind='Avaliação de espécie' LIMIT 1),''),COALESCE((SELECT url FROM scientific_sources src JOIN species_sources l ON l.source_id=src.id WHERE l.species_id=sp.id AND src.url LIKE '%1-667-2026%' LIMIT 1),''),COALESCE((SELECT url FROM scientific_sources src JOIN species_sources l ON l.source_id=src.id WHERE l.species_id=sp.id AND src.kind='Ficha taxonômica' LIMIT 1),'') FROM species sp WHERE published=1 AND archived=0 ORDER BY common_name")
	if e != nil {
		errOut(w, 500, "Falha ao consultar espécies")
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id int
		var vals [19]string
		var nationalYear, iucnYear int
		if rows.Scan(&id, &vals[0], &vals[1], &vals[2], &vals[3], &vals[4], &vals[5], &vals[6], &vals[7], &vals[8], &vals[9], &vals[10], &vals[11], &vals[12], &vals[13], &vals[14], &nationalYear, &iucnYear, &vals[15], &vals[16], &vals[17], &vals[18]) == nil {
			out = append(out, map[string]any{"id": id, "common_name": vals[0], "scientific_name": vals[1], "image_url": vals[2], "description": vals[3], "physical_characteristics": vals[4], "size": vals[5], "diet": vals[6], "habitat": vals[7], "distribution": vals[8], "brazil_presence": vals[9], "northeast_presence": vals[10], "ecology": vals[11], "threats": vals[12], "brazil_status": vals[13], "iucn_status": vals[14], "national_assessment_year": nationalYear, "iucn_assessment_year": iucnYear, "source_note": vals[15], "iucn_source_url": vals[16], "national_source_url": vals[17], "taxonomic_source_url": vals[18]})
		}
	}
	jsonOut(w, 200, map[string]any{"data": out})
}
func (a *app) speciesDetail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	numericID, parseErr := strconv.Atoi(id)
	if parseErr != nil || numericID < 1 {
		errOut(w, 404, "Espécie não encontrada")
		return
	}
	var name, scientific, image, description, physical, size, diet, habitat, distribution, brazilPresence, northeastPresence, ecology, threats, brazil, iucn string
	var nationalYear, iucnYear int
	var sourceNote string
	e := a.db.QueryRow("SELECT common_name,scientific_name,image_url,description,physical_characteristics,size,diet,habitat,distribution,brazil_presence,northeast_presence,ecology,threats,brazil_status,iucn_status,COALESCE(national_assessment_year,0),COALESCE(iucn_assessment_year,0),source_note FROM species WHERE id=? AND published=1 AND archived=0", id).Scan(&name, &scientific, &image, &description, &physical, &size, &diet, &habitat, &distribution, &brazilPresence, &northeastPresence, &ecology, &threats, &brazil, &iucn, &nationalYear, &iucnYear, &sourceNote)
	if e != nil {
		errOut(w, 404, "Espécie não encontrada")
		return
	}
	jsonOut(w, 200, map[string]any{"data": map[string]any{"id": numericID, "common_name": name, "scientific_name": scientific, "image_url": image, "description": description, "physical_characteristics": physical, "size": size, "diet": diet, "habitat": habitat, "distribution": distribution, "brazil_presence": brazilPresence, "northeast_presence": northeastPresence, "ecology": ecology, "threats": threats, "brazil_status": brazil, "iucn_status": iucn, "national_assessment_year": nationalYear, "iucn_assessment_year": iucnYear, "source_note": sourceNote}})
}
func (a *app) indicators(w http.ResponseWriter, r *http.Request) {
	rows, e := a.db.Query("SELECT i.id,i.title,i.value,i.unit,i.year,i.scope,i.description,s.title,s.url,i.taxonomic_group FROM indicators i LEFT JOIN scientific_sources s ON s.id=i.source_id WHERE i.published=1 ORDER BY i.year DESC")
	if e != nil {
		jsonOut(w, 200, map[string]any{"data": []any{}})
		return
	}
	defer rows.Close()
	out := []any{}
	for rows.Next() {
		var id, year int
		var title, unit, scope, desc, st, url, group string
		var val float64
		_ = rows.Scan(&id, &title, &val, &unit, &year, &scope, &desc, &st, &url, &group)
		out = append(out, map[string]any{"id": id, "title": title, "value": val, "unit": unit, "year": year, "scope": scope, "taxonomic_group": group, "description": desc, "source": st, "url": url})
	}
	jsonOut(w, 200, map[string]any{"data": out})
}
func (a *app) institutions(w http.ResponseWriter, r *http.Request) {
	rows, _ := a.db.Query("SELECT id,name,description,state,city,focus,website FROM institutions ORDER BY name")
	defer rows.Close()
	out := []any{}
	for rows.Next() {
		var id int
		var x [6]string
		_ = rows.Scan(&id, &x[0], &x[1], &x[2], &x[3], &x[4], &x[5])
		out = append(out, map[string]any{"id": id, "name": x[0], "description": x[1], "state": x[2], "city": x[3], "focus": x[4], "website": x[5]})
	}
	jsonOut(w, 200, map[string]any{"data": out})
}
func (a *app) sources(w http.ResponseWriter, r *http.Request) {
	rows, _ := a.db.Query("SELECT id,title,author,year,url,doi,kind,description FROM scientific_sources ORDER BY year DESC")
	defer rows.Close()
	out := []any{}
	for rows.Next() {
		var id, year int
		var x [6]string
		_ = rows.Scan(&id, &x[0], &x[1], &year, &x[2], &x[3], &x[4], &x[5])
		out = append(out, map[string]any{"id": id, "title": x[0], "author": x[1], "year": year, "url": x[2], "doi": x[3], "kind": x[4], "description": x[5]})
	}
	jsonOut(w, 200, map[string]any{"data": out})
}
func (a *app) publicObservations(w http.ResponseWriter, r *http.Request) {
	rows, _ := a.db.Query("SELECT id,state,municipality,occurred_on,category,species,description,created_at,is_demo FROM observations WHERE status='APPROVED' ORDER BY occurred_on DESC LIMIT 100")
	defer rows.Close()
	out := []any{}
	for rows.Next() {
		var id int
		var x [7]string
		var isDemo bool
		_ = rows.Scan(&id, &x[0], &x[1], &x[2], &x[3], &x[4], &x[5], &x[6], &isDemo)
		out = append(out, map[string]any{"id": id, "state": x[0], "municipality": x[1], "date": x[2], "category": x[3], "species": x[4], "description": x[5], "created_at": x[6], "citizen_report": true, "scientifically_validated": false, "demonstration": isDemo})
	}
	jsonOut(w, 200, map[string]any{"data": out})
}
func (a *app) createObservation(w http.ResponseWriter, r *http.Request) {
	var x struct {
		State        string `json:"state"`
		Municipality string `json:"municipality"`
		Date         string `json:"date"`
		Category     string `json:"category"`
		Species      string `json:"species"`
		Description  string `json:"description"`
	}
	if readJSON(r, &x) != nil || len(x.Description) < 10 || len(x.Description) > 2000 || len(x.State) != 2 || x.Municipality == "" || x.Date == "" {
		errOut(w, 400, "Dados inválidos. Informe UF, município, data e descrição (10 a 2.000 caracteres).")
		return
	}
	if _, e := time.Parse("2006-01-02", x.Date); e != nil {
		errOut(w, 400, "Data inválida")
		return
	}
	cats := map[string]bool{"AVISTAMENTO": true, "ANIMAL_MORTO": true, "CAPTURA_ACIDENTAL": true, "POLUICAO": true, "DEGRADACAO": true}
	if !cats[x.Category] {
		errOut(w, 400, "Categoria inválida")
		return
	}
	res, e := a.db.Exec("INSERT INTO observations(state,municipality,occurred_on,category,species,description) VALUES(?,?,?,?,?,?)", strings.ToUpper(x.State), x.Municipality, x.Date, x.Category, x.Species, x.Description)
	if e != nil {
		errOut(w, 500, "Não foi possível salvar o relato")
		return
	}
	id, _ := res.LastInsertId()
	jsonOut(w, 201, map[string]any{"data": map[string]any{"id": id, "status": "PENDING"}, "message": "Relato recebido para moderação; não é validação científica."})
}
func (a *app) simulate(w http.ResponseWriter, r *http.Request) {
	var x struct {
		Sharks float64 `json:"sharks"`
	}
	if readJSON(r, &x) != nil || x.Sharks < 0 || x.Sharks > 100 {
		errOut(w, 400, "Informe tubarões entre 0 e 100")
		return
	}
	s := x.Sharks / 100
	jsonOut(w, 200, map[string]any{"model": "conceitual-educativo", "scientific_forecast": false, "series": []any{map[string]any{"name": "Tubarões", "value": x.Sharks}, map[string]any{"name": "Peixes médios", "value": 100 - 35*s}, map[string]any{"name": "Peixes pequenos", "value": 50 + 20*s}, map[string]any{"name": "Saúde do habitat", "value": 45 + 30*s}}, "note": "Modelo simplificado para aprendizagem; não representa previsão científica nem todas as espécies."})
}
func (a *app) quiz(w http.ResponseWriter, r *http.Request) {
	rows, _ := a.db.Query("SELECT id,prompt,explanation,source_url FROM quiz_questions WHERE published=1 ORDER BY id")
	type question struct {
		id                          int
		prompt, explanation, source string
	}
	questions := []question{}
	for rows.Next() {
		var q question
		_ = rows.Scan(&q.id, &q.prompt, &q.explanation, &q.source)
		questions = append(questions, q)
	}
	_ = rows.Close()
	out := []any{}
	for _, q := range questions {
		or, _ := a.db.Query("SELECT id,label FROM quiz_options WHERE question_id=? ORDER BY id", q.id)
		opts := []any{}
		for or.Next() {
			var oid int
			var l string
			_ = or.Scan(&oid, &l)
			opts = append(opts, map[string]any{"id": oid, "label": l})
		}
		or.Close()
		out = append(out, map[string]any{"id": q.id, "prompt": q.prompt, "explanation": q.explanation, "source_url": q.source, "options": opts})
	}
	jsonOut(w, 200, map[string]any{"data": out})
}
func (a *app) evaluateQuiz(w http.ResponseWriter, r *http.Request) {
	var x struct {
		Answers map[string]int `json:"answers"`
	}
	if readJSON(r, &x) != nil {
		errOut(w, 400, "Respostas inválidas")
		return
	}
	score, total := 0, 0
	results := []any{}
	for qid, oid := range x.Answers {
		var correct int
		var explanation, source string
		e := a.db.QueryRow("SELECT o.correct,q.explanation,q.source_url FROM quiz_options o JOIN quiz_questions q ON q.id=o.question_id WHERE q.id=? AND o.id=? AND q.published=1", qid, oid).Scan(&correct, &explanation, &source)
		if e != nil {
			continue
		}
		total++
		if correct == 1 {
			score++
		}
		results = append(results, map[string]any{"question_id": qid, "correct": correct == 1, "explanation": explanation, "source_url": source})
	}
	jsonOut(w, 200, map[string]any{"score": score, "total": total, "results": results})
}
func (a *app) issue(u user) (string, string, error) {
	now := time.Now()
	c := claims{Role: u.Role, RegisteredClaims: jwt.RegisteredClaims{Issuer: a.issuer, Subject: strconv.Itoa(u.ID), Audience: []string{a.audience}, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute))}}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	access, e := t.SignedString([]byte(a.secret))
	if e != nil {
		return "", "", e
	}
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	refresh := hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(refresh))
	_, e = a.db.Exec("INSERT INTO refresh_tokens(user_id,token_hash,expires_at) VALUES(?,?,?)", u.ID, hex.EncodeToString(sum[:]), now.Add(7*24*time.Hour).Format(time.RFC3339))
	return access, refresh, e
}
func setCookies(w http.ResponseWriter, access, refresh string) {
	secure := os.Getenv("APP_ENV") == "production"
	http.SetCookie(w, &http.Cookie{Name: "mare_access", Value: access, Path: "/", HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: 900})
	http.SetCookie(w, &http.Cookie{Name: "mare_refresh", Value: refresh, Path: "/api/auth", HttpOnly: true, Secure: secure, SameSite: http.SameSiteStrictMode, MaxAge: 604800})
	http.SetCookie(w, &http.Cookie{Name: "mare_csrf", Value: csrfToken(), Path: "/", Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: 604800})
}
func csrfToken() string { b := make([]byte, 24); _, _ = rand.Read(b); return hex.EncodeToString(b) }
func (a *app) login(w http.ResponseWriter, r *http.Request) {
	loginAttempts.Lock()
	key := r.RemoteAddr
	now := time.Now()
	recent := loginAttempts.entries[key][:0]
	for _, at := range loginAttempts.entries[key] {
		if now.Sub(at) < 5*time.Minute {
			recent = append(recent, at)
		}
	}
	if len(recent) >= 10 {
		loginAttempts.entries[key] = recent
		loginAttempts.Unlock()
		errOut(w, http.StatusTooManyRequests, "Muitas tentativas. Aguarde alguns minutos.")
		return
	}
	loginAttempts.entries[key] = append(recent, now)
	loginAttempts.Unlock()
	var x struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if readJSON(r, &x) != nil {
		errOut(w, 400, "Credenciais inválidas")
		return
	}
	var u user
	var hash string
	var active int
	e := a.db.QueryRow("SELECT id,name,email,password_hash,role,active FROM users WHERE email=?", strings.ToLower(strings.TrimSpace(x.Email))).Scan(&u.ID, &u.Name, &u.Email, &hash, &u.Role, &active)
	if e != nil || active != 1 || bcrypt.CompareHashAndPassword([]byte(hash), []byte(x.Password)) != nil {
		errOut(w, 401, "E-mail ou senha inválidos")
		return
	}
	access, refresh, e := a.issue(u)
	if e != nil {
		errOut(w, 500, "Não foi possível iniciar sessão")
		return
	}
	setCookies(w, access, refresh)
	a.auditLog(u.ID, "LOGIN", "session", "", r)
	jsonOut(w, 200, map[string]any{"data": u})
}
func (a *app) refresh(w http.ResponseWriter, r *http.Request) {
	c, e := r.Cookie("mare_refresh")
	if e != nil {
		errOut(w, 401, "Sessão expirada")
		return
	}
	sum := sha256.Sum256([]byte(c.Value))
	hash := hex.EncodeToString(sum[:])
	tx, e := a.db.Begin()
	if e != nil {
		errOut(w, 500, "Falha de sessão")
		return
	}
	var uid int
	var exp string
	e = tx.QueryRow("SELECT user_id,expires_at FROM refresh_tokens WHERE token_hash=? AND revoked_at IS NULL", hash).Scan(&uid, &exp)
	t, _ := time.Parse(time.RFC3339, exp)
	if e != nil || time.Now().After(t) {
		_ = tx.Rollback()
		errOut(w, 401, "Sessão expirada")
		return
	}
	_, _ = tx.Exec("UPDATE refresh_tokens SET revoked_at=CURRENT_TIMESTAMP WHERE token_hash=?", hash)
	var u user
	e = tx.QueryRow("SELECT id,name,email,role FROM users WHERE id=? AND active=1", uid).Scan(&u.ID, &u.Name, &u.Email, &u.Role)
	if e != nil {
		_ = tx.Rollback()
		errOut(w, 401, "Conta indisponível")
		return
	}
	_, _ = tx.Exec("INSERT INTO refresh_tokens(user_id,token_hash,expires_at) VALUES(?,?,?)", uid, "rotating-placeholder", time.Now().Add(7*24*time.Hour).Format(time.RFC3339))
	_, _ = tx.Exec("DELETE FROM refresh_tokens WHERE token_hash='rotating-placeholder'")
	_ = tx.Commit()
	access, refresh, e := a.issue(u)
	if e != nil {
		errOut(w, 500, "Falha de sessão")
		return
	}
	setCookies(w, access, refresh)
	jsonOut(w, 200, map[string]any{"data": u})
}
func (a *app) logout(w http.ResponseWriter, r *http.Request) {
	csrf, _ := r.Cookie("mare_csrf")
	if csrf == nil || csrf.Value == "" || r.Header.Get("X-CSRF-Token") != csrf.Value {
		errOut(w, http.StatusForbidden, "Token CSRF inválido")
		return
	}
	if c, e := r.Cookie("mare_refresh"); e == nil {
		sum := sha256.Sum256([]byte(c.Value))
		_, _ = a.db.Exec("UPDATE refresh_tokens SET revoked_at=CURRENT_TIMESTAMP WHERE token_hash=?", hex.EncodeToString(sum[:]))
	}
	for _, n := range []string{"mare_access", "mare_refresh", "mare_csrf"} {
		path := "/"
		if n == "mare_refresh" {
			path = "/api/auth"
		}
		http.SetCookie(w, &http.Cookie{Name: n, Value: "", Path: path, MaxAge: -1, HttpOnly: n != "mare_csrf", SameSite: http.SameSiteLaxMode})
	}
	jsonOut(w, 200, map[string]any{"message": "Sessão encerrada"})
}
func (a *app) authenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, e := r.Cookie("mare_access")
		if e != nil {
			errOut(w, 401, "Autenticação necessária")
			return
		}
		token, e := jwt.ParseWithClaims(cookie.Value, &claims{}, func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, io.ErrUnexpectedEOF
			}
			return []byte(a.secret), nil
		}, jwt.WithIssuer(a.issuer), jwt.WithAudience(a.audience))
		if e != nil || !token.Valid {
			errOut(w, 401, "Sessão inválida ou expirada")
			return
		}
		c := token.Claims.(*claims)
		uid, _ := strconv.Atoi(c.Subject)
		var u user
		var active int
		if a.db.QueryRow("SELECT id,name,email,role,active FROM users WHERE id=?", uid).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &active) != nil || active != 1 {
			errOut(w, 401, "Conta indisponível")
			return
		}
		if r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS" {
			csrf, _ := r.Cookie("mare_csrf")
			if csrf == nil || csrf.Value == "" || r.Header.Get("X-CSRF-Token") != csrf.Value {
				errOut(w, 403, "Token CSRF inválido")
				return
			}
		}
		r = r.WithContext(context.WithValue(r.Context(), ctxKey("user"), u))
		next(w, r)
	}
}

// Request context helpers keep the authentication data server-side.
type ctxKey string

func contextWithUser(r *http.Request, u user) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), ctxKey("user"), u))
}
func current(r *http.Request) user { u, _ := r.Context().Value(ctxKey("user")).(user); return u }
func (a *app) authorized(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return a.authenticated(func(w http.ResponseWriter, r *http.Request) {
			u := current(r)
			for _, role := range roles {
				if u.Role == role {
					next(w, r)
					return
				}
			}
			errOut(w, 403, "Permissão insuficiente")
		})
	}
}
func (a *app) me(w http.ResponseWriter, r *http.Request) {
	jsonOut(w, 200, map[string]any{"data": current(r)})
}
func (a *app) dashboard(w http.ResponseWriter, r *http.Request) {
	out := map[string]int{}
	for k, q := range map[string]string{"species": "SELECT COUNT(*) FROM species WHERE archived=0", "threatened_species": "SELECT COUNT(*) FROM species WHERE archived=0 AND (iucn_status LIKE '%Endangered%' OR iucn_status LIKE '%Vulnerable%' OR iucn_status LIKE '%En perigo%' OR iucn_status LIKE '%Vulnerável%')", "observations": "SELECT COUNT(*) FROM observations", "demonstration_observations": "SELECT COUNT(*) FROM observations WHERE is_demo=1", "community_reports": "SELECT COUNT(*) FROM observations WHERE is_demo=0", "pending": "SELECT COUNT(*) FROM observations WHERE status='PENDING'", "approved": "SELECT COUNT(*) FROM observations WHERE status='APPROVED'", "institutions": "SELECT COUNT(*) FROM institutions", "sources": "SELECT COUNT(*) FROM scientific_sources"} {
		var count int
		_ = a.db.QueryRow(q).Scan(&count)
		out[k] = count
	}
	rows, _ := a.db.Query("SELECT state,COUNT(*),SUM(is_demo) FROM observations GROUP BY state ORDER BY COUNT(*) DESC")
	defer rows.Close()
	states := []any{}
	for rows.Next() {
		var s string
		var n, demo int
		_ = rows.Scan(&s, &n, &demo)
		states = append(states, map[string]any{"state": s, "count": n, "demonstration_count": demo, "community_count": n - demo})
	}
	outAny := map[string]any{"counts": out, "observations_by_state": states}
	jsonOut(w, 200, map[string]any{"data": outAny})
}
func (a *app) saveSpecies(w http.ResponseWriter, r *http.Request) {
	var x struct {
		Common            string `json:"common_name"`
		Scientific        string `json:"scientific_name"`
		Image             string `json:"image_url"`
		Description       string `json:"description"`
		Physical          string `json:"physical_characteristics"`
		Size              string `json:"size"`
		Diet              string `json:"diet"`
		Habitat           string `json:"habitat"`
		Distribution      string `json:"distribution"`
		BrazilPresence    string `json:"brazil_presence"`
		NortheastPresence string `json:"northeast_presence"`
		Ecology           string `json:"ecology"`
		Threats           string `json:"threats"`
		Brazil            string `json:"brazil_status"`
		IUCN              string `json:"iucn_status"`
		Published         bool   `json:"published"`
	}
	if readJSON(r, &x) != nil || x.Common == "" || x.Scientific == "" {
		errOut(w, 400, "Nome popular e científico são obrigatórios")
		return
	}
	u := current(r)
	if r.Method == "POST" {
		res, e := a.db.Exec("INSERT INTO species(common_name,scientific_name,image_url,description,physical_characteristics,size,diet,habitat,distribution,brazil_presence,northeast_presence,ecology,threats,brazil_status,iucn_status,published) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", x.Common, x.Scientific, x.Image, x.Description, x.Physical, x.Size, x.Diet, x.Habitat, x.Distribution, x.BrazilPresence, x.NortheastPresence, x.Ecology, x.Threats, x.Brazil, x.IUCN, boolInt(x.Published))
		if e != nil {
			errOut(w, 409, "Espécie duplicada ou dados inválidos")
			return
		}
		id, _ := res.LastInsertId()
		a.auditLog(u.ID, "CREATE", "species", strconv.FormatInt(id, 10), r)
		jsonOut(w, 201, map[string]any{"id": id})
		return
	}
	id := r.PathValue("id")
	res, e := a.db.Exec("UPDATE species SET common_name=?,scientific_name=?,image_url=?,description=?,physical_characteristics=?,size=?,diet=?,habitat=?,distribution=?,brazil_presence=?,northeast_presence=?,ecology=?,threats=?,brazil_status=?,iucn_status=?,published=?,updated_at=CURRENT_TIMESTAMP WHERE id=?", x.Common, x.Scientific, x.Image, x.Description, x.Physical, x.Size, x.Diet, x.Habitat, x.Distribution, x.BrazilPresence, x.NortheastPresence, x.Ecology, x.Threats, x.Brazil, x.IUCN, boolInt(x.Published), id)
	if e != nil {
		errOut(w, 400, "Falha ao atualizar espécie")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		errOut(w, 404, "Espécie não encontrada")
		return
	}
	a.auditLog(u.ID, "UPDATE", "species", id, r)
	jsonOut(w, 200, map[string]bool{"updated": true})
}
func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
func (a *app) deleteSpecies(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	res, e := a.db.Exec("UPDATE species SET archived=1,updated_at=CURRENT_TIMESTAMP WHERE id=?", id)
	if e != nil {
		errOut(w, 500, "Falha ao arquivar")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		errOut(w, 404, "Espécie não encontrada")
		return
	}
	u := current(r)
	a.auditLog(u.ID, "ARCHIVE", "species", id, r)
	jsonOut(w, 200, map[string]bool{"archived": true})
}
func (a *app) adminObservations(w http.ResponseWriter, r *http.Request) {
	q := "SELECT id,state,municipality,occurred_on,category,species,description,status,internal_note,rejection_reason,created_at,is_demo FROM observations WHERE 1=1"
	args := []any{}
	for _, key := range []string{"state", "status", "category", "municipality"} {
		if v := r.URL.Query().Get(key); v != "" {
			col := map[string]string{"state": "state", "status": "status", "category": "category", "municipality": "municipality"}[key]
			if key == "municipality" {
				q += " AND municipality LIKE ?"
				args = append(args, "%"+v+"%")
			} else {
				q += " AND " + col + "=?"
				args = append(args, v)
			}
		}
	}
	q += " ORDER BY created_at DESC LIMIT 200"
	rows, e := a.db.Query(q, args...)
	if e != nil {
		errOut(w, 500, "Falha ao listar ocorrências")
		return
	}
	defer rows.Close()
	out := []any{}
	for rows.Next() {
		var id int
		var x [10]string
		var isDemo bool
		_ = rows.Scan(&id, &x[0], &x[1], &x[2], &x[3], &x[4], &x[5], &x[6], &x[7], &x[8], &x[9], &isDemo)
		out = append(out, map[string]any{"id": id, "state": x[0], "municipality": x[1], "date": x[2], "category": x[3], "species": x[4], "description": x[5], "status": x[6], "internal_note": x[7], "rejection_reason": x[8], "created_at": x[9], "demonstration": isDemo})
	}
	jsonOut(w, 200, map[string]any{"data": out})
}
func (a *app) observationDetail(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	rows, e := a.db.Query("SELECT id,state,municipality,occurred_on,category,species,description,status,internal_note,rejection_reason,created_at FROM observations WHERE id=?", id)
	if e != nil {
		errOut(w, 500, "Falha")
		return
	}
	defer rows.Close()
	if !rows.Next() {
		errOut(w, 404, "Ocorrência não encontrada")
		return
	}
	var n int
	var x [10]string
	_ = rows.Scan(&n, &x[0], &x[1], &x[2], &x[3], &x[4], &x[5], &x[6], &x[7], &x[8], &x[9])
	jsonOut(w, 200, map[string]any{"data": map[string]any{"id": n, "state": x[0], "municipality": x[1], "date": x[2], "category": x[3], "species": x[4], "description": x[5], "status": x[6], "internal_note": x[7], "rejection_reason": x[8], "created_at": x[9]}})
}
func (a *app) moderate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var x struct {
		Status string `json:"status"`
		Note   string `json:"internal_note"`
		Reason string `json:"rejection_reason"`
	}
	if readJSON(r, &x) != nil {
		errOut(w, 400, "Dados inválidos")
		return
	}
	if x.Status != "UNDER_REVIEW" && x.Status != "APPROVED" && x.Status != "REJECTED" && x.Status != "PENDING" {
		errOut(w, 400, "Status inválido")
		return
	}
	if x.Status == "REJECTED" && strings.TrimSpace(x.Reason) == "" {
		errOut(w, 400, "Informe o motivo da rejeição")
		return
	}
	tx, e := a.db.Begin()
	if e != nil {
		errOut(w, 500, "Falha na transação")
		return
	}
	var old string
	if e = tx.QueryRow("SELECT status FROM observations WHERE id=?", id).Scan(&old); e != nil {
		_ = tx.Rollback()
		errOut(w, 404, "Ocorrência não encontrada")
		return
	}
	_, e = tx.Exec("UPDATE observations SET status=?,internal_note=?,rejection_reason=?,updated_at=CURRENT_TIMESTAMP WHERE id=?", x.Status, x.Note, x.Reason, id)
	if e == nil {
		u := current(r)
		_, e = tx.Exec("INSERT INTO observation_history(observation_id,actor_id,old_status,new_status,note) VALUES(?,?,?,?,?)", id, u.ID, old, x.Status, x.Note)
	}
	if e != nil {
		_ = tx.Rollback()
		errOut(w, 500, "Falha ao moderar ocorrência")
		return
	}
	_ = tx.Commit()
	u := current(r)
	a.auditLog(u.ID, "OBSERVATION_"+x.Status, "observation", id, r)
	jsonOut(w, 200, map[string]any{"status": x.Status, "scientifically_validated": false})
}
func (a *app) auditLog(uid int, action, entity, id string, r *http.Request) {
	_, _ = a.db.Exec("INSERT INTO audit_logs(user_id,action,entity,entity_id) VALUES(?,?,?,?)", uid, action, entity, id)
}
func (a *app) audit(w http.ResponseWriter, r *http.Request) {
	rows, _ := a.db.Query("SELECT id,user_id,action,entity,entity_id,created_at FROM audit_logs ORDER BY id DESC LIMIT 100")
	defer rows.Close()
	out := []any{}
	for rows.Next() {
		var id, uid int
		var action, entity, eid, at string
		_ = rows.Scan(&id, &uid, &action, &entity, &eid, &at)
		out = append(out, map[string]any{"id": id, "user_id": uid, "action": action, "entity": entity, "entity_id": eid, "created_at": at})
	}
	jsonOut(w, 200, map[string]any{"data": out})
}
