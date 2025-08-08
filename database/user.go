package database

import (
	"PLIC/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
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

func (db Database) CreateUser(ctx context.Context, user models.DBUsers) error {
	_, err := db.Database.NamedExecContext(ctx, `
		INSERT INTO users (id, username, email, bio, password, created_at, updated_at)
		VALUES (:id, :username, :email, :bio, :password, :created_at, :updated_at)`, user)
	if err != nil {
		return fmt.Errorf("échec de len'insertion utilisateur : %w", err)
	}
	return nil
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
		return nil, fmt.Errorf("error fetching user sports: %w", err)
	}
	return sports, nil
}
