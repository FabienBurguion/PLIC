package database

import (
	"PLIC/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
)

type Database struct {
	Database *sqlx.DB
}

func (db Database) CreateUser(ctx context.Context, user models.DBUser) error {
	_, err := db.Database.NamedExecContext(ctx, `
		INSERT INTO users (id, name, email)
		VALUES (:id, :name, :email)`, user)
	if err != nil {
		return err
	}
	return nil
}

func (db Database) GetUser(ctx context.Context, id string) (*models.DBUser, error) {
	var user models.DBUser

	err := db.Database.GetContext(ctx, &user, `
		SELECT id, name, email
		FROM users
		WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("aucun utilisateur trouvé avec l'ID %s", id)
		}
		return nil, fmt.Errorf("échec de la requête SQL : %w", err)
	}

	return &user, nil
}
