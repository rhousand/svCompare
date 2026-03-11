package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rhousand/svcompare/internal/auth"
	"github.com/rhousand/svcompare/internal/db"
	mw "github.com/rhousand/svcompare/internal/middleware"
	"github.com/rhousand/svcompare/internal/models"
	"github.com/rhousand/svcompare/internal/scoring"
)

// Handlers holds shared dependencies for all HTTP handlers.
type Handlers struct {
	db        *db.DB
	auth      auth.Authenticator
	jwtSecret string
}

// New constructs a Handlers instance.
func New(database *db.DB, authenticator auth.Authenticator, jwtSecret string) *Handlers {
	return &Handlers{db: database, auth: authenticator, jwtSecret: jwtSecret}
}

// --- Helpers ---

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}

// --- Auth Handlers ---

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.auth.Login(r.Context(), body.Username, body.Password)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := auth.IssueToken(user.ID, h.jwtSecret, 24*time.Hour)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not issue token")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "svcompare_token",
		Value:    token,
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	respondJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "svcompare_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	respondJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserID(r)
	user, err := h.db.GetUserByID(userID)
	if err != nil || user == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"user": user})
}

// --- Comparison Handlers ---

func (h *Handlers) ListComparisons(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserID(r)
	list, err := h.db.ListComparisons(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}
	if list == nil {
		list = []models.Comparison{}
	}
	respondJSON(w, http.StatusOK, map[string]any{"data": list, "count": len(list)})
}

func (h *Handlers) CreateComparison(w http.ResponseWriter, r *http.Request) {
	userID := mw.GetUserID(r)
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}
	c, err := h.db.CreateComparison(userID, body.Name)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}
	respondJSON(w, http.StatusCreated, c)
}

// getAndAuthorize fetches a comparison by ID and verifies the caller owns it.
// Returns nil, false and writes the error response if authorization fails.
func (h *Handlers) getAndAuthorize(w http.ResponseWriter, r *http.Request) (*models.Comparison, bool) {
	id := chi.URLParam(r, "id")
	userID := mw.GetUserID(r)

	c, err := h.db.GetComparisonByID(id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return nil, false
	}
	if c == nil {
		respondError(w, http.StatusNotFound, "comparison not found")
		return nil, false
	}
	if c.OwnerID != userID {
		respondError(w, http.StatusForbidden, "forbidden")
		return nil, false
	}
	return c, true
}

func (h *Handlers) GetComparison(w http.ResponseWriter, r *http.Request) {
	c, ok := h.getAndAuthorize(w, r)
	if !ok {
		return
	}
	h.renderDetail(w, c)
}

func (h *Handlers) UpdateComparison(w http.ResponseWriter, r *http.Request) {
	c, ok := h.getAndAuthorize(w, r)
	if !ok {
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}
	if err := h.db.UpdateComparison(c.ID, body.Name); err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "updated"})
}

func (h *Handlers) DeleteComparison(w http.ResponseWriter, r *http.Request) {
	_, ok := h.getAndAuthorize(w, r)
	if !ok {
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.db.DeleteComparison(id); err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// renderDetail loads boats+scores, runs scoring, and writes the full comparison response.
func (h *Handlers) renderDetail(w http.ResponseWriter, c *models.Comparison) {
	boats, err := h.db.GetBoatsWithScores(c.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}
	if boats == nil {
		boats = []models.Boat{}
	}

	results := make([]models.BoatResult, 0, len(boats))
	for _, b := range boats {
		results = append(results, scoring.Calculate(b))
	}

	respondJSON(w, http.StatusOK, models.ComparisonDetail{
		ID:         c.ID,
		OwnerID:    c.OwnerID,
		Name:       c.Name,
		ShareToken: c.ShareToken,
		ExpiresAt:  c.ExpiresAt,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
		Results:    results,
	})
}

// --- Boat Handlers ---

func (h *Handlers) AddBoat(w http.ResponseWriter, r *http.Request) {
	c, ok := h.getAndAuthorize(w, r)
	if !ok {
		return
	}
	count, err := h.db.CountBoats(c.ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}
	if count >= 5 {
		respondError(w, http.StatusConflict, "maximum 5 boats per comparison")
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}
	boat, err := h.db.AddBoat(c.ID, body.Name)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}
	respondJSON(w, http.StatusCreated, boat)
}

func (h *Handlers) UpdateBoat(w http.ResponseWriter, r *http.Request) {
	_, ok := h.getAndAuthorize(w, r)
	if !ok {
		return
	}
	bid := chi.URLParam(r, "bid")
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}
	if err := h.db.UpdateBoat(bid, body.Name); err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "updated"})
}

func (h *Handlers) DeleteBoat(w http.ResponseWriter, r *http.Request) {
	_, ok := h.getAndAuthorize(w, r)
	if !ok {
		return
	}
	bid := chi.URLParam(r, "bid")
	if err := h.db.DeleteBoat(bid); err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Score Handlers ---

func (h *Handlers) UpsertScores(w http.ResponseWriter, r *http.Request) {
	c, ok := h.getAndAuthorize(w, r)
	if !ok {
		return
	}
	bid := chi.URLParam(r, "bid")

	// Verify the boat belongs to this comparison.
	boat, err := h.db.GetBoatByID(bid)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}
	if boat == nil || boat.ComparisonID != c.ID {
		respondError(w, http.StatusNotFound, "boat not found")
		return
	}

	var body struct {
		Scores []models.ScoreInput `json:"scores"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	for _, s := range body.Scores {
		if s.QuestionID < 1 || s.QuestionID > 26 {
			respondError(w, http.StatusBadRequest, "question_id must be between 1 and 26")
			return
		}
		if s.Value != nil && (*s.Value < 1 || *s.Value > 10) {
			respondError(w, http.StatusBadRequest, "score value must be between 1 and 10")
			return
		}
	}

	if err := h.db.UpsertScores(bid, body.Scores); err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}

	// Reset the 30-day expiry window on every score save.
	_ = h.db.TouchComparison(c.ID)

	respondJSON(w, http.StatusOK, map[string]string{"message": "scores saved"})
}

// --- Share Handler ---

func (h *Handlers) GetShare(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	c, err := h.db.GetComparisonByShareToken(token)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "database error")
		return
	}
	if c == nil {
		respondError(w, http.StatusNotFound, "comparison not found or has expired")
		return
	}
	h.renderDetail(w, c)
}
