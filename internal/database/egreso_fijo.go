package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// EgresoFijo representa un gasto recurrente programado
type EgresoFijo struct {
	ID                   string          `json:"id"`
	Razon                string          `json:"razon"`
	Descripcion          *string         `json:"descripcion"`
	CategoriaID          *string         `json:"categoria_id"`
	Monto                float64         `json:"monto"`
	ProgramacionPago     json.RawMessage `json:"programacion_pago"` // JSONB
	RecordatorioDiasAntes int            `json:"recordatorio_dias_antes"`
	Activo               bool            `json:"activo"`
	FechaInicio          *time.Time      `json:"fecha_inicio"`
	FechaFin             *time.Time      `json:"fecha_fin"`
	CreatedAt            time.Time       `json:"created_at"`
}

// CreateEgresoFijo inserta un egreso fijo y genera pagos programados iniciales
func (db *DB) CreateEgresoFijo(ctx context.Context, razon string, descripcion, categoriaID *string, monto float64, programacionPago string, recordatorioDias int, fechaInicio, fechaFin *time.Time) (string, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var id string
	query := `INSERT INTO egreso_fijo (razon, descripcion, categoria_id, monto, programacion_pago, recordatorio_dias_antes, fecha_inicio, fecha_fin)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7, $8) RETURNING id`

	err = tx.QueryRow(ctx, query, razon, descripcion, categoriaID, monto, programacionPago, recordatorioDias, fechaInicio, fechaFin).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("error al insertar egreso_fijo: %w", err)
	}

	// Generar pagos programados iniciales (próximos 3 meses o según fechas)
	generarPagosProgramados(ctx, tx, id, monto, programacionPago, recordatorioDias, fechaInicio, fechaFin)

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return id, nil
}

// Genera pagos programados y alertas automáticas
func generarPagosProgramados(ctx context.Context, tx pgx.Tx, egresoFijoID string, monto float64, progJSON string, recordatorioDias int, fechaInicio, fechaFin *time.Time) {
	var prog map[string]interface{}
	if err := json.Unmarshal([]byte(progJSON), &prog); err != nil {
		return
	}

	frecuencia, _ := prog["frecuencia"].(string)
	now := time.Now()
	startDate := now
	if fechaInicio != nil && fechaInicio.After(now) {
		startDate = *fechaInicio
	}

	endDate := now.AddDate(0, 3, 0) // 3 meses hacia adelante por defecto
	if fechaFin != nil && fechaFin.Before(endDate) {
		endDate = *fechaFin
	}

	// Cálculo de fechas según frecuencia
	var fechas []time.Time
	diaMes := 1
	if dm, ok := prog["dia_mes"].(float64); ok && dm > 0 {
		diaMes = int(dm)
	}

	switch frecuencia {
	case "mensual":
		for curr := startDate; curr.Before(endDate) || curr.Equal(endDate); curr = curr.AddDate(0, 1, 0) {
			year, month, _ := curr.Date()
			d := diaMes
			if d > 28 && month == 2 {
				d = 28
			}
			pagoDate := time.Date(year, month, d, 0, 0, 0, 0, time.UTC)
			fechas = append(fechas, pagoDate)
		}
	case "diario":
		for curr := startDate; curr.Before(endDate) || curr.Equal(endDate); curr = curr.AddDate(0, 0, 1) {
			fechas = append(fechas, curr)
		}
	case "semanal":
		for curr := startDate; curr.Before(endDate) || curr.Equal(endDate); curr = curr.AddDate(0, 0, 7) {
			fechas = append(fechas, curr)
		}
	default:
		fechas = append(fechas, startDate)
	}

	for _, f := range fechas {
		var pagoID string
		pagoQuery := `INSERT INTO pago_programado (egreso_fijo_id, fecha_programada, monto_esperado, estado)
			VALUES ($1, $2, $3, 'pendiente')
			ON CONFLICT (egreso_fijo_id, fecha_programada) DO NOTHING
			RETURNING id`
		_ = tx.QueryRow(ctx, pagoQuery, egresoFijoID, f, monto).Scan(&pagoID)

		if pagoID != "" {
			// Crear alerta
			alertaMsg := fmt.Sprintf("Recordatorio de pago programado de S/ %.2f para %s", monto, f.Format("2006-01-02"))
			alertaFecha := f.AddDate(0, 0, -recordatorioDias)
			alertaQuery := `INSERT INTO alerta_pago (pago_programado_id, tipo_alerta, mensaje, fecha_alerta)
				VALUES ($1, 'recordatorio', $2, $3)`
			_, _ = tx.Exec(ctx, alertaQuery, pagoID, alertaMsg, alertaFecha)
		}
	}
}

// GetEgresosFijos obtiene todos los egresos fijos activos
func (db *DB) GetEgresosFijos(ctx context.Context) ([]EgresoFijo, error) {
	query := `SELECT id, razon, descripcion, categoria_id, monto, programacion_pago, 
		recordatorio_dias_antes, activo, fecha_inicio, fecha_fin, created_at
		FROM egreso_fijo WHERE activo = true ORDER BY created_at ASC`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []EgresoFijo
	for rows.Next() {
		var ef EgresoFijo
		if err := rows.Scan(
			&ef.ID, &ef.Razon, &ef.Descripcion, &ef.CategoriaID, &ef.Monto,
			&ef.ProgramacionPago, &ef.RecordatorioDiasAntes, &ef.Activo,
			&ef.FechaInicio, &ef.FechaFin, &ef.CreatedAt,
		); err != nil {
			return nil, err
		}
		lista = append(lista, ef)
	}
	return lista, nil
}

// GetEgresoFijo busca un egreso fijo por ID
func (db *DB) GetEgresoFijo(ctx context.Context, id string) (EgresoFijo, error) {
	query := `SELECT id, razon, descripcion, categoria_id, monto, programacion_pago, 
		recordatorio_dias_antes, activo, fecha_inicio, fecha_fin, created_at
		FROM egreso_fijo WHERE id = $1`

	var ef EgresoFijo
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&ef.ID, &ef.Razon, &ef.Descripcion, &ef.CategoriaID, &ef.Monto,
		&ef.ProgramacionPago, &ef.RecordatorioDiasAntes, &ef.Activo,
		&ef.FechaInicio, &ef.FechaFin, &ef.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EgresoFijo{}, fmt.Errorf("egreso fijo no encontrado: %w", id, err)
		}
		return EgresoFijo{}, err
	}
	return ef, nil
}

// UpdateEgresoFijo actualiza un egreso fijo
func (db *DB) UpdateEgresoFijo(ctx context.Context, id, razon string, descripcion, categoriaID *string, monto float64, programacionPago string, recordatorioDias int, activo bool, fechaInicio, fechaFin *time.Time) (string, error) {
	query := `UPDATE egreso_fijo
		SET razon = $1, descripcion = $2, categoria_id = $3, monto = $4, programacion_pago = $5::jsonb,
			recordatorio_dias_antes = $6, activo = $7, fecha_inicio = $8, fecha_fin = $9
		WHERE id = $10 RETURNING id`

	var updatedID string
	err := db.Pool.QueryRow(ctx, query, razon, descripcion, categoriaID, monto, programacionPago, recordatorioDias, activo, fechaInicio, fechaFin, id).Scan(&updatedID)
	return updatedID, err
}

// DeleteEgresoFijo realiza soft delete (activo = false)
func (db *DB) DeleteEgresoFijo(ctx context.Context, id string) (string, error) {
	query := `UPDATE egreso_fijo SET activo = false WHERE id = $1 RETURNING id`
	var deletedID string
	err := db.Pool.QueryRow(ctx, query, id).Scan(&deletedID)
	return deletedID, err
}
