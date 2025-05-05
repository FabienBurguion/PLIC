package database

import (
	"PLIC/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (db Database) CheckUserExist(ctx context.Context, id string) (bool, error) {
	var user models.DBUsers

	err := db.Database.GetContext(ctx, &user, `
		SELECT id, username, password, created_at, updated_at
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
		SELECT id, username, password, created_at, updated_at
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

func (db Database) GetUserById(ctx context.Context, id string) (*models.DBUsers, error) {
	var user models.DBUsers

	err := db.Database.GetContext(ctx, &user, `
		SELECT id, username, password, created_at, updated_at
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

func (db Database) CreateUser(ctx context.Context, user models.DBUsers) error {
	_, err := db.Database.NamedExecContext(ctx, `
		INSERT INTO users (id, username, password, created_at, updated_at)
		VALUES (:id, :username, :password, :created_at, :updated_at)`, user)
	if err != nil {
		return fmt.Errorf("échec de l'insertion utilisateur : %w", err)
	}
	return nil
}

func (db Database) ChangePassword(ctx context.Context, username string, newPasswordHash string) error {
	_, err := db.Database.ExecContext(ctx, `
		UPDATE users
		SET password = $1, updated_at = NOW()
		WHERE username = $2
	`, newPasswordHash, username)
	if err != nil {
		return fmt.Errorf("échec de la mise à jour du mot de passe : %w", err)
	}
	return nil
}
