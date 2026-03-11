package models

import "time"

// User is the public-facing user struct (no password).
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// UserRow is used internally for DB scanning (includes password hash).
type UserRow struct {
	ID           string
	Username     string
	PasswordHash string
	CreatedAt    int64
}

type Comparison struct {
	ID         string    `json:"id"`
	OwnerID    string    `json:"owner_id"`
	Name       string    `json:"name"`
	ShareToken string    `json:"share_token"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Boat struct {
	ID           string  `json:"id"`
	ComparisonID string  `json:"comparison_id"`
	Name         string  `json:"name"`
	Position     int     `json:"position"`
	Scores       []Score `json:"scores,omitempty"`
}

type Score struct {
	BoatID     string `json:"boat_id"`
	QuestionID int    `json:"question_id"`
	Value      *int   `json:"value"` // nil = unscored
	Notes      string `json:"notes"`
}

type ScoreInput struct {
	QuestionID int    `json:"question_id"`
	Value      *int   `json:"value"`
	Notes      string `json:"notes"`
}

type SectionResult struct {
	Name          string  `json:"name"`
	Weight        float64 `json:"weight"`
	RawAverage    float64 `json:"raw_average"`
	WeightedScore float64 `json:"weighted_score"`
	ScoredCount   int     `json:"scored_count"`
	TotalCount    int     `json:"total_count"`
}

type BoatResult struct {
	Boat          Boat            `json:"boat"`
	Sections      []SectionResult `json:"sections"`
	TotalWeighted float64         `json:"total_weighted"`
}

type ComparisonDetail struct {
	ID         string       `json:"id"`
	OwnerID    string       `json:"owner_id"`
	Name       string       `json:"name"`
	ShareToken string       `json:"share_token"`
	ExpiresAt  time.Time    `json:"expires_at"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
	Results    []BoatResult `json:"results"`
}
