package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Movimiento representa un ingreso (I) o egreso (E)
type Movimiento struct {
	ID              string    `json:"id"`
	MedioID         string    `json:"medio_id"`
	CategoriaID     *string   `json:"categoria_id"`
	Tipo            string    `json:"tipo"` // 'I' (Ingreso) o 'E' (Egreso)
	FechaMovimiento time.Time `json:"fecha_movimiento"`
	Descripcion     *string   `json:"descripcion"`
	Monto           float64   `json:"monto"`
	EgresoFijoID    *string   `json:"egreso_fijo_id"`
	CreatedAt       time.Time `json:"created_at"`
}

// CreateMovimiento inserta un nuevo movimiento y actualiza atómicamente el saldo en saldo_actual
func (db *DB) CreateMovimiento(ctx context.Context, id *string, medioID string, categoriaID *string, tipo string, fechaMovimiento time.Time, descripcion *string, monto float64, egresoFijoID *string) (string, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("error iniciando transacción: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Asegurar que medioID existe en la base de datos para no violar FK
	var medioExists bool
	_ = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM medio WHERE id = $1::uuid)", medioID).Scan(&medioExists)
	if !medioExists {
		_, err = tx.Exec(ctx, "INSERT INTO medio (id, medio, tipo_medio) VALUES ($1::uuid, 'Cuenta', 'efectivo') ON CONFLICT (id) DO NOTHING", medioID)
		if err != nil {
			return "", fmt.Errorf("error asegurando medio: %w", err)
		}
	}

	// 2. Si categoriaID está presente, validar que exista en categoria para no violar FK
	var cleanCategoriaID *string
	if categoriaID != nil && *categoriaID != "" {
		var catExists bool
		_ = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM categoria WHERE id = $1::uuid)", *categoriaID).Scan(&catExists)
		if !catExists {
			_, err = tx.Exec(ctx, "INSERT INTO categoria (id, categoria) VALUES ($1::uuid, 'General') ON CONFLICT (id) DO NOTHING", *categoriaID)
			if err == nil {
				cleanCategoriaID = categoriaID
			}
		} else {
			cleanCategoriaID = categoriaID
		}
	}

	// 3. Verificación de Idempotencia / Deduplicación reciente (10 segundos)
	// Previene que múltiples clics rápidos con UUIDs diferentes dupliquen el movimiento y el saldo
	var existingDupID string
	dupQuery := `SELECT id::text FROM movimiento 
		WHERE medio_id = $1::uuid 
		  AND tipo = $2 
		  AND monto = $3 
		  AND COALESCE(descripcion, '') = COALESCE($4, '') 
		  AND created_at >= NOW() - INTERVAL '10 seconds'
		ORDER BY created_at DESC LIMIT 1`
	err = tx.QueryRow(ctx, dupQuery, medioID, tipo, monto, descripcion).Scan(&existingDupID)
	if err == nil && existingDupID != "" {
		// Duplicado reciente detectado -> Retornamos el id existente sin volver a descontar/sumar saldo
		_ = tx.Commit(ctx)
		return existingDupID, nil
	}

	var movID string
	var query string
	if id != nil && *id != "" {
		query = `INSERT INTO movimiento (id, medio_id, categoria_id, tipo, fecha_movimiento, descripcion, monto, egreso_fijo_id)
			VALUES ($1::uuid, $2::uuid, NULLIF($3, '')::uuid, $4, $5, $6, $7, NULLIF($8, '')::uuid)
			ON CONFLICT (id) DO NOTHING
			RETURNING id`
		err = tx.QueryRow(ctx, query, *id, medioID, cleanCategoriaID, tipo, fechaMovimiento, descripcion, monto, egresoFijoID).Scan(&movID)
		if errors.Is(err, sql.ErrNoRows) || movID == "" {
			// Ya existía el movimiento
			_ = tx.Commit(ctx)
			return *id, nil
		}
	} else {
		query = `INSERT INTO movimiento (medio_id, categoria_id, tipo, fecha_movimiento, descripcion, monto, egreso_fijo_id)
			VALUES ($1::uuid, NULLIF($2, '')::uuid, $3, $4, $5, $6, NULLIF($7, '')::uuid) RETURNING id`
		err = tx.QueryRow(ctx, query, medioID, cleanCategoriaID, tipo, fechaMovimiento, descripcion, monto, egresoFijoID).Scan(&movID)
	}
	if err != nil {
		return "", fmt.Errorf("error al insertar movimiento: %w", err)
	}

	// Ajuste de saldo en saldo_actual
	var delta float64
	if tipo == "I" {
		delta = monto
	} else {
		delta = -monto
	}

	var saldoExists bool
	_ = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM saldo_actual WHERE medio_id = $1::uuid)", medioID).Scan(&saldoExists)
	if saldoExists {
		_, err = tx.Exec(ctx, "UPDATE saldo_actual SET saldo = saldo + $1 WHERE medio_id = $2::uuid", delta, medioID)
	} else {
		_, err = tx.Exec(ctx, "INSERT INTO saldo_actual (id, saldo, medio_id) VALUES (gen_random_uuid(), $1, $2::uuid)", delta, medioID)
	}
	if err != nil {
		return "", fmt.Errorf("error al actualizar saldo_actual: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("error haciendo commit: %w", err)
	}

	return movID, nil
}

// GetMovimientos obtiene todos los movimientos ordenados cronológicamente
func (db *DB) GetMovimientos(ctx context.Context) ([]Movimiento, error) {
	query := `SELECT id, medio_id, categoria_id, tipo, fecha_movimiento, descripcion, monto, egreso_fijo_id, created_at
		FROM movimiento ORDER BY fecha_movimiento DESC, created_at DESC`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []Movimiento
	for rows.Next() {
		var m Movimiento
		if err := rows.Scan(
			&m.ID, &m.MedioID, &m.CategoriaID, &m.Tipo, &m.FechaMovimiento,
			&m.Descripcion, &m.Monto, &m.EgresoFijoID, &m.CreatedAt,
		); err != nil {
			return nil, err
		}
		lista = append(lista, m)
	}
	return lista, nil
}

// GetMovimiento busca un movimiento por ID
func (db *DB) GetMovimiento(ctx context.Context, id string) (Movimiento, error) {
	query := `SELECT id, medio_id, categoria_id, tipo, fecha_movimiento, descripcion, monto, egreso_fijo_id, created_at
		FROM movimiento WHERE id = $1`

	var m Movimiento
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.MedioID, &m.CategoriaID, &m.Tipo, &m.FechaMovimiento,
		&m.Descripcion, &m.Monto, &m.EgresoFijoID, &m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Movimiento{}, fmt.Errorf("movimiento no encontrado: %w", id, err)
		}
		return Movimiento{}, err
	}
	return m, nil
}

// UpdateMovimiento actualiza un movimiento revirtiendo el impacto anterior en el saldo y aplicando el nuevo
func (db *DB) UpdateMovimiento(ctx context.Context, id, medioID string, categoriaID *string, tipo string, fechaMovimiento time.Time, descripcion *string, monto float64, egresoFijoID *string) (string, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	// 1. Obtener movimiento anterior
	var oldMedioID, oldTipo string
	var oldMonto float64
	err = tx.QueryRow(ctx, `SELECT medio_id, tipo, monto FROM movimiento WHERE id = $1 FOR UPDATE`, id).Scan(&oldMedioID, &oldTipo, &oldMonto)
	if err != nil {
		return "", fmt.Errorf("error al obtener movimiento previo: %w", err)
	}

	// 2. Revertir saldo anterior
	var revertDelta float64
	if oldTipo == "I" {
		revertDelta = -oldMonto
	} else {
		revertDelta = oldMonto
	}
	_, err = tx.Exec(ctx, `UPDATE saldo_actual SET saldo = saldo + $1 WHERE medio_id = $2`, revertDelta, oldMedioID)
	if err != nil {
		return "", fmt.Errorf("error al revertir saldo anterior: %w", err)
	}

	// 3. Aplicar nuevo saldo
	var newDelta float64
	if tipo == "I" {
		newDelta = monto
	} else {
		newDelta = -monto
	}
	_, err = tx.Exec(ctx, `INSERT INTO saldo_actual (saldo, medio_id) VALUES ($1, $2)
		ON CONFLICT (medio_id) DO UPDATE SET saldo = saldo_actual.saldo + EXCLUDED.saldo`, newDelta, medioID)
	if err != nil {
		return "", fmt.Errorf("error al aplicar nuevo saldo: %w", err)
	}

	// 4. Actualizar movimiento
	updateQuery := `UPDATE movimiento
		SET medio_id = $1, categoria_id = $2, tipo = $3, fecha_movimiento = $4, descripcion = $5, monto = $6, egreso_fijo_id = $7
		WHERE id = $8 RETURNING id`

	var updatedID string
	err = tx.QueryRow(ctx, updateQuery, medioID, categoriaID, tipo, fechaMovimiento, descripcion, monto, egresoFijoID, id).Scan(&updatedID)
	if err != nil {
		return "", fmt.Errorf("error al actualizar movimiento: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return updatedID, nil
}

// DeleteMovimiento elimina un movimiento y revierte su impacto en el saldo
func (db *DB) DeleteMovimiento(ctx context.Context, id string) (string, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var medioID, tipo string
	var monto float64
	err = tx.QueryRow(ctx, `DELETE FROM movimiento WHERE id = $1 RETURNING medio_id, tipo, monto`, id).Scan(&medioID, &tipo, &monto)
	if err != nil {
		return "", fmt.Errorf("error al eliminar movimiento: %w", err)
	}

	// Reversión de saldo
	var revertDelta float64
	if tipo == "I" {
		revertDelta = -monto
	} else {
		revertDelta = monto
	}

	_, err = tx.Exec(ctx, `UPDATE saldo_actual SET saldo = saldo + $1 WHERE medio_id = $2`, revertDelta, medioID)
	if err != nil {
		return "", fmt.Errorf("error al ajustar saldo tras eliminar movimiento: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return id, nil
}
