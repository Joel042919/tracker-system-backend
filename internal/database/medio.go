package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Medio representa una cuenta bancaria, billetera digital, tarjeta o efectivo
type Medio struct {
	ID           string    `json:"id"`
	Medio        string    `json:"medio"`
	TipoMedio    string    `json:"tipo_medio"` // 'yape', 'efectivo', 'cuenta_bancaria', 'tarjeta_credito', etc.
	NumeroCuenta *string   `json:"numero_cuenta"`
	Banco        *string   `json:"banco"`
	Estado       bool      `json:"estado"`
	SaldoActual  float64   `json:"saldo_actual"` // Saldo obtenido de la tabla saldo_actual
	CreatedAt    time.Time `json:"created_at"`
}

// SaldoActual representa el saldo monetario de un medio
type SaldoActual struct {
	ID      string  `json:"id"`
	Saldo   float64 `json:"saldo"`
	MedioID string  `json:"medio_id"`
}

// CreateMedio inserta un nuevo medio y su saldo inicial en una sola transacción
func (db *DB) CreateMedio(ctx context.Context, id *string, nombre, tipoMedio string, numeroCuenta, banco *string, saldoInicial float64) (string, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var medioID string
	var medioQuery string
	if id != nil && *id != "" {
		medioQuery = `INSERT INTO medio (id, medio, tipo_medio, numero_cuenta, banco)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE SET medio = EXCLUDED.medio, tipo_medio = EXCLUDED.tipo_medio, numero_cuenta = EXCLUDED.numero_cuenta, banco = EXCLUDED.banco
			RETURNING id`
		err = tx.QueryRow(ctx, medioQuery, *id, nombre, tipoMedio, numeroCuenta, banco).Scan(&medioID)
	} else {
		medioQuery = `INSERT INTO medio (medio, tipo_medio, numero_cuenta, banco)
			VALUES ($1, $2, $3, $4) RETURNING id`
		err = tx.QueryRow(ctx, medioQuery, nombre, tipoMedio, numeroCuenta, banco).Scan(&medioID)
	}
	if err != nil {
		return "", fmt.Errorf("error al insertar medio: %w", err)
	}

	var saldoExists bool
	_ = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM saldo_actual WHERE medio_id = $1::uuid)", medioID).Scan(&saldoExists)
	if saldoExists {
		_, err = tx.Exec(ctx, "UPDATE saldo_actual SET saldo = $1 WHERE medio_id = $2::uuid", saldoInicial, medioID)
	} else {
		_, err = tx.Exec(ctx, "INSERT INTO saldo_actual (id, saldo, medio_id) VALUES (gen_random_uuid(), $1, $2::uuid)", saldoInicial, medioID)
	}
	if err != nil {
		return "", fmt.Errorf("error al inicializar saldo_actual: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return medioID, nil
}

// GetMedios obtiene todos los medios activos con su saldo_actual
func (db *DB) GetMedios(ctx context.Context) ([]Medio, error) {
	query := `SELECT m.id, m.medio, m.tipo_medio, m.numero_cuenta, m.banco, m.estado, 
		COALESCE(s.saldo, 0.00) as saldo_actual, m.created_at
		FROM medio m
		LEFT JOIN saldo_actual s ON m.id = s.medio_id
		WHERE m.estado = true
		ORDER BY m.created_at ASC`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []Medio
	for rows.Next() {
		var m Medio
		if err := rows.Scan(
			&m.ID, &m.Medio, &m.TipoMedio, &m.NumeroCuenta, &m.Banco, &m.Estado,
			&m.SaldoActual, &m.CreatedAt,
		); err != nil {
			return nil, err
		}
		lista = append(lista, m)
	}
	return lista, nil
}

// GetMedio busca un medio por ID con su saldo actual
func (db *DB) GetMedio(ctx context.Context, id string) (Medio, error) {
	query := `SELECT m.id, m.medio, m.tipo_medio, m.numero_cuenta, m.banco, m.estado, 
		COALESCE(s.saldo, 0.00) as saldo_actual, m.created_at
		FROM medio m
		LEFT JOIN saldo_actual s ON m.id = s.medio_id
		WHERE m.id = $1`

	var m Medio
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.Medio, &m.TipoMedio, &m.NumeroCuenta, &m.Banco, &m.Estado,
		&m.SaldoActual, &m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Medio{}, fmt.Errorf("medio no encontrado: %w", err)
		}
		return Medio{}, err
	}
	return m, nil
}

// UpdateMedio actualiza datos informativos del medio
func (db *DB) UpdateMedio(ctx context.Context, id string, nombre, tipoMedio string, numeroCuenta, banco *string, estado bool) (string, error) {
	query := `UPDATE medio
		SET medio = $1, tipo_medio = $2, numero_cuenta = $3, banco = $4, estado = $5
		WHERE id = $6 RETURNING id`

	var updatedID string
	err := db.Pool.QueryRow(ctx, query, nombre, tipoMedio, numeroCuenta, banco, estado, id).Scan(&updatedID)
	return updatedID, err
}

// SetSaldoMedio permite ajustar manualmente el saldo de un medio
func (db *DB) SetSaldoMedio(ctx context.Context, medioID string, nuevoSaldo float64) error {
	var saldoExists bool
	_ = db.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM saldo_actual WHERE medio_id = $1::uuid)", medioID).Scan(&saldoExists)
	if saldoExists {
		_, err := db.Pool.Exec(ctx, "UPDATE saldo_actual SET saldo = $1 WHERE medio_id = $2::uuid", nuevoSaldo, medioID)
		return err
	}
	_, err := db.Pool.Exec(ctx, "INSERT INTO saldo_actual (id, saldo, medio_id) VALUES (gen_random_uuid(), $1, $2::uuid)", nuevoSaldo, medioID)
	return err
}

// DeleteMedio realiza soft delete (estado = false)
func (db *DB) DeleteMedio(ctx context.Context, id string) (string, error) {
	query := `UPDATE medio SET estado = false WHERE id = $1 RETURNING id`
	var deletedID string
	err := db.Pool.QueryRow(ctx, query, id).Scan(&deletedID)
	return deletedID, err
}
