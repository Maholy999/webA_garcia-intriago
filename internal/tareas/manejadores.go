package tareas

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"webA/internal/respuesta"
)

// Manejador guarda la conexión a la base de datos.
type Manejador struct{ DB *gorm.DB }

// Rutas registra todas las rutas HTTP del paquete
func (m *Manejador) Rutas(r chi.Router) {
	// Fase 1
	r.Post("/tareas", m.crear)
	r.Get("/tareas", m.listar)

	// Fase 2a: CRUD completo para Tarea
	r.Get("/tareas/{id}", m.verUno)
	r.Put("/tareas/{id}", m.actualizar)
	r.Delete("/tareas/{id}", m.borrar)

	// Fase 2c: Listado optimizado sin problema N+1
	r.Get("/usuarios", m.listarUsuarios)
}

// crear procesa las peticiones POST para guardar una nueva Tarea
func (m *Manejador) crear(w http.ResponseWriter, r *http.Request) {
	var t Tarea

	// 1. Validar lectura de JSON (HTTP 400)
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "json_invalido", "El cuerpo no es un JSON válido")
		return
	}

	t.ID = 0 // Asegura que PostgreSQL asigne la clave primaria

	// 2. Validar que el estado pertenezca al mapa estadosValidos (HTTP 422)
	if !estadosValidos[t.Estado] {
		respuesta.Error(w, http.StatusUnprocessableEntity, "estado_invalido", "Estado no válido")
		return
	}

	// 3. Regla de negocio adicional: Validar existencia del Usuario asociado (HTTP 422)
	var u Usuario
	if err := m.DB.First(&u, t.UsuarioID).Error; err != nil {
		respuesta.Error(w, http.StatusUnprocessableEntity, "usuario_inexistente", "El usuario especificado no existe")
		return
	}

	if err := m.DB.Debug().Create(&t).Error; err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "No se pudo guardar la tarea")
		return
	}

	respuesta.Exito(w, http.StatusCreated, t)
}

// listar obtiene todas las tareas de la base de datos con soporte para filtrado seguro (Fase 2d)
func (m *Manejador) listar(w http.ResponseWriter, r *http.Request) {
	var lista []Tarea
	query := m.DB.Debug()

	// 1. Obtener parámetro "estado" de la URL (ej. /tareas?estado=completada)
	estadoFiltro := r.URL.Query().Get("estado")

	// 2. Si se envió parámetro, se valida el mapa y se aplica el filtro parametrizado
	if estadoFiltro != "" {
		if !estadosValidos[estadoFiltro] {
			respuesta.Error(w, http.StatusUnprocessableEntity, "estado_invalido", "Estado no válido para filtrado")
			return
		}

		// Previene inyección SQL mediante uso de placeholder '?'
		query = query.Where("estado = ?", estadoFiltro)
	}

	// 3. Ejecutar la consulta final
	if err := query.Find(&lista).Error; err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "No se pudieron obtener las tareas")
		return
	}

	respuesta.Exito(w, http.StatusOK, lista)
}

// verUno obtiene una tarea específica por su ID (Fase 2a)
func (m *Manejador) verUno(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		respuesta.Error(w, http.StatusBadRequest, "id_invalido", "El ID debe ser un número entero positivo")
		return
	}

	var t Tarea
	if err := m.DB.Debug().First(&t, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respuesta.Error(w, http.StatusNotFound, "no_encontrado", "La tarea no existe")
			return
		}
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "Error al buscar la tarea")
		return
	}

	respuesta.Exito(w, http.StatusOK, t)
}

// actualizar modifica los campos de una tarea existente (Fase 2a)
func (m *Manejador) actualizar(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		respuesta.Error(w, http.StatusBadRequest, "id_invalido", "El ID debe ser un número entero positivo")
		return
	}

	var t Tarea
	if err := m.DB.Debug().First(&t, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respuesta.Error(w, http.StatusNotFound, "no_encontrado", "La tarea no existe")
			return
		}
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "Error al buscar la tarea")
		return
	}

	var datos NuevosDatos
	if err := json.NewDecoder(r.Body).Decode(&datos); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "json_invalido", "El cuerpo no es un JSON válido")
		return
	}

	// Validación de estado si se proporciona uno nuevo (Fase 2b)
	if datos.Estado != "" && !estadosValidos[datos.Estado] {
		respuesta.Error(w, http.StatusUnprocessableEntity, "estado_invalido", "Estado no válido")
		return
	}

	if datos.Titulo != "" {
		t.Titulo = datos.Titulo
	}
	if datos.Estado != "" {
		t.Estado = datos.Estado
	}

	if err := m.DB.Debug().Save(&t).Error; err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "No se pudo actualizar la tarea")
		return
	}

	respuesta.Exito(w, http.StatusOK, t)
}

// borrar elimina una tarea por su ID (Fase 2a)
func (m *Manejador) borrar(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		respuesta.Error(w, http.StatusBadRequest, "id_invalido", "El ID debe ser un número entero positivo")
		return
	}

	var t Tarea
	if err := m.DB.Debug().First(&t, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			respuesta.Error(w, http.StatusNotFound, "no_encontrado", "La tarea no existe")
			return
		}
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "Error al buscar la tarea")
		return
	}

	if err := m.DB.Debug().Delete(&t).Error; err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "No se pudo eliminar la tarea")
		return
	}

	respuesta.Exito(w, http.StatusOK, map[string]string{"mensaje": "Tarea eliminada correctamente"})
}

// listarUsuarios resuelve el requerimiento 2c: obtiene los usuarios con Preload para evitar el problema N+1
func (m *Manejador) listarUsuarios(w http.ResponseWriter, r *http.Request) {
	var usuarios []Usuario

	if err := m.DB.Debug().Preload("Tareas").Find(&usuarios).Error; err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "No se pudieron obtener los usuarios")
		return
	}

	respuesta.Exito(w, http.StatusOK, usuarios)
}

// Struct auxiliar para los campos modificables en el PUT
type NuevosDatos struct {
	Titulo string `json:"titulo"`
	Estado string `json:"estado"`
}
