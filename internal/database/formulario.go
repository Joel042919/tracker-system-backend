package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Formulario representa la tabla formulario
type Formulario struct {
	IDFormulario      string     `json:"id_formulario"` // UUID
	Gender            *string    `json:"gender"`
	Edad              *int       `json:"edad"`
	Peso              *float64   `json:"peso"` // numeric(5,2)
	Altura            *int       `json:"altura"`
	NivelActividad    *int       `json:"nivelActividad"`
	Cuello            *float64   `json:"cuello"`
	Cintura           *float64   `json:"cintura"`
	Cadera            *float64   `json:"cadera"`
	Meta              *string    `json:"meta"`
	VelocidadKgSemana *float64   `json:"velocidadKgSemana"` // numeric(2,1)
	FechaRegistro     *time.Time `json:"fechaRegistro"`     // DATE, puede ser null (default CURRENT_DATE)
	Active            *bool      `json:"active"`
}

// CreateFormulario inserta un nuevo formulario (el UUID lo genera la BD)
func (db *DB) CreateFormulario(ctx context.Context, gender *string, edad *int, peso *float64, altura *int,
	nivelActividad *int, cuello *float64, cintura *float64, cadera *float64, meta *string,
	velocidadKgSemana *float64, fechaRegistro *time.Time, active *bool) (string, error) {

	query := `INSERT INTO formulario (gender, edad, peso, altura, nivelActividad, cuello, cintura, cadera,
		meta, velocidadKgSemana, fechaRegistro, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING idFormulario`

	var id string
	err := db.Pool.QueryRow(ctx, query,
		gender, edad, peso, altura, nivelActividad, cuello, cintura, cadera,
		meta, velocidadKgSemana, fechaRegistro, active,
	).Scan(&id)
	return id, err
}

// GetFormularios obtiene todos los formularios activos
func (db *DB) GetFormularios(ctx context.Context) ([]Formulario, error) {
	query := `SELECT idFormulario, gender, edad, peso, altura, nivelActividad, cuello, cintura, cadera,
		meta, velocidadKgSemana, fechaRegistro, active
		FROM formulario WHERE active = true ORDER BY fechaRegistro DESC`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []Formulario
	for rows.Next() {
		var f Formulario
		if err := rows.Scan(&f.IDFormulario, &f.Gender, &f.Edad, &f.Peso, &f.Altura,
			&f.NivelActividad, &f.Cuello, &f.Cintura, &f.Cadera, &f.Meta,
			&f.VelocidadKgSemana, &f.FechaRegistro, &f.Active); err != nil {
			return nil, err
		}
		lista = append(lista, f)
	}
	return lista, nil
}

// GetFormulario busca un formulario por su UUID (solo si está activo)
func (db *DB) GetFormulario(ctx context.Context, id string) (Formulario, error) {
	query := `SELECT idFormulario, gender, edad, peso, altura, nivelActividad, cuello, cintura, cadera,
		meta, velocidadKgSemana, fechaRegistro, active
		FROM formulario WHERE idFormulario = $1 AND active = true`

	var f Formulario
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&f.IDFormulario, &f.Gender, &f.Edad, &f.Peso, &f.Altura,
		&f.NivelActividad, &f.Cuello, &f.Cintura, &f.Cadera, &f.Meta,
		&f.VelocidadKgSemana, &f.FechaRegistro, &f.Active,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Formulario{}, fmt.Errorf("formulario con id %s no encontrado: %w", id, err)
		}
		return Formulario{}, err
	}
	return f, nil
}

// UpdateFormulario actualiza todos los campos editables
func (db *DB) UpdateFormulario(ctx context.Context, id string, gender *string, edad *int, peso *float64,
	altura *int, nivelActividad *int, cuello *float64, cintura *float64, cadera *float64,
	meta *string, velocidadKgSemana *float64, fechaRegistro *time.Time, active *bool) (string, error) {

	query := `UPDATE formulario SET
		gender = $1, edad = $2, peso = $3, altura = $4, nivelActividad = $5,
		cuello = $6, cintura = $7, cadera = $8, meta = $9, velocidadKgSemana = $10,
		fechaRegistro = $11, active = $12
		WHERE idFormulario = $13 RETURNING idFormulario`

	var updatedID string
	err := db.Pool.QueryRow(ctx, query,
		gender, edad, peso, altura, nivelActividad,
		cuello, cintura, cadera, meta, velocidadKgSemana,
		fechaRegistro, active, id,
	).Scan(&updatedID)
	return updatedID, err
}

// DeleteFormulario borrado lógico (active = false)
func (db *DB) DeleteFormulario(ctx context.Context, id string) (string, error) {
	query := `UPDATE formulario SET active = false WHERE idFormulario = $1 RETURNING idFormulario`
	var deletedID string
	err := db.Pool.QueryRow(ctx, query, id).Scan(&deletedID)
	return deletedID, err
}

// DeactivateAllFormularios sets active = false for all active formularios
func (db *DB) DeactivateAllFormularios(ctx context.Context) error {
	query := `UPDATE formulario SET active = false WHERE active = true`
	_, err := db.Pool.Exec(ctx, query)
	return err
}
