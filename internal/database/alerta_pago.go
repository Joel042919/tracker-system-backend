package database

import (
	"context"
	"time"
)

// AlertaPago representa una notificación o recordatorio de pago
type AlertaPago struct {
	ID               string    `json:"id"`
	PagoProgramadoID string    `json:"pago_programado_id"`
	TipoAlerta       string    `json:"tipo_alerta"` // 'recordatorio', 'vencimiento', 'vencido'
	Mensaje          string    `json:"mensaje"`
	Leida            bool      `json:"leida"`
	FechaAlerta      time.Time `json:"fecha_alerta"`
	CreatedAt        time.Time `json:"created_at"`
}

// CreateAlertaPago inserta una alerta de pago
func (db *DB) CreateAlertaPago(ctx context.Context, pagoProgramadoID, tipoAlerta, mensaje string, fechaAlerta time.Time) (string, error) {
	query := `INSERT INTO alerta_pago (pago_programado_id, tipo_alerta, mensaje, fecha_alerta)
		VALUES ($1, $2, $3, $4) RETURNING id`

	var id string
	err := db.Pool.QueryRow(ctx, query, pagoProgramadoID, tipoAlerta, mensaje, fechaAlerta).Scan(&id)
	return id, err
}

// GetAlertasPago obtiene las alertas ordenadas por fecha
func (db *DB) GetAlertasPago(ctx context.Context) ([]AlertaPago, error) {
	query := `SELECT id, pago_programado_id, tipo_alerta, mensaje, leida, fecha_alerta, created_at
		FROM alerta_pago ORDER BY fecha_alerta DESC, created_at DESC`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []AlertaPago
	for rows.Next() {
		var a AlertaPago
		if err := rows.Scan(
			&a.ID, &a.PagoProgramadoID, &a.TipoAlerta, &a.Mensaje, &a.Leida,
			&a.FechaAlerta, &a.CreatedAt,
		); err != nil {
			return nil, err
		}
		lista = append(lista, a)
	}
	return lista, nil
}

// MarcarAlertaLeida actualiza el estado de una alerta a leída
func (db *DB) MarcarAlertaLeida(ctx context.Context, id string, leida bool) (string, error) {
	query := `UPDATE alerta_pago SET leida = $1 WHERE id = $2 RETURNING id`
	var updatedID string
	err := db.Pool.QueryRow(ctx, query, leida, id).Scan(&updatedID)
	return updatedID, err
}

// DeleteAlertaPago elimina una alerta
func (db *DB) DeleteAlertaPago(ctx context.Context, id string) (string, error) {
	query := `DELETE FROM alerta_pago WHERE id = $1 RETURNING id`
	var deletedID string
	err := db.Pool.QueryRow(ctx, query, id).Scan(&deletedID)
	return deletedID, err
}
