package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// PagoProgramado representa una cuota o pago generado por un egreso fijo
type PagoProgramado struct {
	ID              string     `json:"id"`
	EgresoFijoID    string     `json:"egreso_fijo_id"`
	FechaProgramada time.Time  `json:"fecha_programada"`
	MontoEsperado   float64    `json:"monto_esperado"`
	Estado          string     `json:"estado"` // 'pendiente', 'pagado', 'vencido', 'cancelado'
	FechaPago       *time.Time `json:"fecha_pago"`
	MedioID         *string    `json:"medio_id"`
	MovimientoID    *string    `json:"movimiento_id"`
	Notas           *string    `json:"notas"`
	CreatedAt       time.Time  `json:"created_at"`
}

// CreatePagoProgramado inserta un pago programado manual
func (db *DB) CreatePagoProgramado(ctx context.Context, id *string, egresoFijoID string, fechaProgramada time.Time, monto float64, notas *string) (string, error) {
	var pID string
	var query string
	var err error
	if id != nil && *id != "" {
		query = `INSERT INTO pago_programado (id, egreso_fijo_id, fecha_programada, monto_esperado, notas)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (egreso_fijo_id, fecha_programada) DO UPDATE SET monto_esperado = EXCLUDED.monto_esperado
			RETURNING id`
		err = db.Pool.QueryRow(ctx, query, *id, egresoFijoID, fechaProgramada, monto, notas).Scan(&pID)
	} else {
		query = `INSERT INTO pago_programado (egreso_fijo_id, fecha_programada, monto_esperado, notas)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (egreso_fijo_id, fecha_programada) DO UPDATE SET monto_esperado = EXCLUDED.monto_esperado
			RETURNING id`
		err = db.Pool.QueryRow(ctx, query, egresoFijoID, fechaProgramada, monto, notas).Scan(&pID)
	}
	return pID, err
}

// GetPagosProgramados obtiene todos los pagos programados ordenados por fecha
func (db *DB) GetPagosProgramados(ctx context.Context) ([]PagoProgramado, error) {
	query := `SELECT id, egreso_fijo_id, fecha_programada, monto_esperado, estado,
		fecha_pago, medio_id, movimiento_id, notas, created_at
		FROM pago_programado ORDER BY fecha_programada ASC, created_at ASC`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []PagoProgramado
	for rows.Next() {
		var p PagoProgramado
		if err := rows.Scan(
			&p.ID, &p.EgresoFijoID, &p.FechaProgramada, &p.MontoEsperado, &p.Estado,
			&p.FechaPago, &p.MedioID, &p.MovimientoID, &p.Notas, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		lista = append(lista, p)
	}
	return lista, nil
}

// GetPagoProgramado busca un pago programado por ID
func (db *DB) GetPagoProgramado(ctx context.Context, id string) (PagoProgramado, error) {
	query := `SELECT id, egreso_fijo_id, fecha_programada, monto_esperado, estado,
		fecha_pago, medio_id, movimiento_id, notas, created_at
		FROM pago_programado WHERE id = $1`

	var p PagoProgramado
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.EgresoFijoID, &p.FechaProgramada, &p.MontoEsperado, &p.Estado,
		&p.FechaPago, &p.MedioID, &p.MovimientoID, &p.Notas, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PagoProgramado{}, fmt.Errorf("pago programado no encontrado: %w", id, err)
		}
		return PagoProgramado{}, err
	}
	return p, nil
}

// PagarPagoProgramado registra el pago de una cuota: crea el movimiento, descuenta saldo y actualiza estado
func (db *DB) PagarPagoProgramado(ctx context.Context, pagoID, medioID string, montoReal float64, notas *string) (string, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	// 1. Obtener pago programado y egreso fijo relacionado
	var egresoFijoID string
	var razon string
	var categoriaID *string
	err = tx.QueryRow(ctx, `
		SELECT p.egreso_fijo_id, e.razon, e.categoria_id 
		FROM pago_programado p
		JOIN egreso_fijo e ON p.egreso_fijo_id = e.id
		WHERE p.id = $1 FOR UPDATE`, pagoID).Scan(&egresoFijoID, &razon, &categoriaID)
	if err != nil {
		return "", fmt.Errorf("error al obtener pago programado: %w", err)
	}

	// 2. Crear Movimiento de Egreso (tipo 'E')
	now := time.Now()
	desc := fmt.Sprintf("Pago programado: %s", razon)
	if notas != nil && *notas != "" {
		desc = fmt.Sprintf("Pago programado: %s - %s", razon, *notas)
	}

	var movID string
	movQuery := `INSERT INTO movimiento (medio_id, categoria_id, tipo, fecha_movimiento, descripcion, monto, egreso_fijo_id)
		VALUES ($1, $2, 'E', $3, $4, $5, $6) RETURNING id`

	err = tx.QueryRow(ctx, movQuery, medioID, categoriaID, now, desc, montoReal, egresoFijoID).Scan(&movID)
	if err != nil {
		return "", fmt.Errorf("error al crear movimiento de pago: %w", err)
	}

	// 3. Descontar de saldo_actual
	_, err = tx.Exec(ctx, `INSERT INTO saldo_actual (saldo, medio_id) VALUES ($1, $2)
		ON CONFLICT (medio_id) DO UPDATE SET saldo = saldo_actual.saldo - EXCLUDED.saldo`, montoReal, medioID)
	if err != nil {
		return "", fmt.Errorf("error al actualizar saldo en pago programado: %w", err)
	}

	// 4. Actualizar pago_programado a 'pagado'
	updatePagoQuery := `UPDATE pago_programado
		SET estado = 'pagado', fecha_pago = $1, medio_id = $2, movimiento_id = $3, notas = $4
		WHERE id = $5`
	_, err = tx.Exec(ctx, updatePagoQuery, now, medioID, movID, notas, pagoID)
	if err != nil {
		return "", fmt.Errorf("error al actualizar estado de pago_programado: %w", err)
	}

	// 5. Marcar alertas asociadas como leídas
	_, _ = tx.Exec(ctx, `UPDATE alerta_pago SET leida = true WHERE pago_programado_id = $1`, pagoID)

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return movID, nil
}

// DeletePagoProgramado elimina un pago programado
func (db *DB) DeletePagoProgramado(ctx context.Context, id string) (string, error) {
	query := `DELETE FROM pago_programado WHERE id = $1 RETURNING id`
	var deletedID string
	err := db.Pool.QueryRow(ctx, query, id).Scan(&deletedID)
	return deletedID, err
}
