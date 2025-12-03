package database

import (
	"PLIC/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

func (db Database) CheckUserExist(ctx context.Context, id string) (bool, error) {
	var user models.DBUsers

	err := db.Database.GetContext(ctx, &user, `
		SELECT id, username, email, bio, password, created_at, updated_at
		FROM users
		WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("échec de la requête SQL : %w", err)
	}

	return true, nil
}

func (db Database) GetUserByUsername(ctx context.Context, username string) (*models.DBUsers, error) {
	var user models.DBUsers

	err := db.Database.GetContext(ctx, &user, `
		SELECT id, username, email, bio, password, created_at, updated_at
		FROM users
		WHERE username = $1`, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("échec de la requête SQL : %w", err)
	}

	return &user, nil
}

func (db Database) GetUserByEmail(ctx context.Context, email string) (*models.DBUsers, error) {
	var user models.DBUsers

	err := db.Database.GetContext(ctx, &user, `
		SELECT id, username, email, bio, password, created_at, updated_at
		FROM users
		WHERE email = $1`, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("échec de la requête SQL : %w", err)
	}

	return &user, nil
}

func (db Database) GetUserById(ctx context.Context, id string) (*models.DBUsers, error) {
	var user models.DBUsers

	err := db.Database.GetContext(ctx, &user, `
		SELECT id, username, email, bio, current_field_id, password, created_at, updated_at
		FROM users
		WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("échec de la requête SQL : %w", err)
	}

	return &user, nil
}

func (db Database) UpdateUser(ctx context.Context, data models.UserPatchRequest, userId string, now time.Time) error {
	query := "UPDATE users SET"
	var args []interface{}
	argPos := 1

	if data.Username != nil {
		query += fmt.Sprintf(" username = $%d,", argPos)
		args = append(args, *data.Username)
		argPos++
	}
	if data.Bio != nil {
		query += fmt.Sprintf(" bio = $%d,", argPos)
		args = append(args, *data.Bio)
		argPos++
	}
	if data.Email != nil {
		query += fmt.Sprintf(" email = $%d,", argPos)
		args = append(args, *data.Email)
		argPos++
	}
	if data.CurrentFieldId != nil {
		query += fmt.Sprintf(" current_field_id = $%d,", argPos)
		args = append(args, *data.CurrentFieldId)
		argPos++
	}

	if len(args) == 0 {
		return nil
	}

	query += fmt.Sprintf(" updated_at = $%d", argPos)
	args = append(args, now)
	argPos++

	query += fmt.Sprintf(" WHERE id = $%d", argPos)
	args = append(args, userId)

	_, err := db.Database.ExecContext(ctx, query, args...)
	return err
}

var (
	ErrEmailTaken    = errors.New("email already taken")
	ErrUsernameTaken = errors.New("username already taken")
)

func (db Database) CreateUser(ctx context.Context, user models.DBUsers) error {
	_, err := db.Database.NamedExecContext(ctx, `
		INSERT INTO users (id, username, email, bio, password, created_at, updated_at)
		VALUES (:id, :username, :email, :bio, :password, :created_at, :updated_at)`, user)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "users_email_uniq", "users_email_key":
			return ErrEmailTaken
		case "users_username_uniq", "users_username_key", "unique_username":
			return ErrUsernameTaken
		}
	}
	return err
}

func (db Database) ChangePassword(ctx context.Context, email string, newPasswordHash string) error {
	_, err := db.Database.ExecContext(ctx, `
		UPDATE users
		SET password = $1, updated_at = NOW()
		WHERE email = $2
	`, newPasswordHash, email)
	if err != nil {
		return fmt.Errorf("échec de la mise à jour du mot de passe : %w", err)
	}
	return nil
}

func (db Database) DeleteUser(ctx context.Context, userId string) error {
	_, err := db.Database.ExecContext(ctx, `
		DELETE FROM users
		WHERE id = $1
	`, userId)
	if err != nil {
		return fmt.Errorf("échec de la suppresion du user "+userId+" : %w", err)
	}
	return nil
}

func (db Database) GetFavoriteFieldByUserID(ctx context.Context, userID string) (*string, error) {
	var name string
	err := db.Database.GetContext(ctx, &name, `
		SELECT c.name
		FROM user_match um
		JOIN matches m ON m.id = um.match_id
		JOIN courts c ON c.id = m.court_id
		WHERE um.user_id = $1
		GROUP BY c.name
		ORDER BY COUNT(*) DESC
		LIMIT 1
	`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error fetching favorite field: %w", err)
	}
	return &name, nil
}

func (db Database) GetFavoriteSportByUserID(ctx context.Context, userID string) (*models.Sport, error) {
	var sport models.Sport
	err := db.Database.GetContext(ctx, &sport, `
		SELECT m.sport
		FROM user_match um
		JOIN matches m ON m.id = um.match_id
		WHERE um.user_id = $1
		GROUP BY m.sport
		ORDER BY COUNT(*) DESC
		LIMIT 1
	`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error fetching favorite sport: %w", err)
	}
	return &sport, nil
}

func (db Database) GetPlayedSportsByUserID(ctx context.Context, userID string) ([]models.Sport, error) {
	var sports []models.Sport
	err := db.Database.SelectContext(ctx, &sports, `
		SELECT DISTINCT m.sport
		FROM user_match um
		JOIN matches m ON m.id = um.match_id
		WHERE um.user_id = $1
	`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error fetching user sports: %w", err)
	}
	return sports, nil
}

func (db Database) GetUserStatsByIDs(ctx context.Context, userIDs []string) (map[string]*models.UserStats, error) {
	stats := make(map[string]*models.UserStats, len(userIDs))
	initUser := func(uid string) *models.UserStats {
		if s, ok := stats[uid]; ok {
			return s
		}
		s := &models.UserStats{}
		stats[uid] = s
		return s
	}

	{
		type row struct {
			UserID string `db:"user_id"`
			Cnt    int    `db:"cnt"`
		}
		var rows []row
		q := `
			SELECT um.user_id, COUNT(*) AS cnt
			FROM user_match um
			JOIN matches m ON m.id = um.match_id
			WHERE um.user_id = ANY($1)
			  AND m.current_state IN ('Termine','Manque Score')
			GROUP BY um.user_id
		`
		if err := db.Database.SelectContext(ctx, &rows, q, userIDs); err != nil {
			return nil, fmt.Errorf("GetUserStatsByIDs.match_count: %w", err)
		}
		for _, r := range rows {
			initUser(r.UserID).MatchCount = r.Cnt
		}
	}

	{
		type row struct {
			UserID string `db:"user_id"`
			Cnt    int    `db:"cnt"`
		}
		var rows []row
		q := `
			SELECT um.user_id, COALESCE(COUNT(DISTINCT m.court_id), 0) AS cnt
			FROM user_match um
			JOIN matches m ON m.id = um.match_id
			WHERE um.user_id = ANY($1)
			GROUP BY um.user_id
		`
		if err := db.Database.SelectContext(ctx, &rows, q, userIDs); err != nil {
			return nil, fmt.Errorf("GetUserStatsByIDs.visited_fields: %w", err)
		}
		for _, r := range rows {
			initUser(r.UserID).VisitedFields = r.Cnt
		}
	}

	{
		type row struct {
			UserID string  `db:"user_id"`
			Name   *string `db:"name"`
		}
		var rows []row
		q := `
			SELECT DISTINCT ON (t.user_id) t.user_id, t.name
			FROM (
				SELECT um.user_id, c.name, COUNT(*) AS cnt
				FROM user_match um
				JOIN matches m ON m.id = um.match_id
				JOIN courts c  ON c.id = m.court_id
				WHERE um.user_id = ANY($1)
				GROUP BY um.user_id, c.name
			) AS t
			ORDER BY t.user_id, t.cnt DESC
		`
		if err := db.Database.SelectContext(ctx, &rows, q, userIDs); err != nil {
			return nil, fmt.Errorf("GetUserStatsByIDs.favorite_field: %w", err)
		}
		for _, r := range rows {
			initUser(r.UserID).FavoriteField = r.Name
		}
	}

	{
		type row struct {
			UserID string        `db:"user_id"`
			Sport  *models.Sport `db:"sport"`
		}
		var rows []row
		q := `
			SELECT DISTINCT ON (t.user_id) t.user_id, t.sport
			FROM (
				SELECT um.user_id, m.sport, COUNT(*) AS cnt
				FROM user_match um
				JOIN matches m ON m.id = um.match_id
				WHERE um.user_id = ANY($1)
				GROUP BY um.user_id, m.sport
			) AS t
			ORDER BY t.user_id, t.cnt DESC
		`
		if err := db.Database.SelectContext(ctx, &rows, q, userIDs); err != nil {
			return nil, fmt.Errorf("GetUserStatsByIDs.favorite_sport: %w", err)
		}
		for _, r := range rows {
			initUser(r.UserID).FavoriteSport = r.Sport
		}
	}

	{
		type row struct {
			UserID string `db:"user_id"`
			Sports string `db:"sports"`
		}
		var rows []row
		q := `
			SELECT um.user_id, STRING_AGG(DISTINCT m.sport::text, ',') AS sports
			FROM user_match um
			JOIN matches m ON m.id = um.match_id
			WHERE um.user_id = ANY($1)
			GROUP BY um.user_id
		`
		if err := db.Database.SelectContext(ctx, &rows, q, userIDs); err != nil {
			return nil, fmt.Errorf("GetUserStatsByIDs.played_sports: %w", err)
		}
		for _, r := range rows {
			parts := strings.Split(r.Sports, ",")
			ms := make([]models.Sport, 0, len(parts))
			for _, s := range parts {
				ms = append(ms, models.Sport(s))
			}
			initUser(r.UserID).Sports = ms
		}
	}

	{
		type row struct {
			UserID  string `db:"user_id"`
			Ranking int    `db:"ranking"`
			Name    string `db:"name"`
			Elo     int    `db:"elo"`
		}
		var rows []row
		q := `
			WITH ranked AS (
				SELECT
					r.user_id,
					RANK() OVER (PARTITION BY r.court_id ORDER BY r.elo DESC) AS ranking,
					c.name,
					r.elo
				FROM ranking r
				JOIN courts c ON c.id = r.court_id
				WHERE r.user_id = ANY($1)
			)
			SELECT user_id, ranking, name, elo
			FROM ranked
		`
		if err := db.Database.SelectContext(ctx, &rows, q, userIDs); err != nil {
			return nil, fmt.Errorf("GetUserStatsByIDs.ranked_fields: %w", err)
		}
		for _, r := range rows {
			s := initUser(r.UserID)
			s.Fields = append(s.Fields, models.Field{
				Ranking: r.Ranking,
				Name:    r.Name,
				Elo:     r.Elo,
			})
		}
	}

	{
		type row struct {
			UserID string `db:"user_id"`
			Wins   int    `db:"wins"`
			Total  int    `db:"total"`
		}
		var rows []row
		q := `
			SELECT
			  um.user_id,
			  COALESCE(SUM(
				CASE
				  WHEN (um.team = 1 AND m.score1 > m.score2) OR
					   (um.team = 2 AND m.score2 > m.score1)
				  THEN 1 ELSE 0
				END
			  ), 0) AS wins,
			  COUNT(*) FILTER (
				  WHERE m.current_state = 'Termine'
					AND m.score1 IS NOT NULL
					AND m.score2 IS NOT NULL
					AND m.score1 <> m.score2
			  ) AS total
			FROM user_match um
			JOIN matches m ON m.id = um.match_id
			WHERE um.user_id = ANY($1)
			GROUP BY um.user_id
		`
		if err := db.Database.SelectContext(ctx, &rows, q, userIDs); err != nil {
			return nil, fmt.Errorf("GetUserStatsByIDs.winrate: %w", err)
		}
		for _, r := range rows {
			var pct *int
			if r.Total > 0 {
				v := int((float64(r.Wins)/float64(r.Total))*100.0 + 0.5)
				pct = &v
			}
			initUser(r.UserID).Winrate = pct
		}
	}

	return stats, nil
}
