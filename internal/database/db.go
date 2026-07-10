package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func Connect() (*DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL no esta configurada en el entorno")
	}

	// pgxpool maneja múltiples conexiones concurrentes de forma automática
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("error conectando a la base de datos: %w", err)
	}

	// Verificamos que la conexión esté viva
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("error en el ping a la base de datos: %w", err)
	}

	log.Println("✅ Conexión exitosa a Neon DB")
	return &DB{Pool: pool}, nil
}

func (db *DB) Migrate() error {
	schema := `
		CREATE TABLE area (
		id SERIAL PRIMARY KEY,
		nombre VARCHAR(50) NOT NULL,
		descripcion TEXT,
		estado BOOLEAN DEFAULT true,
		created_at TIMESTAMPTZ DEFAULT now()
		);

		CREATE TABLE proyecto (
		id SERIAL PRIMARY KEY,
		id_area INT NOT NULL REFERENCES area(id),
		nombre VARCHAR(100) NOT NULL,
		descripcion TEXT,
		meta TEXT,
		fecha_inicio DATE NOT NULL,
		fecha_fin_planeado DATE NOT NULL,
		fecha_fin_real DATE,
		estado BOOLEAN DEFAULT true,
		created_at TIMESTAMPTZ DEFAULT now()
		);

		CREATE TABLE metrica (
		id SERIAL PRIMARY KEY,
		id_area INT NOT NULL REFERENCES area(id),
		nombre VARCHAR(100) NOT NULL,
		descripcion TEXT,
		schema_esperado JSONB NOT NULL,
		resultados_esperado JSONB,
		points INT NOT NULL,
		estado BOOLEAN DEFAULT true,
		created_at TIMESTAMPTZ DEFAULT now()
		);

		CREATE TABLE proyecto_metrica (
		id SERIAL PRIMARY KEY,
		id_proyecto INT NOT NULL REFERENCES proyecto(id),
		id_metrica INT NOT NULL REFERENCES metrica(id),
		config_programacion JSONB NOT NULL,  
		activo BOOLEAN DEFAULT true,
		created_at TIMESTAMPTZ DEFAULT now()
		);

		CREATE TABLE registro_evaluacion (
		id SERIAL PRIMARY KEY,
		id_proyecto_metrica INT NOT NULL REFERENCES proyecto_metrica(id),
		fecha_evaluacion TIMESTAMPTZ NOT NULL,
		valores JSONB NOT NULL,
		notas TEXT,
		created_at TIMESTAMPTZ DEFAULT now()
		);

		CREATE TABLE rewards (
		id SERIAL PRIMARY KEY,
		reward VARCHAR(50) NOT NULL,
		points_need INT NOT NULL,
		description TEXT,
		estado BOOLEAN DEFAULT true
		);

		CREATE TABLE puntos_usados (
		id SERIAL PRIMARY KEY,
		id_reward INT NOT NULL REFERENCES rewards(id),
		reclaim_date TIMESTAMPTZ DEFAULT now()
		);

		CREATE TABLE task (
		id SERIAL PRIMARY KEY,
		taskname VARCHAR(50) NOT NULL,
		description TEXT,
		due_date TIMESTAMPTZ,
		status VARCHAR(10) NOT NULL CHECK (status IN ('do','doing','done')),
		points INT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT now()
		);

		CREATE TABLE puntos_ganados (
		id SERIAL PRIMARY KEY,
		id_registro_evaluacion INT REFERENCES registro_evaluacion(id),
		id_task INT REFERENCES task(id),
		points INT NOT NULL,
		fecha_registro TIMESTAMPTZ DEFAULT now()
		);

		CREATE TABLE formulario (
		idFormulario UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		gender VARCHAR(6),
		edad INT,
		peso NUMERIC(5,2),
		altura INT,
		nivelActividad INT,
		cuello NUMERIC(5,2),
		cintura NUMERIC(5,2),
		cadera NUMERIC(5,2),
		meta VARCHAR(9),
		velocidadKgSemana NUMERIC(2,1),
		fechaRegistro DATE DEFAULT CURRENT_DATE,
		active BOOLEAN DEFAULT true
		);

		CREATE TABLE macros (
		idMacro UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		idFormulario UUID NOT NULL REFERENCES formulario(idFormulario),
		Calories INT,
		protein INT,
		carbs INT,
		fat INT,
		fiber INT,
		water NUMERIC(4,2)
		);

		CREATE TABLE dayliTrack (
		idDayliTrack UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		idMacro UUID REFERENCES macros(idMacro),
		caloriesCount INT,
		protein INT,
		carbs INT,
		fat INT,
		fiber INT,
		water INT,
		dateTrack DATE NOT NULL DEFAULT CURRENT_DATE
		);

		CREATE TABLE foodLog (
		idFoodLog UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		idDayliTrack UUID NOT NULL REFERENCES dayliTrack(idDayliTrack),
		type_meal VARCHAR(10) CHECK (type_meal IN ('breakfast','lunch','dinner','snack')),
		food TEXT,
		calories INT,
		protein INT,
		carbs INT,
		fat INT,
		fiber INT,
		created_at TIMESTAMPTZ DEFAULT now()
		);
	`
	_, err := db.Pool.Exec(context.Background(), schema)
	if err != nil {
		return fmt.Errorf("error ejecutando migraciones: %w", err)
	}
	log.Println("✅ Tablas migradas/verificadas correctamente")
	return nil
}
