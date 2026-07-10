package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ProyectoMetrica representa la tabla intermedia entre proyecto y metrica
type ProyectoMetrica struct {
	ID                 int             `json:"id"`
	IDProyecto         int             `json:"id_proyecto"`
	IDMetrica          int             `json:"id_metrica"`
	ConfigProgramacion json.RawMessage `json:"config_programacion"` // JSONB
	Activo             bool            `json:"activo"`
	CreatedAt          time.Time       `json:"created_at"`
}

// CreateProyectoMetrica inserta una nueva relación y devuelve su ID
func (db *DB) CreateProyectoMetrica(ctx context.Context, idProyecto, idMetrica int, configProgramacion string) (int, error) {
	query := `INSERT INTO proyecto_metrica (id_proyecto, id_metrica, config_programacion)
		VALUES ($1, $2, $3::jsonb) RETURNING id`

	var id int
	err := db.Pool.QueryRow(ctx, query, idProyecto, idMetrica, configProgramacion).Scan(&id)
	return id, err
}

// GetProyectoMetricas obtiene todas las relaciones activas
func (db *DB) GetProyectoMetricas(ctx context.Context) ([]ProyectoMetrica, error) {
	query := `SELECT id, id_proyecto, id_metrica, config_programacion, activo, created_at
		FROM proyecto_metrica WHERE activo = true ORDER BY created_at ASC`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []ProyectoMetrica
	for rows.Next() {
		var pm ProyectoMetrica
		if err := rows.Scan(&pm.ID, &pm.IDProyecto, &pm.IDMetrica, &pm.ConfigProgramacion, &pm.Activo, &pm.CreatedAt); err != nil {
			return nil, err
		}
		lista = append(lista, pm)
	}
	return lista, nil
}

// GetProyectoMetrica busca una relación por ID (solo activa)
func (db *DB) GetProyectoMetrica(ctx context.Context, id int) (ProyectoMetrica, error) {
	query := `SELECT id, id_proyecto, id_metrica, config_programacion, activo, created_at
		FROM proyecto_metrica WHERE id = $1 AND activo = true`

	var pm ProyectoMetrica
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&pm.ID, &pm.IDProyecto, &pm.IDMetrica, &pm.ConfigProgramacion, &pm.Activo, &pm.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ProyectoMetrica{}, fmt.Errorf("relación proyecto-métrica con id %d no encontrada: %w", id, err)
		}
		return ProyectoMetrica{}, err
	}
	return pm, nil
}

// UpdateProyectoMetrica actualiza todos los campos de una relación existente
func (db *DB) UpdateProyectoMetrica(ctx context.Context, id int, idProyecto, idMetrica int, configProgramacion string, activo bool) (int, error) {
	query := `UPDATE proyecto_metrica
		SET id_proyecto = $1, id_metrica = $2, config_programacion = $3::jsonb, activo = $4
		WHERE id = $5 RETURNING id`

	var updatedID int
	err := db.Pool.QueryRow(ctx, query, idProyecto, idMetrica, configProgramacion, activo, id).Scan(&updatedID)
	return updatedID, err
}

// DeleteProyectoMetrica borrado lógico (activo = false)
func (db *DB) DeleteProyectoMetrica(ctx context.Context, id int) (int, error) {
	query := `UPDATE proyecto_metrica SET activo = false WHERE id = $1 RETURNING id`
	var deletedID int
	err := db.Pool.QueryRow(ctx, query, id).Scan(&deletedID)
	return deletedID, err
}
