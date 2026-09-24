package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"webA/internal/config"
	"webA/internal/tareas"
)

func main() {
	// 1. Cargar la configuración desde el entorno (.env)
	cfg, err := config.Cargar()
	if err != nil {
		log.Fatal("configuración: ", err)
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
		db.Migrator().DropTable(&tareas.Tarea{}, &tareas.Usuario{})
	}

	err = db.Debug().AutoMigrate(&tareas.Usuario{}, &tareas.Tarea{})
	if err != nil {
		log.Fatal("no se pudo migrar: ", err)
	}

	// Se conserva la semilla de la semana 3
	tareas.Sembrar(db)

	// 3. Rutas
	r := chi.NewRouter()

	(&tareas.Manejador{DB: db}).Rutas(r)

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
