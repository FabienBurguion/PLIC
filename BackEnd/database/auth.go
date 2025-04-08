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
		SELECT id, username, password
		FROM users
		WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("aucun utilisateur trouvé avec l'ID %s", id)
		}
		return false, fmt.Errorf("échec de la requête SQL : %w", err)
	}

	return true, nil
}
