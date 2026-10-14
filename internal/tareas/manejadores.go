package tareas

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"webA/internal/respuesta"
	"webA/internal/sesion"
)

// Manejador guarda la conexión a la base de datos.
type Manejador struct{ DB *gorm.DB }

// Rutas registra todas las rutas HTTP del paquete
func (m *Manejador) Rutas(r chi.Router, tokens *sesion.Tokens) {
	// Nivel 1 — pública: iniciar sesión.
	r.Post("/login", (&sesion.Manejador{DB: m.DB, Tokens: tokens}).Login)

	// Nivel 2 — con sesión: cualquier usuario autenticado puede acceder.
	r.Group(func(conSesion chi.Router) {
		conSesion.Use(tokens.Autenticar)
		conSesion.Post("/tareas", m.crear)
		conSesion.Get("/tareas", m.listar)
		conSesion.Get("/tareas/{id}", m.verUno)
		conSesion.Put("/tareas/{id}", m.actualizar)
		conSesion.Patch("/tareas/{id}", m.cambiarEstado)
		conSesion.Delete("/tareas/{id}", m.borrar)
	})

	// Nivel 3 — solo propietario: la administración de usuarios no es pública.
	r.Group(func(conRol chi.Router) {
		conRol.Use(tokens.Autenticar)
		conRol.Use(sesion.SoloRol(sesion.RolPropietario))
		conRol.Get("/usuarios", m.listarUsuarios)
	})
}

// crear procesa las peticiones POST para guardar una nueva Tarea
func (m *Manejador) crear(w http.ResponseWriter, r *http.Request) {
	var entrada NuevaTarea
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "json_invalido", "El cuerpo no es un JSON válido")
		return
	}

	if entrada.Estado != nil && !estadosValidos[*entrada.Estado] {
		respuesta.Error(w, http.StatusUnprocessableEntity, "estado_no_permitido",
			fmt.Sprintf("estado no permitido: %s", *entrada.Estado))
		return
	}
	if entrada.Estado != nil && *entrada.Estado != "pendiente" {
		respuesta.Error(w, http.StatusUnprocessableEntity, "estado_inicial",
			"Una tarea nueva debe iniciar en estado pendiente")
		return
	}
	if entrada.Titulo == "" {
		respuesta.Error(w, http.StatusUnprocessableEntity, "titulo_requerido", "El título es obligatorio")
		return
	}

	identidad, ok := sesion.De(r)
	if !ok {
		respuesta.Error(w, http.StatusUnauthorized, "falta_token", "Falta una identidad autenticada")
		return
	}
	usuarioID := identidad.UsuarioID
	if identidad.Rol == sesion.RolPropietario && entrada.UsuarioID != 0 {
		usuarioID = entrada.UsuarioID
	} else if identidad.Rol != sesion.RolPropietario && entrada.UsuarioID != 0 &&
		entrada.UsuarioID != identidad.UsuarioID {
		respuesta.Error(w, http.StatusForbidden, "sin_permiso",
			"Un colaborador solo puede crear tareas para sí mismo")
		return
	}

	var u Usuario
	if err := m.DB.First(&u, usuarioID).Error; err != nil {
		respuesta.Error(w, http.StatusUnprocessableEntity, "usuario_inexistente", "El usuario especificado no existe")
		return
	}

	t := Tarea{
		Titulo:    entrada.Titulo,
		Estado:    "pendiente",
		UsuarioID: usuarioID,
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

	// 1. Obtener parámetro "estado" de la URL (ej. /tareas?estado=completada)
	estadoFiltro := r.URL.Query().Get("estado")

	// 2. Si se envió parámetro, se valida el mapa antes de consultar la base
	if estadoFiltro != "" {
		if !estadosValidos[estadoFiltro] {
			respuesta.Error(w, http.StatusUnprocessableEntity, "estado_no_permitido",
				"Estado no válido para filtrado")
			return
		}
	}

	query := m.DB.Debug()

	// 3. Aplicar el filtro parametrizado solo si el valor es válido
	if estadoFiltro != "" {
		// Previene inyección SQL mediante uso de placeholder '?'
		query = query.Where("estado = ?", estadoFiltro)
	}

	// 4. Ejecutar la consulta final
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
	if !puedeModificar(t, r) {
		respuesta.Error(w, http.StatusForbidden, "sin_permiso",
			"Un colaborador solo puede modificar sus propias tareas")
		return
	}

	var datos NuevosDatos
	if err := json.NewDecoder(r.Body).Decode(&datos); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "json_invalido", "El cuerpo no es un JSON válido")
		return
	}

	if datos.Titulo != "" {
		t.Titulo = datos.Titulo
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
	if !puedeModificar(t, r) {
		respuesta.Error(w, http.StatusForbidden, "sin_permiso",
			"Un colaborador solo puede eliminar sus propias tareas")
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
}

func idDeRuta(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		respuesta.Error(w, http.StatusBadRequest, "id_invalido",
			"El ID debe ser un número entero positivo")
		return 0, false
	}
	return id, true
}

func puedeModificar(t Tarea, r *http.Request) bool {
	identidad, ok := sesion.De(r)
	if !ok {
		return false
	}
	return identidad.Rol == sesion.RolPropietario || t.UsuarioID == identidad.UsuarioID
}
