package tareas

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"

	"gorm.io/gorm"

	"webA/internal/respuesta"
)

// siguientes es la máquina de estados: desde cada estado, a cuáles se puede pasar.
var siguientes = map[string][]string{
	"pendiente":  {"en_proceso"},
	"en_proceso": {"completada"},
	"completada": {},
}

// puedePasar dice si la máquina de estados permite ir de un estado a otro.
func puedePasar(de, a string) bool {
	return slices.Contains(siguientes[de], a)
}

// cambiarEstado procesa PATCH /tareas/{id} con un cuerpo {"estado":"..."}.
func (m *Manejador) cambiarEstado(w http.ResponseWriter, r *http.Request) {
	id, ok := idDeRuta(w, r)
	if !ok {
		return
	}

	var pedido struct {
		Estado string `json:"estado"`
	}
	if err := json.NewDecoder(r.Body).Decode(&pedido); err != nil {
		respuesta.Error(w, http.StatusBadRequest, "json_invalido", "El cuerpo no es JSON válido")
		return
	}
	if _, existe := siguientes[pedido.Estado]; !existe {
		respuesta.Error(w, http.StatusUnprocessableEntity, "estado_no_permitido",
			"estado no permitido: "+pedido.Estado)
		return
	}

	var tarea Tarea
	if err := m.DB.First(&tarea, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respuesta.Error(w, http.StatusNotFound, "no_encontrado", "La tarea no existe")
			return
		}
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "No se pudo leer la tarea")
		return
	}
	if !puedeModificar(tarea, r) {
		respuesta.Error(w, http.StatusForbidden, "sin_permiso",
			"El colaborador solo puede cambiar tareas propias")
		return
	}
	if !puedePasar(tarea.Estado, pedido.Estado) {
		respuesta.Error(w, http.StatusUnprocessableEntity, "transicion_prohibida",
			"de "+tarea.Estado+" no se puede pasar a "+pedido.Estado)
		return
	}

	tarea.Estado = pedido.Estado
	if err := m.DB.Save(&tarea).Error; err != nil {
		respuesta.Error(w, http.StatusInternalServerError, "error_base", "No se pudo guardar la tarea")
		return
	}
	respuesta.Exito(w, http.StatusOK, tarea)
}
