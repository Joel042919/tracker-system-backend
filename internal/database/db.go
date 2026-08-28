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
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		return nil, fmt.Errorf("la variable DATABASE_URL no está definida")
	}

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("error al parsear configuración de la base de datos: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("error al conectar a la base de datos: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("error en el ping a la base de datos: %w", err)
	}

	log.Println("✅ Conexión exitosa a Neon DB")
	return &DB{Pool: pool}, nil
}

func (db *DB) Migrate() error {
	schema := `
		CREATE TABLE IF NOT EXISTS area (
		id SERIAL PRIMARY KEY,
		nombre VARCHAR(50) NOT NULL,
		descripcion TEXT,
		estado BOOLEAN DEFAULT true,
		created_at TIMESTAMPTZ DEFAULT now()
		);

		CREATE TABLE IF NOT EXISTS proyecto (
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

		CREATE TABLE IF NOT EXISTS metrica (
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

		CREATE TABLE IF NOT EXISTS proyecto_metrica (
		id SERIAL PRIMARY KEY,
		id_proyecto INT NOT NULL REFERENCES proyecto(id),
		id_metrica INT NOT NULL REFERENCES metrica(id),
		config_programacion JSONB NOT NULL,  
		activo BOOLEAN DEFAULT true,
		created_at TIMESTAMPTZ DEFAULT now()
		);

		CREATE TABLE IF NOT EXISTS registro_evaluacion (
		id SERIAL PRIMARY KEY,
		id_proyecto_metrica INT NOT NULL REFERENCES proyecto_metrica(id),
		fecha_evaluacion TIMESTAMPTZ NOT NULL,
		valores JSONB NOT NULL,
		notas TEXT,
		created_at TIMESTAMPTZ DEFAULT now()
		);

		CREATE TABLE IF NOT EXISTS rewards (
		id SERIAL PRIMARY KEY,
		reward VARCHAR(50) NOT NULL,
		points_need INT NOT NULL,
		description TEXT,
		estado BOOLEAN DEFAULT true
		);

		CREATE TABLE IF NOT EXISTS puntos_usados (
		id SERIAL PRIMARY KEY,
		id_reward INT NOT NULL REFERENCES rewards(id),
		reclaim_date TIMESTAMPTZ DEFAULT now()
		);

		CREATE TABLE IF NOT EXISTS task (
		id SERIAL PRIMARY KEY,
		taskname VARCHAR(50) NOT NULL,
		description TEXT,
		due_date TIMESTAMPTZ,
		status VARCHAR(10) NOT NULL CHECK (status IN ('do','doing','done')),
		points INT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT now()
		);

		CREATE TABLE IF NOT EXISTS proyecto_habito (
		id SERIAL PRIMARY KEY,
		id_proyecto INT NOT NULL REFERENCES proyecto(id),
		dias_semana JSONB NOT NULL,
		hora_objetivo TIME,
		points_por_completar INT NOT NULL DEFAULT 10,
		record_streak INT DEFAULT 0,
		best_streak INT DEFAULT 0,
		ultima_fecha_completada DATE,
		activo BOOLEAN DEFAULT true,
		created_at TIMESTAMPTZ DEFAULT now()
		);

		CREATE TABLE IF NOT EXISTS proyecto_tarea (
		id SERIAL PRIMARY KEY,
		id_proyecto_habito INT NOT NULL REFERENCES proyecto_habito(id),
		nombre VARCHAR(100) NOT NULL,
		descripcion TEXT,
		tiempo_estimado_minutos INT NOT NULL,
		orden INT NOT NULL,
		activo BOOLEAN DEFAULT true,
		created_at TIMESTAMPTZ DEFAULT now()
		);

		CREATE TABLE IF NOT EXISTS registro_habito (
		id SERIAL PRIMARY KEY,
		id_proyecto_habito INT NOT NULL REFERENCES proyecto_habito(id),
		fecha DATE NOT NULL,
		completado BOOLEAN DEFAULT false,
		fecha_completado TIMESTAMPTZ,
		points_ganados INT DEFAULT 0,
		streak_actual INT DEFAULT 0,
		notas TEXT,
		created_at TIMESTAMPTZ DEFAULT now(),
		UNIQUE (id_proyecto_habito, fecha)
		);

		CREATE TABLE IF NOT EXISTS registro_tarea (
		id SERIAL PRIMARY KEY,
		id_proyecto_tarea INT NOT NULL REFERENCES proyecto_tarea(id),
		id_registro_habito INT NOT NULL REFERENCES registro_habito(id),
		completado BOOLEAN DEFAULT false,
		fecha_completado TIMESTAMPTZ,
		tiempo_real_minutos INT,
		created_at TIMESTAMPTZ DEFAULT now(),
		UNIQUE (id_registro_habito, id_proyecto_tarea)
		);

		CREATE TABLE IF NOT EXISTS puntos_ganados (
		id SERIAL PRIMARY KEY,
		id_registro_evaluacion INT REFERENCES registro_evaluacion(id),
		id_task INT REFERENCES task(id),
		id_registro_habito INT REFERENCES registro_habito(id),
		points INT NOT NULL,
		tipo_origen VARCHAR(20) NOT NULL DEFAULT 'evaluacion',
		fecha_registro TIMESTAMPTZ DEFAULT now()
		);

		CREATE TABLE IF NOT EXISTS formulario (
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

		CREATE TABLE IF NOT EXISTS macros (
		idMacro UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		idFormulario UUID NOT NULL REFERENCES formulario(idFormulario),
		Calories INT,
		protein INT,
		carbs INT,
		fat INT,
		fiber INT,
		water NUMERIC(4,2)
		);

		CREATE TABLE IF NOT EXISTS dayliTrack (
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

		CREATE TABLE IF NOT EXISTS foodLog (
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

		CREATE TABLE IF NOT EXISTS medio (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		medio VARCHAR(50) NOT NULL,
		tipo_medio VARCHAR(50) NOT NULL,
		numero_cuenta VARCHAR(50),
		banco VARCHAR(50),
		estado BOOLEAN DEFAULT true,
		created_at TIMESTAMPTZ DEFAULT now()
		);

		CREATE TABLE IF NOT EXISTS saldo_actual (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		saldo DECIMAL(10,2) NOT NULL DEFAULT 0.00,
		medio_id UUID NOT NULL UNIQUE REFERENCES medio(id) ON DELETE CASCADE
		);

		CREATE UNIQUE INDEX IF NOT EXISTS idx_saldo_actual_medio_id ON saldo_actual (medio_id);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_registro_habito_unique ON registro_habito (id_proyecto_habito, fecha);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_registro_tarea_unique ON registro_tarea (id_registro_habito, id_proyecto_tarea);

		CREATE TABLE IF NOT EXISTS categoria (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		categoria VARCHAR(30) NOT NULL
		);

		CREATE TABLE IF NOT EXISTS egreso_fijo (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		razon VARCHAR(50) NOT NULL,
		descripcion TEXT,
		categoria_id UUID REFERENCES categoria(id) ON DELETE SET NULL,
		monto DECIMAL(10,2) NOT NULL,
		programacion_pago JSONB NOT NULL,
		recordatorio_dias_antes INT DEFAULT 3,
		activo BOOLEAN DEFAULT true,
		fecha_inicio DATE,
		fecha_fin DATE,
		created_at TIMESTAMPTZ DEFAULT now()
		);

		CREATE TABLE IF NOT EXISTS movimiento (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		medio_id UUID NOT NULL REFERENCES medio(id) ON DELETE CASCADE,
		categoria_id UUID REFERENCES categoria(id) ON DELETE SET NULL,
		tipo CHAR(1) NOT NULL CHECK (tipo IN ('I', 'E')),
		fecha_movimiento TIMESTAMPTZ NOT NULL DEFAULT now(),
		descripcion TEXT,
		monto DECIMAL(10,2) NOT NULL,
		egreso_fijo_id UUID REFERENCES egreso_fijo(id) ON DELETE SET NULL,
		created_at TIMESTAMPTZ DEFAULT now()
		);

		CREATE TABLE IF NOT EXISTS pago_programado (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		egreso_fijo_id UUID NOT NULL REFERENCES egreso_fijo(id) ON DELETE CASCADE,
		fecha_programada DATE NOT NULL,
		monto_esperado DECIMAL(10,2) NOT NULL,
		estado VARCHAR(20) DEFAULT 'pendiente',
		fecha_pago TIMESTAMPTZ,
		medio_id UUID REFERENCES medio(id) ON DELETE SET NULL,
		movimiento_id UUID REFERENCES movimiento(id) ON DELETE SET NULL,
		notas TEXT,
		created_at TIMESTAMPTZ DEFAULT now(),
		UNIQUE (egreso_fijo_id, fecha_programada)
		);

		CREATE TABLE IF NOT EXISTS presupuesto (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		nombre VARCHAR(100) NOT NULL,
		categoria_id UUID REFERENCES categoria(id) ON DELETE SET NULL,
		monto_limite DECIMAL(10,2) NOT NULL,
		periodo VARCHAR(20) NOT NULL,
		fecha_inicio DATE NOT NULL,
		fecha_fin DATE,
		activo BOOLEAN DEFAULT true,
		created_at TIMESTAMPTZ DEFAULT now()
		);

		CREATE TABLE IF NOT EXISTS alerta_pago (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		pago_programado_id UUID NOT NULL REFERENCES pago_programado(id) ON DELETE CASCADE,
		tipo_alerta VARCHAR(20) NOT NULL,
		mensaje TEXT NOT NULL,
		leida BOOLEAN DEFAULT false,
		fecha_alerta TIMESTAMPTZ DEFAULT now(),
		created_at TIMESTAMPTZ DEFAULT now()
		);

		CREATE INDEX IF NOT EXISTS idx_puntos_ganados_habito ON puntos_ganados (id_registro_habito);
		CREATE INDEX IF NOT EXISTS idx_puntos_ganados_task ON puntos_ganados (id_task);
		CREATE INDEX IF NOT EXISTS idx_puntos_ganados_eval ON puntos_ganados (id_registro_evaluacion);
	`
	_, err := db.Pool.Exec(context.Background(), schema)
	if err != nil {
		return fmt.Errorf("error ejecutando migraciones: %w", err)
	}
	log.Println("✅ Tablas migradas/verificadas correctamente")
	return nil
}
