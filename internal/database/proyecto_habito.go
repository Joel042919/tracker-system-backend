package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ProyectoHabito representa la tabla proyecto_habito
type ProyectoHabito struct {
	ID                    int             `json:"id"`
	IDProyecto            int             `json:"id_proyecto"`
	DiasSemana            json.RawMessage `json:"dias_semana"` // JSONB: {"lunes": true, "martes": false, ...}
	HoraObjetivo          *string         `json:"hora_objetivo"` // HH:MM:SS o null
	PointsPorCompletar    int             `json:"points_por_completar"`
	RecordStreak          int             `json:"record_streak"`
	BestStreak            int             `json:"best_streak"`
	UltimaFechaCompletada *time.Time      `json:"ultima_fecha_completada"`
	Activo                bool            `json:"activo"`
	CreatedAt             time.Time       `json:"created_at"`
}

// CreateProyectoHabito inserta un nuevo proyecto_habito
func (db *DB) CreateProyectoHabito(ctx context.Context, idProyecto int, diasSemana string, horaObjetivo *string, pointsPorCompletar int) (int, error) {
	query := `INSERT INTO proyecto_habito (id_proyecto, dias_semana, hora_objetivo, points_por_completar)
		VALUES ($1, $2::jsonb, $3, $4) RETURNING id`

	var id int
	err := db.Pool.QueryRow(ctx, query, idProyecto, diasSemana, horaObjetivo, pointsPorCompletar).Scan(&id)
	return id, err
}

// GetProyectoHabitos obtiene todos los hábitos activos
func (db *DB) GetProyectoHabitos(ctx context.Context) ([]ProyectoHabito, error) {
	query := `SELECT id, id_proyecto, dias_semana, hora_objetivo::text, points_por_completar,
		record_streak, best_streak, ultima_fecha_completada, activo, created_at
		FROM proyecto_habito WHERE activo = true ORDER BY created_at ASC`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []ProyectoHabito
	for rows.Next() {
		var ph ProyectoHabito
		if err := rows.Scan(
			&ph.ID, &ph.IDProyecto, &ph.DiasSemana, &ph.HoraObjetivo, &ph.PointsPorCompletar,
			&ph.RecordStreak, &ph.BestStreak, &ph.UltimaFechaCompletada, &ph.Activo, &ph.CreatedAt,
		); err != nil {
			return nil, err
		}
		lista = append(lista, ph)
	}
	return lista, nil
}

// GetProyectoHabito busca un hábito por su ID
func (db *DB) GetProyectoHabito(ctx context.Context, id int) (ProyectoHabito, error) {
	query := `SELECT id, id_proyecto, dias_semana, hora_objetivo::text, points_por_completar,
		record_streak, best_streak, ultima_fecha_completada, activo, created_at
		FROM proyecto_habito WHERE id = $1`

	var ph ProyectoHabito
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&ph.ID, &ph.IDProyecto, &ph.DiasSemana, &ph.HoraObjetivo, &ph.PointsPorCompletar,
		&ph.RecordStreak, &ph.BestStreak, &ph.UltimaFechaCompletada, &ph.Activo, &ph.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ProyectoHabito{}, fmt.Errorf("proyecto_habito con id %d no encontrado: %w", id, err)
		}
		return ProyectoHabito{}, err
	}
	return ph, nil
}

// GetProyectoHabitoByProyectoID busca el hábito activo asociado a un proyecto
func (db *DB) GetProyectoHabitoByProyectoID(ctx context.Context, idProyecto int) (ProyectoHabito, error) {
	query := `SELECT id, id_proyecto, dias_semana, hora_objetivo::text, points_por_completar,
		record_streak, best_streak, ultima_fecha_completada, activo, created_at
		FROM proyecto_habito WHERE id_proyecto = $1 AND activo = true`

	var ph ProyectoHabito
	err := db.Pool.QueryRow(ctx, query, idProyecto).Scan(
		&ph.ID, &ph.IDProyecto, &ph.DiasSemana, &ph.HoraObjetivo, &ph.PointsPorCompletar,
		&ph.RecordStreak, &ph.BestStreak, &ph.UltimaFechaCompletada, &ph.Activo, &ph.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ProyectoHabito{}, fmt.Errorf("proyecto_habito para proyecto %d no encontrado: %w", idProyecto, err)
		}
		return ProyectoHabito{}, err
	}
	return ph, nil
}

// UpdateProyectoHabito actualiza un proyecto_habito de forma parcial o total
func (db *DB) UpdateProyectoHabito(ctx context.Context, id int, diasSemana *string, horaObjetivo *string, pointsPorCompletar *int, recordStreak *int, bestStreak *int, ultimaFechaCompletada *time.Time, activo *bool) (int, error) {
	current, err := db.GetProyectoHabito(ctx, id)
	if err != nil {
		return 0, err
	}

	finalDiasSemana := string(current.DiasSemana)
	if diasSemana != nil && *diasSemana != "" && *diasSemana != "{}" {
		finalDiasSemana = *diasSemana
	}

	finalHoraObjetivo := current.HoraObjetivo
	if horaObjetivo != nil {
		finalHoraObjetivo = horaObjetivo
	}

	finalPoints := current.PointsPorCompletar
	if pointsPorCompletar != nil {
		finalPoints = *pointsPorCompletar
	}

	finalRecordStreak := current.RecordStreak
	if recordStreak != nil {
		finalRecordStreak = *recordStreak
	}

	finalBestStreak := current.BestStreak
	if bestStreak != nil {
		finalBestStreak = *bestStreak
	}

	finalUltimaFecha := current.UltimaFechaCompletada
	if ultimaFechaCompletada != nil {
		finalUltimaFecha = ultimaFechaCompletada
	}

	finalActivo := current.Activo
	if activo != nil {
		finalActivo = *activo
	}

	query := `UPDATE proyecto_habito
		SET dias_semana = $1::jsonb, hora_objetivo = $2, points_por_completar = $3,
		    record_streak = $4, best_streak = $5, ultima_fecha_completada = $6, activo = $7
		WHERE id = $8 RETURNING id`

	var updatedID int
	err = db.Pool.QueryRow(ctx, query, finalDiasSemana, finalHoraObjetivo, finalPoints, finalRecordStreak, finalBestStreak, finalUltimaFecha, finalActivo, id).Scan(&updatedID)
	return updatedID, err
}

// UpdateStreak actualiza el streak de un hábito
func (db *DB) UpdateStreak(ctx context.Context, id int, recordStreak, bestStreak int, ultimaFechaCompletada *time.Time) error {
	query := `UPDATE proyecto_habito
		SET record_streak = $1, best_streak = $2, ultima_fecha_completada = $3
		WHERE id = $4`

	_, err := db.Pool.Exec(ctx, query, recordStreak, bestStreak, ultimaFechaCompletada, id)
	return err
}

// DeleteProyectoHabito realiza soft delete (activo = false)
func (db *DB) DeleteProyectoHabito(ctx context.Context, id int) (int, error) {
	query := `UPDATE proyecto_habito SET activo = false WHERE id = $1 RETURNING id`
	var deletedID int
	err := db.Pool.QueryRow(ctx, query, id).Scan(&deletedID)
	return deletedID, err
}
