package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
)

type seedSource struct {
	Key, Title, Author          string
	Year                        int
	URL, DOI, Kind, Description string
}
type seedSpecies struct {
	Key, Common, Scientific, Image, Description, Physical, Size, Diet, Habitat, Distribution, BrazilPresence, NortheastPresence, Ecology, Threats, BrazilStatus, IUCNStatus, SourceKey string
	NationalYear, IUCNYear                                                                                                                                                             int
}
type seedIndicator struct {
	Key, Title                           string
	Value                                float64
	Unit                                 string
	Year                                 int
	Scope, Group, Description, SourceKey string
}
type seedInstitution struct{ Key, Name, Description, State, City, Focus, Website string }
type seedQuestion struct {
	Key, Prompt, Explanation, SourceURL string
	Options                             []string
	Correct                             int
}
type scienceSeed struct {
	Sources      []seedSource
	Species      []seedSpecies
	Indicators   []seedIndicator
	Institutions []seedInstitution
	Questions    []seedQuestion
}

func migrate(db *sql.DB) error {
	base, err := readProjectFile("migrations/001_init.sql")
	if err != nil {
		return err
	}
	if _, err = db.Exec(string(base)); err != nil {
		return err
	}
	if _, err = db.Exec("CREATE TABLE IF NOT EXISTS schema_migrations(version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)"); err != nil {
		return err
	}
	var applied int
	_ = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version=2").Scan(&applied)
	if applied == 0 {
		sqlBytes, e := readProjectFile("migrations/002_enrichment.sql")
		if e != nil {
			return e
		}
		tx, e := db.Begin()
		if e != nil {
			return e
		}
		if _, e = tx.Exec(string(sqlBytes)); e != nil {
			tx.Rollback()
			return e
		}
		if _, e = tx.Exec("INSERT OR IGNORE INTO schema_migrations(version) VALUES(2)"); e != nil {
			tx.Rollback()
			return e
		}
		if e = tx.Commit(); e != nil {
			return e
		}
	}
	return nil
}

func loadScienceSeed() (scienceSeed, error) {
	b, err := readProjectFile("seeds/science.json")
	if err != nil {
		return scienceSeed{}, err
	}
	var s scienceSeed
	err = json.Unmarshal(b, &s)
	return s, err
}
func readProjectFile(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err == nil {
		return b, nil
	}
	return os.ReadFile("../../" + path)
}

// seed is retained for package tests and callers that need the complete initial schema/data.
func seed(db *sql.DB) {
	if err := migrate(db); err != nil {
		log.Printf("test seed migration: %v", err)
		return
	}
	if err := seedScience(db); err != nil {
		log.Printf("test seed data: %v", err)
	}
}
func seedScience(db *sql.DB) error {
	s, err := loadScienceSeed()
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	sourceIDs := map[string]int64{}
	for _, x := range s.Sources {
		var id int64
		err = tx.QueryRow("SELECT id FROM scientific_sources WHERE url=?", x.URL).Scan(&id)
		if err == sql.ErrNoRows {
			res, e := tx.Exec("INSERT INTO scientific_sources(title,author,year,url,doi,kind,description) VALUES(?,?,?,?,?,?,?)", x.Title, x.Author, x.Year, x.URL, x.DOI, x.Kind, x.Description)
			if e != nil {
				return e
			}
			id, _ = res.LastInsertId()
		} else if err != nil {
			return err
		}
		sourceIDs[x.Key] = id
	}
	for _, x := range s.Species {
		var exists int
		_ = tx.QueryRow("SELECT COUNT(*) FROM species WHERE scientific_name=?", x.Scientific).Scan(&exists)
		if exists > 0 {
			// Preenche imagens específicas nas espécies ainda usando a silhueta padrão;
			// imagens escolhidas pela equipe no painel são preservadas.
			if x.Image != "" && x.Image != "/images/shark-silhouette.svg" {
				if _, e := tx.Exec("UPDATE species SET image_url=? WHERE scientific_name=? AND image_url IN ('', '/images/shark-silhouette.svg')", x.Image, x.Scientific); e != nil {
					return e
				}
			}
			continue
		}
		res, e := tx.Exec(`INSERT INTO species(common_name,scientific_name,image_url,description,physical_characteristics,size,diet,habitat,distribution,brazil_presence,northeast_presence,ecology,threats,brazil_status,iucn_status,published,source_note,national_assessment_year,iucn_assessment_year)
  VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,1,?,?,?)`, x.Common, x.Scientific, x.Image, x.Description, x.Physical, x.Size, x.Diet, x.Habitat, x.Distribution, x.BrazilPresence, x.NortheastPresence, x.Ecology, x.Threats, x.BrazilStatus, x.IUCNStatus, "Avaliação nacional e global são independentes; consulte as fontes identificadas.", nullableYear(x.NationalYear), nullableYear(x.IUCNYear))
		if e != nil {
			return e
		}
		id, _ := res.LastInsertId()
		if sid := sourceIDs[x.SourceKey]; sid > 0 {
			if _, e = tx.Exec("INSERT OR IGNORE INTO species_sources(species_id,source_id) VALUES(?,?)", id, sid); e != nil {
				return e
			}
		}
		if sid := sourceIDs["iucn-"+x.Scientific]; sid > 0 {
			if _, e = tx.Exec("INSERT OR IGNORE INTO species_sources(species_id,source_id) VALUES(?,?)", id, sid); e != nil {
				return e
			}
		}
		if sid := sourceIDs["mma"]; sid > 0 {
			if _, e = tx.Exec("INSERT OR IGNORE INTO species_sources(species_id,source_id) VALUES(?,?)", id, sid); e != nil {
				return e
			}
		}
		if sid := sourceIDs["fishbase-"+x.Scientific]; sid > 0 {
			if _, e = tx.Exec("INSERT OR IGNORE INTO species_sources(species_id,source_id) VALUES(?,?)", id, sid); e != nil {
				return e
			}
		}
	}
	for _, x := range s.Indicators {
		sid := sourceIDs[x.SourceKey]
		if sid == 0 {
			continue
		}
		if _, err = tx.Exec(`INSERT INTO indicators(title,value,unit,year,scope,description,source_id,published,taxonomic_group,source_url)
   SELECT ?,?,?,?,?,?,?,1,?,? WHERE NOT EXISTS(SELECT 1 FROM indicators WHERE title=? AND year=? AND scope=?)`, x.Title, x.Value, x.Unit, x.Year, x.Scope, x.Description, sid, x.Group, sourceURL(s.Sources, x.SourceKey), x.Title, x.Year, x.Scope); err != nil {
			return err
		}
	}
	for _, x := range s.Institutions {
		if _, err = tx.Exec("INSERT INTO institutions(name,description,state,city,focus,website) SELECT ?,?,?,?,?,? WHERE NOT EXISTS(SELECT 1 FROM institutions WHERE name=? AND state=?)", x.Name, x.Description, x.State, x.City, x.Focus, x.Website, x.Name, x.State); err != nil {
			return err
		}
	}
	for _, x := range s.Questions {
		key := "quiz:" + x.Key
		var id int64
		err = tx.QueryRow("SELECT record_id FROM seed_keys WHERE entity='quiz' AND seed_key=?", key).Scan(&id)
		if err == sql.ErrNoRows {
			res, e := tx.Exec("INSERT INTO quiz_questions(prompt,explanation,source_url) VALUES(?,?,?)", x.Prompt, x.Explanation, x.SourceURL)
			if e != nil {
				return e
			}
			id, _ = res.LastInsertId()
			for i, label := range x.Options {
				if _, e = tx.Exec("INSERT INTO quiz_options(question_id,label,correct) VALUES(?,?,?)", id, label, boolInt(i == x.Correct)); e != nil {
					return e
				}
			}
			if _, e = tx.Exec("INSERT INTO seed_keys(entity,seed_key,record_id) VALUES('quiz',?,?)", key, id); e != nil {
				return e
			}
		} else if err != nil {
			return err
		}
	}
	return tx.Commit()
}
func nullableYear(y int) any {
	if y <= 0 {
		return nil
	}
	return y
}
func sourceURL(s []seedSource, key string) string {
	for _, x := range s {
		if x.Key == key {
			return x.URL
		}
	}
	return ""
}

func seedDemo(db *sql.DB) error {
	type demoRecord struct{ Key, State, Municipality, Date, Category, Status, Description string }
	b, e := readProjectFile("seeds/demo-observations.json")
	if e != nil {
		return e
	}
	var records []demoRecord
	if e = json.Unmarshal(b, &records); e != nil {
		return e
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, r := range records {
		_, err = tx.Exec(`INSERT INTO observations(state,municipality,occurred_on,category,species,description,status,is_demo,demo_key)
	 SELECT ?,?,?,?,?,?,?,1,? WHERE NOT EXISTS(SELECT 1 FROM observations WHERE demo_key=?)`, r.State, r.Municipality, r.Date, r.Category, "", r.Description, r.Status, r.Key, r.Key)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
