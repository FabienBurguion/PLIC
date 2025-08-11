package main

import (
	"PLIC/models"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func Test_MatchLifecycle_WithConsensus(t *testing.T) {
	s := &Service{}
	cleanup := s.InitServiceTest()
	defer func() { _ = cleanup() }()

	// === Fixtures ===
	court := models.NewDBCourtFixture()

	creator := models.NewDBUsersFixture().
		WithUsername("u_creator").
		WithEmail("creator@example.com")
	// 3 joueurs additionnels
	u2 := models.NewDBUsersFixture().
		WithUsername("u_two").
		WithEmail("u2@example.com")
	u3 := models.NewDBUsersFixture().
		WithUsername("u_three").
		WithEmail("u3@example.com")
	u4 := models.NewDBUsersFixture().
		WithUsername("u_four").
		WithEmail("u4@example.com")

	// IMPORTANT: pas de ranking dans les fixtures
	s.loadFixtures(DBFixtures{
		Courts: []models.DBCourt{court},
		Users:  []models.DBUsers{creator, u2, u3, u4},
	})

	// === 1) Create match (4 joueurs) par le "creator" ===
	matchReq := models.MatchRequest{
		Sport:           models.Foot,
		CourtID:         court.Id,
		Date:            time.Now().Add(1 * time.Hour),
		NbreParticipant: 4, // 2 équipes de 2
	}
	bodyCreate, _ := json.Marshal(matchReq)
	rCreate := httptest.NewRequest("POST", "/match", bytes.NewReader(bodyCreate))
	wCreate := httptest.NewRecorder()

	err := s.CreateMatch(wCreate, rCreate, models.AuthInfo{
		IsConnected: true,
		UserID:      creator.Id,
	})
	require.NoError(t, err)
	respCreate := wCreate.Result()
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(respCreate.Body)
	require.Equal(t, http.StatusCreated, respCreate.StatusCode)

	var createRes models.CreateMatchResponse
	readCreateBody, _ := io.ReadAll(respCreate.Body)
	require.NoError(t, json.Unmarshal(readCreateBody, &createRes))
	matchID := createRes.Id
	require.NotEmpty(t, matchID, "match id should be returned")

	// === 2) 3 joueurs rejoignent ===
	join := func(user models.DBUsers, team int, expectedCode int) {
		joinReq := models.JoinMatchRequest{Team: team}
		b, _ := json.Marshal(joinReq)
		url := "/match/join/" + matchID
		r := httptest.NewRequest("POST", url, bytes.NewReader(b))
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", matchID)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()
		err := s.JoinMatch(w, r, models.AuthInfo{
			IsConnected: true,
			UserID:      user.Id,
		})
		require.NoError(t, err)
		resp := w.Result()
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)
		require.Equal(t, expectedCode, resp.StatusCode)
	}

	// Le créateur est déjà team1 (dans CreateMatch)
	join(u2, 1, http.StatusOK) // complète team1
	join(u3, 2, http.StatusOK)
	join(u4, 2, http.StatusOK) // complète team2

	// Le match doit être "Valide"
	m, err := s.db.GetMatchById(context.Background(), matchID)
	require.NoError(t, err)
	require.NotNil(t, m)
	require.Equal(t, models.Valide, m.CurrentState)

	// === 3) Start match par creator ===
	{
		url := "/match/" + matchID + "/start"
		r := httptest.NewRequest("PATCH", url, nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", matchID)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()

		err := s.StartMatch(w, r, models.AuthInfo{
			IsConnected: true,
			UserID:      creator.Id,
		})
		require.NoError(t, err)
		resp := w.Result()
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		m, err = s.db.GetMatchById(context.Background(), matchID)
		require.NoError(t, err)
		require.Equal(t, models.EnCours, m.CurrentState)
	}

	// === 4) Finish match par u2 ===
	{
		url := "/match/" + matchID + "/finish"
		r := httptest.NewRequest("PATCH", url, nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", matchID)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()

		err := s.FinishMatch(w, r, models.AuthInfo{
			IsConnected: true,
			UserID:      u2.Id,
		})
		require.NoError(t, err)
		resp := w.Result()
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		m, err = s.db.GetMatchById(context.Background(), matchID)
		require.NoError(t, err)
		require.Equal(t, models.ManqueScore, m.CurrentState)
	}

	// === 5) Votes de score ===
	updateScore := func(user models.DBUsers, s1, s2 int, expectedCode int) {
		body, _ := json.Marshal(models.UpdateScoreRequest{Score1: s1, Score2: s2})
		url := "/match/" + matchID + "/score"
		r := httptest.NewRequest("PATCH", url, bytes.NewReader(body))
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", matchID)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()

		err := s.UpdateMatchScore(w, r, models.AuthInfo{
			IsConnected: true,
			UserID:      user.Id,
		})
		require.NoError(t, err)
		resp := w.Result()
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)
		require.Equal(t, expectedCode, resp.StatusCode)
	}

	// Désaccord (state reste ManqueScore, score dernier proposé)
	updateScore(creator, 5, 3, http.StatusOK) // team1 = creator/u2 propose 5-3
	updateScore(u3, 2, 2, http.StatusOK)      // team2 propose 2-2

	m, err = s.db.GetMatchById(context.Background(), matchID)
	require.NoError(t, err)
	require.Equal(t, models.ManqueScore, m.CurrentState)
	require.NotNil(t, m.Score1)
	require.NotNil(t, m.Score2)
	require.Equal(t, 2, *m.Score1)
	require.Equal(t, 2, *m.Score2)

	// Consensus (team2 met à jour 5-3) -> state = Termine
	updateScore(u3, 5, 3, http.StatusOK)

	m, err = s.db.GetMatchById(context.Background(), matchID)
	require.NoError(t, err)
	require.Equal(t, models.Termine, m.CurrentState)
	require.NotNil(t, m.Score1)
	require.NotNil(t, m.Score2)
	require.Equal(t, 5, *m.Score1)
	require.Equal(t, 3, *m.Score2)

	// === Vérification des rankings créés et MAJ ===
	// Supposons DefaultElo=1000 et K=32 => delta par joueur = 16 (1v1 sur moyennes égales)
	const base = 1000
	const delta = 16

	ctx := context.Background()
	get := func(u models.DBUsers) *models.DBRanking {
		rk, err := s.db.GetRankingByUserAndCourt(ctx, u.Id, court.Id)
		require.NoError(t, err)
		require.NotNil(t, rk, "ranking should exist for user %s", u.Username)
		return rk
	}

	rCreator := get(creator)
	rU2 := get(u2)
	rU3 := get(u3)
	rU4 := get(u4)

	// Team1 a gagné (score1 > score2)
	require.Equal(t, base+delta, rCreator.Elo)
	require.Equal(t, base+delta, rU2.Elo)
	require.Equal(t, base-delta, rU3.Elo)
	require.Equal(t, base-delta, rU4.Elo)

	// Bonus: empêcher un second vote team1 (si tenté après consensus, on tombe de toute façon sur "wrong state")
	{
		body, _ := json.Marshal(models.UpdateScoreRequest{Score1: 9, Score2: 9})
		url := "/match/" + matchID + "/score"
		r := httptest.NewRequest("PATCH", url, bytes.NewReader(body))
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", matchID)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()

		err := s.UpdateMatchScore(w, r, models.AuthInfo{
			IsConnected: true,
			UserID:      u2.Id,
		})
		require.NoError(t, err)

		resp := w.Result()
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	}
}
