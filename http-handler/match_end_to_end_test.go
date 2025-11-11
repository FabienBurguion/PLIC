package main

import (
	"PLIC/mailer"
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

func Test_MatchLifecycle(t *testing.T) {
	s := &Service{}
	cleanup := s.InitServiceTest()
	defer func() { _ = cleanup() }()

	// --- Injecter le mock mailer
	mockMailer := mailer.NewMockMailer()
	s.mailer = mockMailer

	// === Fixtures ===
	court := models.NewDBCourtFixture()

	creator := models.NewDBUsersFixture().
		WithUsername("u_creator").
		WithEmail("creator@example.com")
	u2 := models.NewDBUsersFixture().
		WithUsername("u_two").
		WithEmail("u2@example.com")
	u3 := models.NewDBUsersFixture().
		WithUsername("u_three").
		WithEmail("u3@example.com")
	u4 := models.NewDBUsersFixture().
		WithUsername("u_four").
		WithEmail("u4@example.com")

	s.loadFixtures(DBFixtures{
		Courts: []models.DBCourt{court},
		Users:  []models.DBUsers{creator, u2, u3, u4},
	})

	// === 1) Create match (4 joueurs) par le "creator" ===
	matchReq := models.MatchRequest{
		Sport:           models.Foot,
		CourtID:         court.Id,
		Date:            time.Now().Add(1 * time.Hour),
		NbreParticipant: 4,
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
	defer func(Body io.ReadCloser) { _ = Body.Close() }(respCreate.Body)
	require.Equal(t, http.StatusCreated, respCreate.StatusCode)

	var createRes models.CreateMatchResponse
	readCreateBody, _ := io.ReadAll(respCreate.Body)
	require.NoError(t, json.Unmarshal(readCreateBody, &createRes))
	matchID := createRes.Id
	require.NotEmpty(t, matchID)

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
		err := s.JoinMatch(w, r, models.AuthInfo{IsConnected: true, UserID: user.Id})
		require.NoError(t, err)
		resp := w.Result()
		defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)
		require.Equal(t, expectedCode, resp.StatusCode)
	}

	join(u2, 1, http.StatusOK)
	join(u3, 2, http.StatusOK)
	join(u4, 2, http.StatusOK)

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

		err := s.StartMatch(w, r, models.AuthInfo{IsConnected: true, UserID: creator.Id})
		require.NoError(t, err)
		resp := w.Result()
		defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)
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

		err := s.FinishMatch(w, r, models.AuthInfo{IsConnected: true, UserID: u2.Id})
		require.NoError(t, err)
		resp := w.Result()
		defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		m, err = s.db.GetMatchById(context.Background(), matchID)
		require.NoError(t, err)
		require.Equal(t, models.ManqueScore, m.CurrentState)
	}

	// === 4.1) Vote-status interdit tant que ManqueScore
	{
		url := "/match/" + matchID + "/vote-status"
		r := httptest.NewRequest("GET", url, nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", matchID)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()

		err := s.GetMatchVoteStatus(w, r, models.AuthInfo{IsConnected: true, UserID: creator.Id})
		require.NoError(t, err)
		resp := w.Result()
		defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
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

		err := s.UpdateMatchScore(w, r, models.AuthInfo{IsConnected: true, UserID: user.Id})
		require.NoError(t, err)
		resp := w.Result()
		defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)
		require.Equal(t, expectedCode, resp.StatusCode)
	}

	// Désaccord
	updateScore(creator, 5, 3, http.StatusOK) // team1 propose 5-3
	updateScore(u3, 2, 2, http.StatusOK)      // team2 propose 2-2

	// === MAIL ASSERT: aucun mail "result" pendant le désaccord
	require.Equal(t, 0, mockMailer.GetSentCounts("result"), "no result email should be sent before consensus")

	m, err = s.db.GetMatchById(context.Background(), matchID)
	require.NoError(t, err)
	require.Equal(t, models.ManqueScore, m.CurrentState)
	require.NotNil(t, m.Score1)
	require.NotNil(t, m.Score2)
	require.Equal(t, 2, *m.Score1)
	require.Equal(t, 2, *m.Score2)

	// Consensus (team2 met 5-3) -> state = Termine
	updateScore(u3, 5, 3, http.StatusOK)

	// === MAIL ASSERT: un mail "result" à l’instant du consensus
	require.Equal(t, 1, mockMailer.GetSentCounts("result"), "one result email should be sent when consensus is reached")

	m, err = s.db.GetMatchById(context.Background(), matchID)
	require.NoError(t, err)
	require.Equal(t, models.Termine, m.CurrentState)
	require.NotNil(t, m.Score1)
	require.NotNil(t, m.Score2)
	require.Equal(t, 5, *m.Score1)
	require.Equal(t, 3, *m.Score2)

	// === Vérification ELO
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

	require.Equal(t, base+delta, rCreator.Elo)
	require.Equal(t, base+delta, rU2.Elo)
	require.Equal(t, base-delta, rU3.Elo)
	require.Equal(t, base-delta, rU4.Elo)

	// === 5.2) Vote-status OK pour creator
	{
		url := "/match/" + matchID + "/vote-status"
		r := httptest.NewRequest("GET", url, nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", matchID)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()

		err := s.GetMatchVoteStatus(w, r, models.AuthInfo{IsConnected: true, UserID: creator.Id})
		require.NoError(t, err)
		resp := w.Result()
		defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var st models.MatchVoteStatusResponse
		bodyBytes, _ := io.ReadAll(resp.Body)
		require.NoError(t, json.Unmarshal(bodyBytes, &st))

		require.Equal(t, matchID, st.MatchID)
		require.Equal(t, 1, st.PlayerTeam)
		require.True(t, st.MyTeam.HasVoted)
		require.True(t, st.Opponent.HasVoted)
		require.NotNil(t, st.MyTeam.Score)
		require.NotNil(t, st.Opponent.Score)
		require.Equal(t, 5, st.MyTeam.Score.Score1)
		require.Equal(t, 3, st.MyTeam.Score.Score2)
		require.Equal(t, 5, st.Opponent.Score.Score1)
		require.Equal(t, 3, st.Opponent.Score.Score2)
	}

	// === 5.3) Vote-status OK pour u3
	{
		url := "/match/" + matchID + "/vote-status"
		r := httptest.NewRequest("GET", url, nil)
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", matchID)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()

		err := s.GetMatchVoteStatus(w, r, models.AuthInfo{IsConnected: true, UserID: u3.Id})
		require.NoError(t, err)
		resp := w.Result()
		defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var st models.MatchVoteStatusResponse
		bodyBytes, _ := io.ReadAll(resp.Body)
		require.NoError(t, json.Unmarshal(bodyBytes, &st))

		require.Equal(t, matchID, st.MatchID)
		require.Equal(t, 2, st.PlayerTeam)
		require.True(t, st.MyTeam.HasVoted)
		require.True(t, st.Opponent.HasVoted)
		require.NotNil(t, st.MyTeam.Score)
		require.NotNil(t, st.Opponent.Score)
		require.Equal(t, 5, st.MyTeam.Score.Score1)
		require.Equal(t, 3, st.MyTeam.Score.Score2)
		require.Equal(t, 5, st.Opponent.Score.Score1)
		require.Equal(t, 3, st.Opponent.Score.Score2)
	}

	// Bonus: tentative de vote après consensus -> aucun nouvel email
	{
		body, _ := json.Marshal(models.UpdateScoreRequest{Score1: 9, Score2: 9})
		url := "/match/" + matchID + "/score"
		r := httptest.NewRequest("PATCH", url, bytes.NewReader(body))
		routeCtx := chi.NewRouteContext()
		routeCtx.URLParams.Add("id", matchID)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))
		w := httptest.NewRecorder()

		err := s.UpdateMatchScore(w, r, models.AuthInfo{IsConnected: true, UserID: u2.Id})
		require.NoError(t, err)

		resp := w.Result()
		defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		// MAIL ASSERT: compteur inchangé
		require.Equal(t, 1, mockMailer.GetSentCounts("result"), "no extra result email after consensus")
	}
}
