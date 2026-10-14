package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"webA/internal/config"
	"webA/internal/middleware"
	"webA/internal/sesion"
	"webA/internal/tareas"
)

func main() {
	// 1. Cargar la configuración desde el entorno (.env)
	cfg, err := config.Cargar()
	if err != nil {
		log.Fatal("configuración: ", err)
	}
	tokens, err := sesion.NuevoTokens(cfg.JWTSecret)
	if err != nil {
		log.Fatal("sesión: ", err)
	}

	// Se conserva la bandera -reset de la semana 3
	reset := flag.Bool("reset", false, "borra las tablas y arranca con la base vacía")
	flag.Parse()

	// 2. Conectar usando cfg.DatabaseURL (sin secretos ni DSN escrito en código)
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("no se pudo conectar: ", err)
	}

	// Se conserva la lógica de reset
	if *reset {
		if err := db.Migrator().DropTable(&tareas.Tarea{}, &sesion.Cuenta{}, &tareas.Usuario{}); err != nil {
			log.Fatal("no se pudieron borrar las tablas: ", err)
		}
	}

	err = db.Debug().AutoMigrate(&tareas.Usuario{}, &tareas.Tarea{}, &sesion.Cuenta{})
	if err != nil {
		log.Fatal("no se pudo migrar: ", err)
	}

	// Se conserva la semilla de la semana 3
	tareas.Sembrar(db)
	if err := sesion.Sembrar(db); err != nil {
		log.Fatal("no se pudieron sembrar las cuentas de sesión: ", err)
	}

	// 3. Rutas
	r := chi.NewRouter()
	r.Use(middleware.Registro)
	r.Use(middleware.Recuperacion)

	(&tareas.Manejador{DB: db}).Rutas(r, tokens)

	// 4. Servidor configurado con cfg.Puerto y cfg.TiempoEspera
	servidor := &http.Server{
		Addr:         ":" + cfg.Puerto,
		Handler:      r,
		ReadTimeout:  cfg.TiempoEspera,
		WriteTimeout: cfg.TiempoEspera,
	}

	log.Println("Servidor Notix escuchando en el puerto", cfg.Puerto)
	log.Fatal(servidor.ListenAndServe())
}
