package main

import (
	"log"
	"net/http"
	"tracker/internal/database"
	"tracker/internal/handlers"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Advertencia: No se encontró archivo .env, usando variables del sistema")
	}
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Error crítico: %v", err)
	}
	defer db.Pool.Close()

	// Inicializar handlers inyectando la dependencia de la DB
	nutritionHandler := &handlers.NutritionHandler{DB: db}
	areaHandler := &handlers.AreaHandler{DB: db}
	proyectoHandler := &handlers.ProyectoHandler{DB: db}
	metricaHandler := &handlers.MetricaHandler{DB: db}
	pmHandler := &handlers.ProyectoMetricaHandler{DB: db}
	reHandler := &handlers.RegistroEvaluacionHandler{DB: db}
	rewardHandler := &handlers.RewardHandler{DB: db}
	taskHandler := &handlers.TaskHandler{DB: db}
	puHandler := &handlers.PuntosUsadosHandler{DB: db}
	pgHandler := &handlers.PuntosGanadosHandler{DB: db}
	prHandler := &handlers.PointReviewHandler{DB: db}
	formHandler := &handlers.FormularioHandler{DB: db}
	macrosHandler := &handlers.MacrosHandler{DB: db}
	dayliHandler := &handlers.DayliTrackHandler{DB: db}
	foodLogHandler := &handlers.FoodLogHandler{DB: db}
	proyectoHabitoHandler := &handlers.ProyectoHabitoHandler{DB: db}
	proyectoTareaHandler := &handlers.ProyectoTareaHandler{DB: db}
	registroHabitoHandler := &handlers.RegistroHabitoHandler{DB: db}
	registroTareaHandler := &handlers.RegistroTareaHandler{DB: db}

	//Ejecutar migraciones
	/*if err := db.Migrate(); err != nil {
		log.Fatalf("Error en migración: %v", err)
	}*/
	mux := http.NewServeMux()

	// Point Review
	mux.HandleFunc("GET /api/point-review", prHandler.GetTotal)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("API Tracker funcionando al 100% 🚀"))
	})

	mux.HandleFunc("POST /api/nutricion", nutritionHandler.CreateProfile)

	//Areas
	mux.HandleFunc("POST /api/areas", areaHandler.Create)
	mux.HandleFunc("GET /api/areas", areaHandler.GetAll)
	mux.HandleFunc("GET /api/areas/{id}", areaHandler.GetArea)
	mux.HandleFunc("PUT /api/areas/{id}", areaHandler.UpdateArea)

	//Descripcion
	mux.HandleFunc("POST /api/proyectos", proyectoHandler.Create)
	mux.HandleFunc("GET /api/proyectos", proyectoHandler.GetAll)
	mux.HandleFunc("GET /api/proyectos/{id}", proyectoHandler.GetProyecto)
	mux.HandleFunc("PUT /api/proyectos/{id}", proyectoHandler.UpdateProyecto)
	mux.HandleFunc("DELETE /api/proyectos/{id}", proyectoHandler.DeleteProyecto)

	//Metrica
	mux.HandleFunc("POST /api/metricas", metricaHandler.Create)
	mux.HandleFunc("GET /api/metricas", metricaHandler.GetAll)
	mux.HandleFunc("GET /api/metricas/{id}", metricaHandler.GetMetrica)
	mux.HandleFunc("PUT /api/metricas/{id}", metricaHandler.UpdateMetrica)
	mux.HandleFunc("DELETE /api/metricas/{id}", metricaHandler.DeleteMetrica)

	//Proyecto-Metrica
	mux.HandleFunc("POST /api/proyecto-metricas", pmHandler.Create)
	mux.HandleFunc("GET /api/proyecto-metricas", pmHandler.GetAll)
	mux.HandleFunc("GET /api/proyecto-metricas/{id}", pmHandler.GetProyectoMetrica)
	mux.HandleFunc("PUT /api/proyecto-metricas/{id}", pmHandler.UpdateProyectoMetrica)
	mux.HandleFunc("DELETE /api/proyecto-metricas/{id}", pmHandler.DeleteProyectoMetrica)

	//Registro evaluacion
	mux.HandleFunc("POST /api/registro-evaluaciones", reHandler.Create)
	mux.HandleFunc("GET /api/registro-evaluaciones", reHandler.GetAll)
	mux.HandleFunc("GET /api/registro-evaluaciones/{id}", reHandler.GetRegistroEvaluacion)
	mux.HandleFunc("PUT /api/registro-evaluaciones/{id}", reHandler.UpdateRegistroEvaluacion)
	mux.HandleFunc("DELETE /api/registro-evaluaciones/{id}", reHandler.DeleteRegistroEvaluacion)

	//Rewards
	mux.HandleFunc("POST /api/rewards", rewardHandler.Create)
	mux.HandleFunc("GET /api/rewards", rewardHandler.GetAll)
	mux.HandleFunc("GET /api/rewards/{id}", rewardHandler.GetReward)
	mux.HandleFunc("PUT /api/rewards/{id}", rewardHandler.UpdateReward)
	mux.HandleFunc("DELETE /api/rewards/{id}", rewardHandler.DeleteReward)

	// Puntos usados
	mux.HandleFunc("POST /api/puntos-usados", puHandler.Create)
	mux.HandleFunc("GET /api/puntos-usados", puHandler.GetAll)
	mux.HandleFunc("GET /api/puntos-usados/{id}", puHandler.GetPuntosUsado)
	mux.HandleFunc("PUT /api/puntos-usados/{id}", puHandler.UpdatePuntosUsado)
	mux.HandleFunc("DELETE /api/puntos-usados/{id}", puHandler.DeletePuntosUsado)

	// Puntos ganados
	mux.HandleFunc("POST /api/puntos-ganados", pgHandler.Create)
	mux.HandleFunc("GET /api/puntos-ganados", pgHandler.GetAll)
	mux.HandleFunc("GET /api/puntos-ganados/{id}", pgHandler.GetPuntosGanado)
	mux.HandleFunc("PUT /api/puntos-ganados/{id}", pgHandler.UpdatePuntosGanado)
	mux.HandleFunc("DELETE /api/puntos-ganados/{id}", pgHandler.DeletePuntosGanado)

	// Tasks
	mux.HandleFunc("POST /api/tasks", taskHandler.Create)
	mux.HandleFunc("GET /api/tasks", taskHandler.GetAll)
	mux.HandleFunc("GET /api/tasks/{id}", taskHandler.GetTask)
	mux.HandleFunc("PUT /api/tasks/{id}", taskHandler.UpdateTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", taskHandler.DeleteTask)

	//Formularios
	mux.HandleFunc("POST /api/formularios", formHandler.Create)
	mux.HandleFunc("GET /api/formularios", formHandler.GetAll)
	mux.HandleFunc("GET /api/formularios/{id}", formHandler.GetFormulario)
	mux.HandleFunc("PUT /api/formularios/{id}", formHandler.UpdateFormulario)
	mux.HandleFunc("DELETE /api/formularios/{id}", formHandler.DeleteFormulario)

	// Macros
	mux.HandleFunc("POST /api/macros", macrosHandler.Create)
	mux.HandleFunc("GET /api/macros", macrosHandler.GetAll)
	mux.HandleFunc("GET /api/macros/{id}", macrosHandler.GetMacro)
	mux.HandleFunc("PUT /api/macros/{id}", macrosHandler.UpdateMacro)
	mux.HandleFunc("DELETE /api/macros/{id}", macrosHandler.DeleteMacro)

	// DayliTrack
	mux.HandleFunc("POST /api/dayli-tracks", dayliHandler.Create)
	mux.HandleFunc("GET /api/dayli-tracks", dayliHandler.GetAll)
	mux.HandleFunc("GET /api/dayli-tracks/{id}", dayliHandler.GetDayliTrack)
	mux.HandleFunc("PUT /api/dayli-tracks/{id}", dayliHandler.UpdateDayliTrack)
	mux.HandleFunc("DELETE /api/dayli-tracks/{id}", dayliHandler.DeleteDayliTrack)

	//FOOD LOGS
	mux.HandleFunc("POST /api/food-logs", foodLogHandler.Create)
	mux.HandleFunc("GET /api/food-logs", foodLogHandler.GetAll)
	mux.HandleFunc("GET /api/food-logs/{id}", foodLogHandler.GetFoodLog)
	mux.HandleFunc("PUT /api/food-logs/{id}", foodLogHandler.UpdateFoodLog)
	mux.HandleFunc("DELETE /api/food-logs/{id}", foodLogHandler.DeleteFoodLog)

	// Proyecto Habitos
	mux.HandleFunc("POST /api/proyecto-habitos", proyectoHabitoHandler.Create)
	mux.HandleFunc("GET /api/proyecto-habitos", proyectoHabitoHandler.GetAll)
	mux.HandleFunc("GET /api/proyecto-habitos/{id}", proyectoHabitoHandler.GetByID)
	mux.HandleFunc("PUT /api/proyecto-habitos/{id}", proyectoHabitoHandler.Update)
	mux.HandleFunc("DELETE /api/proyecto-habitos/{id}", proyectoHabitoHandler.Delete)

	// Proyecto Tareas (Mini-tareas)
	mux.HandleFunc("POST /api/proyecto-tareas", proyectoTareaHandler.Create)
	mux.HandleFunc("GET /api/proyecto-tareas", proyectoTareaHandler.GetAll)
	mux.HandleFunc("GET /api/proyecto-tareas/{id}", proyectoTareaHandler.GetByID)
	mux.HandleFunc("PUT /api/proyecto-tareas/{id}", proyectoTareaHandler.Update)
	mux.HandleFunc("DELETE /api/proyecto-tareas/{id}", proyectoTareaHandler.Delete)

	// Registro Habitos (Tracking Diario)
	mux.HandleFunc("POST /api/registro-habitos", registroHabitoHandler.Create)
	mux.HandleFunc("GET /api/registro-habitos", registroHabitoHandler.GetAll)
	mux.HandleFunc("GET /api/registro-habitos/{id}", registroHabitoHandler.GetByID)
	mux.HandleFunc("PUT /api/registro-habitos/{id}", registroHabitoHandler.Update)
	mux.HandleFunc("DELETE /api/registro-habitos/{id}", registroHabitoHandler.Delete)

	// Registro Tareas (Completitud de Mini-tareas)
	mux.HandleFunc("POST /api/registro-tareas", registroTareaHandler.Create)
	mux.HandleFunc("GET /api/registro-tareas", registroTareaHandler.GetAll)
	mux.HandleFunc("GET /api/registro-tareas/{id}", registroTareaHandler.GetByID)
	mux.HandleFunc("PUT /api/registro-tareas/{id}", registroTareaHandler.Update)
	mux.HandleFunc("DELETE /api/registro-tareas/{id}", registroTareaHandler.Delete)

	// CORS middleware
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	puerto := ":8080"
	log.Printf("Servidor corriendo en el puerto %s", puerto)
	if err := http.ListenAndServe(puerto, corsMiddleware(mux)); err != nil {
		log.Fatalf("Error iniciando el servidor: %v", err)
	}
}
