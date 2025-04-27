package server

import (
	"database/sql"
	"encoding/json"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Storage interface {
	SaveScript(script *Script) error
	ClearScripts() error
	GetScript(id string) (*Script, error)
	DeleteScript(id string) error
	// ListScripts returns scripts with optional filtering by search, category, and tag.
	ListScripts(offset, limit int, search, category, tag string) ([]*Script, error)
	SaveExecutionHistory(history *ExecutionHistory) error
	ListExecutionHistory(scriptID string, offset, limit int) ([]*ExecutionHistory, error)
	GetHistoryByID(id string) (*ExecutionHistory, error)
	DeleteHistoryByID(id string) error
}

// CategoryCount is used for category aggregation
type CategoryCount struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	// Create tables if not exist
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS scripts (
		id TEXT PRIMARY KEY,
		name TEXT,
		description TEXT,
		author TEXT,
		category TEXT,
		tags TEXT,
		inputs TEXT,
		path TEXT
	);
	CREATE TABLE IF NOT EXISTS history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		script_id TEXT,
		executed_at DATETIME,
		finished_at DATETIME,
		execute_request TEXT,
		output TEXT,
		exitcode INTEGER DEFAULT 0,
		incognito BOOLEAN DEFAULT 0,
		command TEXT
	);
	`)
	if err != nil {
		return nil, err
	}
	return &SQLiteStorage{db: db}, nil
}

func (s *SQLiteStorage) ClearScripts() error {
	_, err := s.db.Exec(`DELETE FROM scripts;`)
	return err
}

func (s *SQLiteStorage) SaveScript(script *Script) error {
	tags, _ := json.Marshal(script.Tags)
	inputs, _ := json.Marshal(script.Inputs)
	_, err := s.db.Exec(`
	INSERT OR REPLACE INTO scripts (id, name, description, author, category, tags, inputs, path)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		script.ID, script.Name, script.Description, script.Author, script.Category, string(tags), string(inputs), script.Path)
	return err
}

func (s *SQLiteStorage) GetScript(id string) (*Script, error) {
	row := s.db.QueryRow(`SELECT id, name, description, author, category, tags, inputs, path FROM scripts WHERE id = ?`, id)
	var script Script
	var tags, inputs string
	err := row.Scan(&script.ID, &script.Name, &script.Description, &script.Author, &script.Category, &tags, &inputs, &script.Path)
	if err != nil {
		return nil, err
	}
	// Unmarshal tags
	if err := json.Unmarshal([]byte(tags), &script.Tags); err != nil {
		script.Tags = []string{}
	}
	// Unmarshal inputs with fallback to empty slice
	if err := json.Unmarshal([]byte(inputs), &script.Inputs); err != nil || script.Inputs == nil {
		script.Inputs = []Input{}
	}
	return &script, nil
}

func (s *SQLiteStorage) DeleteScript(id string) error {
	_, err := s.db.Exec(`DELETE FROM scripts WHERE id = ?`, id)
	return err
}

func (s *SQLiteStorage) ListScripts(offset, limit int, search, category, tag string) ([]*Script, error) {
	var args []interface{}
	var wheres []string

	if search != "" {
		wheres = append(wheres, "(name LIKE ? OR description LIKE ? OR author LIKE ? OR category LIKE ? OR tags LIKE ? OR path LIKE ?)")
		q := "%" + search + "%"
		args = append(args, q, q, q, q, q, q)
	}
	if category != "" {
		wheres = append(wheres, "LOWER(category) = ?")
		args = append(args, strings.ToLower(category))
	}
	if tag != "" {
		wheres = append(wheres, "tags LIKE ?") // simple LIKE match for tag string
		args = append(args, "%"+tag+"%")
	}
	query := "SELECT id, name, description, author, category, tags, inputs, path FROM scripts"
	if len(wheres) > 0 {
		query += " WHERE " + strings.Join(wheres, " AND ")
	}
	query += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var scripts []*Script
	for rows.Next() {
		var script Script
		var tags, inputs string
		if err := rows.Scan(&script.ID, &script.Name, &script.Description, &script.Author, &script.Category, &tags, &inputs, &script.Path); err != nil {
			continue
		}
		if err := json.Unmarshal([]byte(tags), &script.Tags); err != nil {
			script.Tags = []string{}
		}
		if inputs == "" {
			script.Inputs = []Input{}
		} else {
			if err := json.Unmarshal([]byte(inputs), &script.Inputs); err != nil {
				return nil, err
			}
		}
		scripts = append(scripts, &script)
	}
	return scripts, nil
}

func (s *SQLiteStorage) SaveExecutionHistory(history *ExecutionHistory) error {
	req, _ := json.Marshal(history.ExecuteRequest)
	_, err := s.db.Exec(`
	INSERT INTO history (script_id, executed_at, finished_at, execute_request, output, exitcode, incognito, command)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		history.ScriptID, history.ExecutedAt, history.FinishedAt, string(req), history.Output, history.ExitCode, history.Incognito, history.Command)
	return err
}

func (s *SQLiteStorage) ListExecutionHistory(scriptID string, offset, limit int) ([]*ExecutionHistory, error) {
	rows, err := s.db.Query(`SELECT id, script_id, executed_at, finished_at, execute_request, output, exitcode, incognito, command FROM history WHERE script_id = ? ORDER BY executed_at DESC LIMIT ? OFFSET ?`, scriptID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var histories []*ExecutionHistory
	for rows.Next() {
		var h ExecutionHistory
		var req string
		var id int
		var incognito sql.NullBool
		var command sql.NullString
		if err := rows.Scan(&id, &h.ScriptID, &h.ExecutedAt, &h.FinishedAt, &req, &h.Output, &h.ExitCode, &incognito, &command); err != nil {
			continue
		}
		json.Unmarshal([]byte(req), &h.ExecuteRequest)
		h.Incognito = incognito.Valid && incognito.Bool
		if command.Valid {
			h.Command = command.String
		}
		histories = append(histories, &h)
	}
	return histories, nil
}

func (s *SQLiteStorage) GetHistoryByID(id string) (*ExecutionHistory, error) {
	row := s.db.QueryRow(`SELECT id, script_id, executed_at, finished_at, execute_request, output, exitcode, incognito, command FROM history WHERE id = ?`, id)
	var h ExecutionHistory
	var req string
	var hid int
	var incognito sql.NullBool
	var command sql.NullString
	err := row.Scan(&hid, &h.ScriptID, &h.ExecutedAt, &h.FinishedAt, &req, &h.Output, &h.ExitCode, &incognito, &command)
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(req), &h.ExecuteRequest)
	h.Incognito = incognito.Valid && incognito.Bool
	if command.Valid {
		h.Command = command.String
	}
	return &h, nil
}

func (s *SQLiteStorage) DeleteHistoryByID(id string) error {
	_, err := s.db.Exec(`DELETE FROM history WHERE id = ?`, id)
	return err
}

// Returns up to `limit` unique script IDs from the last `historyLimit` history entries
func (s *SQLiteStorage) GetRecentScriptIDs(historyLimit, limit int) ([]string, error) {
	rows, err := s.db.Query(`SELECT script_id FROM history ORDER BY executed_at DESC LIMIT ?`, historyLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	uniqueIDs := make([]string, 0, limit)
	idSet := make(map[string]struct{})
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			continue
		}
		if _, exists := idSet[id]; !exists {
			idSet[id] = struct{}{}
			uniqueIDs = append(uniqueIDs, id)
			if len(uniqueIDs) >= limit {
				break
			}
		}
	}
	return uniqueIDs, nil
}

// Returns script metadata for a list of script IDs (no content)
func (s *SQLiteStorage) GetScriptsByIDs(ids []string) ([]Script, error) {
	if len(ids) == 0 {
		return []Script{}, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	query := `SELECT id, name, description, author, category, tags, inputs, path FROM scripts WHERE id IN (` + strings.Join(placeholders, ",") + `)`
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var scripts []Script
	for rows.Next() {
		var s Script
		var tags, inputs string
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.Author, &s.Category, &tags, &inputs, &s.Path); err != nil {
			continue
		}
		json.Unmarshal([]byte(tags), &s.Tags)
		if inputs == "" {
			s.Inputs = []Input{}
		} else {
			json.Unmarshal([]byte(inputs), &s.Inputs)
		}
		scripts = append(scripts, s)
	}
	return scripts, nil
}

// Returns up to `limit` recent scripts (metadata, no content) that have history, using SQL join/group by
func (s *SQLiteStorage) GetRecentScriptsWithHistory(limit int) ([]Script, error) {
	query := `
	SELECT s.id, s.name, s.description, s.author, s.category, s.tags, s.inputs, s.path
	FROM scripts s
	JOIN (
	    SELECT script_id, MAX(executed_at) as last_executed
	    FROM history
	    WHERE script_id IS NOT NULL AND script_id != ''
	    GROUP BY script_id
	    ORDER BY last_executed DESC
	    LIMIT ?
	) h ON s.id = h.script_id
	ORDER BY h.last_executed DESC
	LIMIT ?`
	rows, err := s.db.Query(query, limit, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var scripts []Script
	for rows.Next() {
		var s Script
		var tags, inputs string
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.Author, &s.Category, &tags, &inputs, &s.Path); err != nil {
			continue
		}
		json.Unmarshal([]byte(tags), &s.Tags)
		if inputs == "" {
			s.Inputs = []Input{}
		} else {
			json.Unmarshal([]byte(inputs), &s.Inputs)
		}
		scripts = append(scripts, s)
	}
	return scripts, nil
}

// ListCategoryCounts returns a list of categories and the count of scripts in each
func (s *SQLiteStorage) ListCategoryCounts() ([]CategoryCount, error) {
	rows, err := s.db.Query(`SELECT COALESCE(NULLIF(TRIM(LOWER(category)), ''), 'uncategorized') as category, COUNT(*) as count FROM scripts GROUP BY category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []CategoryCount
	for rows.Next() {
		var cat CategoryCount
		if err := rows.Scan(&cat.Category, &cat.Count); err != nil {
			continue
		}
		result = append(result, cat)
	}
	return result, nil
}
