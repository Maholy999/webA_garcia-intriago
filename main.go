package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"webA/internal/tareas"
)

func main() {
	reset := flag.Bool("reset", false, "borra las tablas y arranca con la base vacía")
	flag.Parse()

	dsn := "host=localhost port=5433 user=postgres password=Cnc0cnc0 dbname=notix sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	if *reset {
		// Primero la tabla de los muchos (Tarea), después la del uno (Usuario)
		db.Migrator().DropTable(&tareas.Tarea{}, &tareas.Usuario{})
	}

	err = db.Debug().AutoMigrate(&tareas.Usuario{}, &tareas.Tarea{})
	if err != nil {
		log.Fatal(err)
	}

	// 1. Invocación de la semilla (fase 1, paso 2)
	tareas.Sembrar(db)

	r := chi.NewRouter()

	// Conecta tu manejador de tareas al router
	(&tareas.Manejador{DB: db}).Rutas(r)

	log.Println("Servidor Notix escuchando en :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
