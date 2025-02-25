package database

import (
	"PLIC/models"
	"context"
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
	rows, err := db.Database.QueryContext(ctx, `
		SELECT id, name, email
		FROM users
		WHERE id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("échec de la requête SQL : %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("aucun utilisateur trouvé avec l'ID %s", id)
	}

	var user models.DBUser

	err = rows.Scan(
		&user.Id,
		&user.Name,
		&user.Email,
	)
	if err != nil {
		return nil, fmt.Errorf("échec de la lecture des données : %w", err)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur après le parcours des lignes : %w", err)
	}

	return &user, nil
}
