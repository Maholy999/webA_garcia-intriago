package tareas

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"webA/internal/respuesta"
)

// Manejador guarda la conexión. Todas las rutas la leen con m.DB.
type Manejador struct{ DB *gorm.DB }

// Rutas registra las rutas de este paquete
func (m *Manejador) Rutas(r chi.Router) {
	r.Post("/tareas", m.crear) // Paso 4: Crear tarea
	r.Get("/tareas", m.listar) // Paso 5: Listar tareas

	// Descomentar en Fase 2a:
	// r.Get("/tareas/{id}", m.verUno)
	// r.Put("/tareas/{id}", m.actualizar)
	// r.Delete("/tareas/{id}", m.borrar)
	// r.Get("/usuarios", m.listarUsuarios)
}

// crear procesa las peticiones POST para guardar una nueva Tarea
func (m *Manejador) crear(w http.ResponseWriter, r *http.Request) {
	// 1. Leer el JSON. Si viene mal formado, responde 400
	var t Tarea
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "json_invalido", "El cuerpo no es un JSON válido")
		return
	}

	t.ID = 0 // Asegura que PostgreSQL asigne la clave primaria

	// 2. Validar que el estado esté en la lista de estados válidos (fase 2b)
	if !estadosValidos[t.Estado] {
		respuesta.Error(w, http.StatusUnprocessableEntity, "estado_invalido", "Estado no válido")
		return
	}

	// 3. Guardar en la base de datos
	if err := m.DB.Debug().Create(&t).Error; err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "No se pudo guardar la tarea")
		return
	}

	// Responde 201 Created con la estructura t
	respuesta.Exito(w, http.StatusCreated, t)
}

// listar obtiene todas las tareas de la base de datos
func (m *Manejador) listar(w http.ResponseWriter, r *http.Request) {
	var lista []Tarea
	if err := m.DB.Debug().Find(&lista).Error; err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "No se pudieron obtener las tareas")
		return
	}

	respuesta.Exito(w, http.StatusOK, lista)
}
