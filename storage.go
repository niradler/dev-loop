package main

import (
	"database/sql"
	"encoding/json"

	_ "github.com/mattn/go-sqlite3"
)

type Storage interface {
	SaveScript(script *Script) error
	GetScript(id string) (*Script, error)
	DeleteScript(id string) error
	ListScripts(offset, limit int) ([]*Script, error)
	SaveExecutionHistory(history *ExecutionHistory) error
	ListExecutionHistory(scriptID string, offset, limit int) ([]*ExecutionHistory, error)
	GetHistoryByID(id string) (*ExecutionHistory, error)
	DeleteHistoryByID(id string) error
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
		version TEXT,
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

func (s *SQLiteStorage) SaveScript(script *Script) error {
	tags, _ := json.Marshal(script.Tags)
	inputs, _ := json.Marshal(script.Inputs)
	_, err := s.db.Exec(`
	INSERT OR REPLACE INTO scripts (id, name, description, author, version, category, tags, inputs, path)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		script.ID, script.Name, script.Description, script.Author, script.Version, script.Category, string(tags), string(inputs), script.Path)
	return err
}

func (s *SQLiteStorage) GetScript(id string) (*Script, error) {
	row := s.db.QueryRow(`SELECT id, name, description, author, version, category, tags, inputs, path FROM scripts WHERE id = ?`, id)
	var script Script
	var tags, inputs string
	err := row.Scan(&script.ID, &script.Name, &script.Description, &script.Author, &script.Version, &script.Category, &tags, &inputs, &script.Path)
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

func (s *SQLiteStorage) ListScripts(offset, limit int) ([]*Script, error) {
	rows, err := s.db.Query(`SELECT id, name, description, author, version, category, tags, inputs, path FROM scripts LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var scripts []*Script
	for rows.Next() {
		var script Script
		var tags, inputs string
		if err := rows.Scan(&script.ID, &script.Name, &script.Description, &script.Author, &script.Version, &script.Category, &tags, &inputs, &script.Path); err != nil {
			continue
		}
		if err := json.Unmarshal([]byte(tags), &script.Tags); err != nil {
			script.Tags = []string{}
		}
		// Improved input unmarshalling
		if inputs == "" {
			script.Inputs = []Input{}
		} else {
			if err := json.Unmarshal([]byte(inputs), &script.Inputs); err != nil {
				return nil, err // propagate error if JSON is invalid
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
