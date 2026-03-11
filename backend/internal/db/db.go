package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/rhousand/svcompare/internal/models"
	_ "modernc.org/sqlite"
)

// DB wraps a SQLite database connection.
type DB struct {
	db *sql.DB
}

// migrations are applied in order. Index 0 creates the migration tracker;
// subsequent indexes are the actual schema changes.
// Timestamps are stored as Unix integers to avoid SQLite time parsing issues.
var migrations = []string{
	// 0: migration tracking table (applied unconditionally before checking versions)
	`CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY)`,
	// 1: users
	`CREATE TABLE IF NOT EXISTS users (
		id       TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		created_at INTEGER NOT NULL DEFAULT (unixepoch())
	)`,
	// 2: comparisons
	`CREATE TABLE IF NOT EXISTS comparisons (
		id          TEXT PRIMARY KEY,
		owner_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		name        TEXT NOT NULL,
		share_token TEXT UNIQUE NOT NULL,
		expires_at  INTEGER NOT NULL,
		created_at  INTEGER NOT NULL DEFAULT (unixepoch()),
		updated_at  INTEGER NOT NULL DEFAULT (unixepoch())
	)`,
	// 3: boats
	`CREATE TABLE IF NOT EXISTS boats (
		id            TEXT PRIMARY KEY,
		comparison_id TEXT NOT NULL REFERENCES comparisons(id) ON DELETE CASCADE,
		name          TEXT NOT NULL,
		position      INTEGER NOT NULL,
		created_at    INTEGER NOT NULL DEFAULT (unixepoch()),
		UNIQUE(comparison_id, position)
	)`,
	// 4: scores (composite PK — one row per boat+question)
	`CREATE TABLE IF NOT EXISTS scores (
		boat_id     TEXT NOT NULL REFERENCES boats(id) ON DELETE CASCADE,
		question_id INTEGER NOT NULL,
		value       INTEGER,
		notes       TEXT NOT NULL DEFAULT '',
		updated_at  INTEGER NOT NULL DEFAULT (unixepoch()),
		PRIMARY KEY (boat_id, question_id)
	)`,
	// 5: indexes
	`CREATE INDEX IF NOT EXISTS idx_comparisons_owner   ON comparisons(owner_id)`,
	`CREATE INDEX IF NOT EXISTS idx_comparisons_share   ON comparisons(share_token)`,
	`CREATE INDEX IF NOT EXISTS idx_comparisons_expires ON comparisons(expires_at)`,
	`CREATE INDEX IF NOT EXISTS idx_boats_comparison    ON boats(comparison_id)`,
	`CREATE INDEX IF NOT EXISTS idx_scores_boat         ON scores(boat_id)`,
}

// Open opens (or creates) the SQLite database at path, applies WAL pragmas,
// runs pending migrations, and returns a ready-to-use DB.
func Open(path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	sqldb, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// SQLite is single-writer; keep one connection to avoid locking contention.
	sqldb.SetMaxOpenConns(1)

	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
	} {
		if _, err := sqldb.Exec(pragma); err != nil {
			return nil, fmt.Errorf("apply pragma %q: %w", pragma, err)
		}
	}

	d := &DB{db: sqldb}
	if err := d.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return d, nil
}

// Close closes the underlying database connection.
func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) migrate() error {
	// Migration 0 is always applied first (idempotent).
	if _, err := d.db.Exec(migrations[0]); err != nil {
		return fmt.Errorf("bootstrap migrations table: %w", err)
	}

	for i, m := range migrations[1:] {
		version := i + 1
		var applied int
		_ = d.db.QueryRow(`SELECT 1 FROM schema_migrations WHERE version = ?`, version).Scan(&applied)
		if applied == 1 {
			continue
		}
		if _, err := d.db.Exec(m); err != nil {
			return fmt.Errorf("migration %d: %w", version, err)
		}
		if _, err := d.db.Exec(`INSERT OR IGNORE INTO schema_migrations (version) VALUES (?)`, version); err != nil {
			return fmt.Errorf("record migration %d: %w", version, err)
		}
	}
	return nil
}

// SeedAdmin inserts an admin user if no users exist yet.
// passwordHash must already be bcrypt-hashed.
func (d *DB) SeedAdmin(username, passwordHash string) error {
	var count int
	if err := d.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := d.db.Exec(
		`INSERT INTO users (id, username, password) VALUES (?, ?, ?)`,
		uuid.New().String(), username, passwordHash,
	)
	return err
}

// GetUserByUsername looks up a user by username for login.
// Returns nil, nil if not found.
func (d *DB) GetUserByUsername(username string) (*models.UserRow, error) {
	row := &models.UserRow{}
	err := d.db.QueryRow(
		`SELECT id, username, password, created_at FROM users WHERE username = ?`, username,
	).Scan(&row.ID, &row.Username, &row.PasswordHash, &row.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return row, err
}

// GetUserByID looks up a user by ID (used by the /me endpoint).
// Returns nil, nil if not found.
func (d *DB) GetUserByID(id string) (*models.User, error) {
	u := &models.User{}
	var createdAt int64
	err := d.db.QueryRow(
		`SELECT id, username, created_at FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Username, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u.CreatedAt = time.Unix(createdAt, 0).UTC()
	return u, nil
}

// CreateComparison creates a new comparison with a random share token and 30-day expiry.
func (d *DB) CreateComparison(ownerID, name string) (*models.Comparison, error) {
	id := uuid.New().String()
	token := uuid.New().String()
	now := time.Now().UTC()
	expiresAt := now.Add(30 * 24 * time.Hour)

	_, err := d.db.Exec(
		`INSERT INTO comparisons (id, owner_id, name, share_token, expires_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, ownerID, name, token, expiresAt.Unix(), now.Unix(), now.Unix(),
	)
	if err != nil {
		return nil, err
	}
	return &models.Comparison{
		ID: id, OwnerID: ownerID, Name: name, ShareToken: token,
		ExpiresAt: expiresAt, CreatedAt: now, UpdatedAt: now,
	}, nil
}

// ListComparisons returns all comparisons for a user, newest-updated first.
func (d *DB) ListComparisons(ownerID string) ([]models.Comparison, error) {
	rows, err := d.db.Query(
		`SELECT id, owner_id, name, share_token, expires_at, created_at, updated_at
		 FROM comparisons WHERE owner_id = ? ORDER BY updated_at DESC`,
		ownerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Comparison
	for rows.Next() {
		var c models.Comparison
		var exp, created, updated int64
		if err := rows.Scan(&c.ID, &c.OwnerID, &c.Name, &c.ShareToken, &exp, &created, &updated); err != nil {
			return nil, err
		}
		c.ExpiresAt = time.Unix(exp, 0).UTC()
		c.CreatedAt = time.Unix(created, 0).UTC()
		c.UpdatedAt = time.Unix(updated, 0).UTC()
		out = append(out, c)
	}
	return out, rows.Err()
}

// GetComparisonByID fetches a comparison by primary key.
func (d *DB) GetComparisonByID(id string) (*models.Comparison, error) {
	return d.scanComparison(`WHERE id = ?`, id)
}

// GetComparisonByShareToken fetches a non-expired comparison by its share token.
func (d *DB) GetComparisonByShareToken(token string) (*models.Comparison, error) {
	return d.scanComparison(`WHERE share_token = ? AND expires_at > unixepoch()`, token)
}

func (d *DB) scanComparison(where string, args ...any) (*models.Comparison, error) {
	c := &models.Comparison{}
	var exp, created, updated int64
	err := d.db.QueryRow(
		`SELECT id, owner_id, name, share_token, expires_at, created_at, updated_at
		 FROM comparisons `+where, args...,
	).Scan(&c.ID, &c.OwnerID, &c.Name, &c.ShareToken, &exp, &created, &updated)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.ExpiresAt = time.Unix(exp, 0).UTC()
	c.CreatedAt = time.Unix(created, 0).UTC()
	c.UpdatedAt = time.Unix(updated, 0).UTC()
	return c, nil
}

// UpdateComparison updates the name of a comparison.
func (d *DB) UpdateComparison(id, name string) error {
	_, err := d.db.Exec(
		`UPDATE comparisons SET name = ?, updated_at = ? WHERE id = ?`,
		name, time.Now().Unix(), id,
	)
	return err
}

// TouchComparison resets the 30-day expiry window and bumps updated_at.
// Called whenever scores are saved.
func (d *DB) TouchComparison(id string) error {
	now := time.Now().UTC()
	_, err := d.db.Exec(
		`UPDATE comparisons SET updated_at = ?, expires_at = ? WHERE id = ?`,
		now.Unix(), now.Add(30*24*time.Hour).Unix(), id,
	)
	return err
}

// DeleteComparison deletes a comparison (boats and scores cascade).
func (d *DB) DeleteComparison(id string) error {
	_, err := d.db.Exec(`DELETE FROM comparisons WHERE id = ?`, id)
	return err
}

// CountBoats returns the number of boats in a comparison.
func (d *DB) CountBoats(comparisonID string) (int, error) {
	var n int
	err := d.db.QueryRow(`SELECT COUNT(*) FROM boats WHERE comparison_id = ?`, comparisonID).Scan(&n)
	return n, err
}

// AddBoat adds a boat to a comparison at the next available position.
func (d *DB) AddBoat(comparisonID, name string) (*models.Boat, error) {
	var maxPos sql.NullInt64
	_ = d.db.QueryRow(
		`SELECT MAX(position) FROM boats WHERE comparison_id = ?`, comparisonID,
	).Scan(&maxPos)
	position := 1
	if maxPos.Valid {
		position = int(maxPos.Int64) + 1
	}

	id := uuid.New().String()
	_, err := d.db.Exec(
		`INSERT INTO boats (id, comparison_id, name, position) VALUES (?, ?, ?, ?)`,
		id, comparisonID, name, position,
	)
	if err != nil {
		return nil, err
	}
	return &models.Boat{ID: id, ComparisonID: comparisonID, Name: name, Position: position}, nil
}

// UpdateBoat renames a boat.
func (d *DB) UpdateBoat(id, name string) error {
	_, err := d.db.Exec(`UPDATE boats SET name = ? WHERE id = ?`, name, id)
	return err
}

// DeleteBoat removes a boat (scores cascade).
func (d *DB) DeleteBoat(id string) error {
	_, err := d.db.Exec(`DELETE FROM boats WHERE id = ?`, id)
	return err
}

// GetBoatByID fetches a boat by primary key.
func (d *DB) GetBoatByID(id string) (*models.Boat, error) {
	b := &models.Boat{}
	err := d.db.QueryRow(
		`SELECT id, comparison_id, name, position FROM boats WHERE id = ?`, id,
	).Scan(&b.ID, &b.ComparisonID, &b.Name, &b.Position)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return b, err
}

// UpsertScores inserts or updates all scores for a boat in a single transaction.
func (d *DB) UpsertScores(boatID string, inputs []models.ScoreInput) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().Unix()
	stmt, err := tx.Prepare(
		`INSERT INTO scores (boat_id, question_id, value, notes, updated_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(boat_id, question_id) DO UPDATE SET
		   value      = excluded.value,
		   notes      = excluded.notes,
		   updated_at = excluded.updated_at`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range inputs {
		if _, err := stmt.Exec(boatID, s.QuestionID, s.Value, s.Notes, now); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetBoatsWithScores loads all boats for a comparison along with their scores.
func (d *DB) GetBoatsWithScores(comparisonID string) ([]models.Boat, error) {
	rows, err := d.db.Query(
		`SELECT id, comparison_id, name, position
		 FROM boats WHERE comparison_id = ? ORDER BY position`,
		comparisonID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var boats []models.Boat
	for rows.Next() {
		var b models.Boat
		if err := rows.Scan(&b.ID, &b.ComparisonID, &b.Name, &b.Position); err != nil {
			return nil, err
		}
		boats = append(boats, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range boats {
		scores, err := d.getScoresForBoat(boats[i].ID)
		if err != nil {
			return nil, err
		}
		boats[i].Scores = scores
	}
	return boats, nil
}

func (d *DB) getScoresForBoat(boatID string) ([]models.Score, error) {
	rows, err := d.db.Query(
		`SELECT boat_id, question_id, value, notes FROM scores WHERE boat_id = ?`, boatID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []models.Score
	for rows.Next() {
		var s models.Score
		if err := rows.Scan(&s.BoatID, &s.QuestionID, &s.Value, &s.Notes); err != nil {
			return nil, err
		}
		scores = append(scores, s)
	}
	return scores, rows.Err()
}

// DeleteExpiredComparisons removes comparisons past their expiry date.
// Called by the background cleanup goroutine.
func (d *DB) DeleteExpiredComparisons() (int64, error) {
	res, err := d.db.Exec(`DELETE FROM comparisons WHERE expires_at <= unixepoch()`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
