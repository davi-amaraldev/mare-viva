package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type adminField struct {
	Name     string
	Required bool
}
type adminCollection struct {
	Table  string
	Fields []adminField
}

var adminCollections = map[string]adminCollection{
	"indicators":   {"indicators", []adminField{{"title", true}, {"value", true}, {"unit", true}, {"year", true}, {"scope", true}, {"description", false}, {"source_id", false}, {"published", false}, {"taxonomic_group", false}, {"source_url", false}}},
	"institutions": {"institutions", []adminField{{"name", true}, {"description", false}, {"state", false}, {"city", false}, {"focus", false}, {"website", false}, {"image_url", false}}},
	"sources":      {"scientific_sources", []adminField{{"title", true}, {"author", false}, {"year", true}, {"url", true}, {"doi", false}, {"kind", false}, {"description", false}}},
}

func (a *app) adminCollection(kind string) http.HandlerFunc {
	c, ok := adminCollections[kind]
	if !ok {
		return func(w http.ResponseWriter, r *http.Request) { errOut(w, 404, "Coleção não encontrada") }
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		switch r.Method {
		case http.MethodGet:
			cols := []string{"id"}
			for _, f := range c.Fields {
				cols = append(cols, f.Name)
			}
			rows, e := a.db.Query("SELECT " + strings.Join(cols, ",") + " FROM " + c.Table + " ORDER BY id DESC LIMIT 500")
			if e != nil {
				errOut(w, 500, "Falha ao listar registros")
				return
			}
			defer rows.Close()
			out := []map[string]any{}
			for rows.Next() {
				vals := make([]any, len(cols))
				ptr := make([]any, len(cols))
				for i := range vals {
					ptr[i] = &vals[i]
				}
				if rows.Scan(ptr...) != nil {
					continue
				}
				item := map[string]any{}
				for i, name := range cols {
					if b, ok := vals[i].([]byte); ok {
						vals[i] = string(b)
					}
					item[name] = vals[i]
				}
				out = append(out, item)
			}
			jsonOut(w, 200, map[string]any{"data": out})
		case http.MethodPost, http.MethodPut:
			var raw map[string]any
			if readJSON(r, &raw) != nil {
				errOut(w, 400, "JSON inválido")
				return
			}
			values := make([]any, 0, len(c.Fields))
			columns := make([]string, 0, len(c.Fields))
			for _, f := range c.Fields {
				v, exists := raw[f.Name]
				if f.Required && (!exists || strings.TrimSpace(fmt.Sprint(v)) == "") {
					errOut(w, 400, "Campo obrigatório: "+f.Name)
					return
				}
				if !exists {
					switch f.Name {
					case "published":
						v = 1
					default:
						v = ""
					}
				}
				if f.Name == "published" {
					if b, ok := v.(bool); ok {
						v = boolInt(b)
					}
				}
				if f.Name == "source_id" && fmt.Sprint(v) == "" {
					v = nil
				}
				columns = append(columns, f.Name)
				values = append(values, v)
			}
			if r.Method == http.MethodPost {
				marks := make([]string, len(columns))
				for i := range marks {
					marks[i] = "?"
				}
				res, e := a.db.Exec("INSERT INTO "+c.Table+"("+strings.Join(columns, ",")+") VALUES("+strings.Join(marks, ",")+")", values...)
				if e != nil {
					errOut(w, 400, "Não foi possível criar o registro")
					return
				}
				newID, _ := res.LastInsertId()
				u := current(r)
				a.auditLog(u.ID, "CREATE", kind, strconv.FormatInt(newID, 10), r)
				jsonOut(w, 201, map[string]any{"id": newID})
				return
			}
			idNum, e := strconv.Atoi(id)
			if e != nil || idNum < 1 {
				errOut(w, 400, "Identificador inválido")
				return
			}
			sets := make([]string, len(columns))
			for i, name := range columns {
				sets[i] = name + "=?"
			}
			values = append(values, idNum)
			res, e := a.db.Exec("UPDATE "+c.Table+" SET "+strings.Join(sets, ",")+" WHERE id=?", values...)
			if e != nil {
				errOut(w, 400, "Não foi possível atualizar o registro")
				return
			}
			n, _ := res.RowsAffected()
			if n == 0 {
				errOut(w, 404, "Registro não encontrado")
				return
			}
			u := current(r)
			a.auditLog(u.ID, "UPDATE", kind, id, r)
			jsonOut(w, 200, map[string]bool{"updated": true})
		case http.MethodDelete:
			idNum, e := strconv.Atoi(id)
			if e != nil || idNum < 1 {
				errOut(w, 400, "Identificador inválido")
				return
			}
			res, e := a.db.Exec("DELETE FROM "+c.Table+" WHERE id=?", idNum)
			if e != nil {
				errOut(w, 409, "Registro relacionado não pode ser removido")
				return
			}
			n, _ := res.RowsAffected()
			if n == 0 {
				errOut(w, 404, "Registro não encontrado")
				return
			}
			u := current(r)
			a.auditLog(u.ID, "DELETE", kind, id, r)
			jsonOut(w, 200, map[string]bool{"deleted": true})
		default:
			errOut(w, 405, "Método não permitido")
		}
	}
}

func (a *app) adminQuiz(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if r.Method == http.MethodGet {
		rows, e := a.db.Query("SELECT id,prompt,explanation,source_url,published FROM quiz_questions ORDER BY id DESC LIMIT 500")
		if e != nil {
			errOut(w, 500, "Falha ao listar quiz")
			return
		}
		type qrow struct {
			id                          int
			prompt, explanation, source string
			published                   int
		}
		qs := []qrow{}
		for rows.Next() {
			var q qrow
			_ = rows.Scan(&q.id, &q.prompt, &q.explanation, &q.source, &q.published)
			qs = append(qs, q)
		}
		_ = rows.Close()
		out := []any{}
		for _, q := range qs {
			or, e := a.db.Query("SELECT id,label,correct FROM quiz_options WHERE question_id=? ORDER BY id", q.id)
			if e != nil {
				continue
			}
			opts := []any{}
			for or.Next() {
				var oid, correct int
				var label string
				_ = or.Scan(&oid, &label, &correct)
				opts = append(opts, map[string]any{"id": oid, "label": label, "correct": correct == 1})
			}
			_ = or.Close()
			out = append(out, map[string]any{"id": q.id, "prompt": q.prompt, "explanation": q.explanation, "source_url": q.source, "published": q.published == 1, "options": opts})
		}
		jsonOut(w, 200, map[string]any{"data": out})
		return
	}
	if r.Method == http.MethodDelete {
		n, e := strconv.Atoi(id)
		if e != nil {
			errOut(w, 400, "Identificador inválido")
			return
		}
		res, e := a.db.Exec("UPDATE quiz_questions SET published=0 WHERE id=?", n)
		if e != nil {
			errOut(w, 500, "Falha ao arquivar questão")
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			errOut(w, 404, "Questão não encontrada")
			return
		}
		u := current(r)
		a.auditLog(u.ID, "ARCHIVE", "quiz", id, r)
		jsonOut(w, 200, map[string]bool{"archived": true})
		return
	}
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		errOut(w, 405, "Método não permitido")
		return
	}
	var x struct {
		Prompt      string `json:"prompt"`
		Explanation string `json:"explanation"`
		SourceURL   string `json:"source_url"`
		Published   bool   `json:"published"`
		Options     []struct {
			Label   string `json:"label"`
			Correct bool   `json:"correct"`
		} `json:"options"`
	}
	if readJSON(r, &x) != nil || strings.TrimSpace(x.Prompt) == "" || strings.TrimSpace(x.Explanation) == "" || len(x.Options) < 2 {
		errOut(w, 400, "Informe pergunta, explicação e ao menos duas alternativas")
		return
	}
	correct := 0
	for _, o := range x.Options {
		if strings.TrimSpace(o.Label) == "" {
			errOut(w, 400, "Alternativa vazia")
			return
		}
		if o.Correct {
			correct++
		}
	}
	if correct != 1 {
		errOut(w, 400, "Defina exatamente uma alternativa correta")
		return
	}
	tx, e := a.db.Begin()
	if e != nil {
		errOut(w, 500, "Falha ao salvar questão")
		return
	}
	defer tx.Rollback()
	var qid int64
	if r.Method == http.MethodPost {
		res, e := tx.Exec("INSERT INTO quiz_questions(prompt,explanation,source_url,published) VALUES(?,?,?,?)", x.Prompt, x.Explanation, x.SourceURL, boolInt(x.Published))
		if e != nil {
			errOut(w, 400, "Falha ao criar questão")
			return
		}
		qid, _ = res.LastInsertId()
	} else {
		n, e := strconv.Atoi(id)
		if e != nil {
			errOut(w, 400, "Identificador inválido")
			return
		}
		qid = int64(n)
		res, e := tx.Exec("UPDATE quiz_questions SET prompt=?,explanation=?,source_url=?,published=? WHERE id=?", x.Prompt, x.Explanation, x.SourceURL, boolInt(x.Published), qid)
		if e != nil {
			errOut(w, 400, "Falha ao atualizar questão")
			return
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			errOut(w, 404, "Questão não encontrada")
			return
		}
		if _, e = tx.Exec("DELETE FROM quiz_options WHERE question_id=?", qid); e != nil {
			errOut(w, 500, "Falha ao atualizar alternativas")
			return
		}
	}
	for _, o := range x.Options {
		if _, e = tx.Exec("INSERT INTO quiz_options(question_id,label,correct) VALUES(?,?,?)", qid, o.Label, boolInt(o.Correct)); e != nil {
			errOut(w, 500, "Falha ao salvar alternativas")
			return
		}
	}
	if e = tx.Commit(); e != nil {
		errOut(w, 500, "Falha ao concluir transação")
		return
	}
	u := current(r)
	action := "CREATE"
	if r.Method == http.MethodPut {
		action = "UPDATE"
	}
	a.auditLog(u.ID, action, "quiz", strconv.FormatInt(qid, 10), r)
	jsonOut(w, 200, map[string]any{"id": qid})
}

func (a *app) adminUsers(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if r.Method == http.MethodGet {
		rows, e := a.db.Query("SELECT id,name,email,role,active,created_at FROM users ORDER BY name")
		if e != nil {
			errOut(w, 500, "Falha ao listar usuários")
			return
		}
		defer rows.Close()
		out := []any{}
		for rows.Next() {
			var uid, active int
			var name, email, role, created string
			_ = rows.Scan(&uid, &name, &email, &role, &active, &created)
			out = append(out, map[string]any{"id": uid, "name": name, "email": email, "role": role, "active": active == 1, "created_at": created})
		}
		jsonOut(w, 200, map[string]any{"data": out})
		return
	}
	if r.Method == http.MethodDelete {
		uid, e := strconv.Atoi(id)
		if e != nil {
			errOut(w, 400, "Identificador inválido")
			return
		}
		if !a.canDeactivateAdmin(uid) {
			errOut(w, 409, "Não é possível desativar o último administrador ativo")
			return
		}
		tx, e := a.db.Begin()
		if e != nil {
			errOut(w, 500, "Falha ao desativar usuário")
			return
		}
		defer tx.Rollback()
		if _, e = tx.Exec("UPDATE users SET active=0 WHERE id=?", uid); e != nil {
			errOut(w, 500, "Falha ao desativar usuário")
			return
		}
		_, _ = tx.Exec("DELETE FROM refresh_tokens WHERE user_id=?", uid)
		if e = tx.Commit(); e != nil {
			errOut(w, 500, "Falha ao concluir operação")
			return
		}
		u := current(r)
		a.auditLog(u.ID, "DEACTIVATE", "user", id, r)
		jsonOut(w, 200, map[string]bool{"deactivated": true})
		return
	}
	var x struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
		Active   bool   `json:"active"`
	}
	if readJSON(r, &x) != nil || strings.TrimSpace(x.Name) == "" || strings.TrimSpace(x.Email) == "" || !validRole(x.Role) {
		errOut(w, 400, "Informe nome, e-mail e papel válido")
		return
	}
	if r.Method == http.MethodPost {
		if len(x.Password) < 12 {
			errOut(w, 400, "A senha inicial precisa ter ao menos 12 caracteres")
			return
		}
		hash, e := bcrypt.GenerateFromPassword([]byte(x.Password), bcrypt.DefaultCost)
		if e != nil {
			errOut(w, 500, "Falha ao proteger senha")
			return
		}
		res, e := a.db.Exec("INSERT INTO users(name,email,password_hash,role,active) VALUES(?,?,?,?,?)", x.Name, strings.ToLower(x.Email), string(hash), x.Role, boolInt(x.Active))
		if e != nil {
			errOut(w, 409, "E-mail já cadastrado ou dados inválidos")
			return
		}
		uid, _ := res.LastInsertId()
		u := current(r)
		a.auditLog(u.ID, "CREATE", "user", strconv.FormatInt(uid, 10), r)
		jsonOut(w, 201, map[string]any{"id": uid})
		return
	}
	uid, e := strconv.Atoi(id)
	if e != nil {
		errOut(w, 400, "Identificador inválido")
		return
	}
	var oldRole string
	var active int
	if e = a.db.QueryRow("SELECT role,active FROM users WHERE id=?", uid).Scan(&oldRole, &active); e != nil {
		errOut(w, 404, "Usuário não encontrado")
		return
	}
	if oldRole == "ADMIN" && (x.Role != "ADMIN" || !x.Active) && !a.canDeactivateAdmin(uid) {
		errOut(w, 409, "Não é possível alterar o último administrador ativo")
		return
	}
	if x.Password != "" && len(x.Password) < 12 {
		errOut(w, 400, "A nova senha precisa ter ao menos 12 caracteres")
		return
	}
	if x.Password != "" {
		hash, e := bcrypt.GenerateFromPassword([]byte(x.Password), bcrypt.DefaultCost)
		if e != nil {
			errOut(w, 500, "Falha ao proteger senha")
			return
		}
		_, e = a.db.Exec("UPDATE users SET name=?,email=?,role=?,active=?,password_hash=? WHERE id=?", x.Name, strings.ToLower(x.Email), x.Role, boolInt(x.Active), string(hash), uid)
		if e != nil {
			errOut(w, 409, "E-mail já cadastrado ou dados inválidos")
			return
		}
		_, _ = a.db.Exec("DELETE FROM refresh_tokens WHERE user_id=?", uid)
	} else {
		_, e = a.db.Exec("UPDATE users SET name=?,email=?,role=?,active=? WHERE id=?", x.Name, strings.ToLower(x.Email), x.Role, boolInt(x.Active), uid)
		if e != nil {
			errOut(w, 409, "E-mail já cadastrado ou dados inválidos")
			return
		}
		if !x.Active || oldRole != x.Role {
			_, _ = a.db.Exec("DELETE FROM refresh_tokens WHERE user_id=?", uid)
		}
	}
	u := current(r)
	a.auditLog(u.ID, "UPDATE", "user", id, r)
	jsonOut(w, 200, map[string]bool{"updated": true})
}
func validRole(role string) bool { return role == "ADMIN" || role == "EDITOR" || role == "MODERATOR" }
func (a *app) canDeactivateAdmin(id int) bool {
	var n int
	if a.db.QueryRow("SELECT COUNT(*) FROM users WHERE role='ADMIN' AND active=1 AND id<>?", id).Scan(&n) != nil {
		return false
	}
	return n > 0
}
