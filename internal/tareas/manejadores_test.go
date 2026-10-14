package tareas

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"webA/internal/sesion"
)

func enrutador(t *testing.T) http.Handler {
	t.Helper()
	return enrutadorCon(rolDePrueba)
}

func peticionAutenticada(metodo, ruta, cuerpo string) *http.Request {
	return httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
}

func TestCrear(t *testing.T) {
	casos := []struct {
		nombre string
		ruta   string
		cuerpo string
		codigo int
	}{
		{
			nombre: "JSON roto responde 400",
			ruta:   "/tareas",
			cuerpo: `{"titulo": "Revisar informe"`,
			codigo: http.StatusBadRequest,
		},
		{
			nombre: "estado inventado responde 422",
			ruta:   "/tareas",
			cuerpo: `{"titulo": "Revisar informe", "estado": "urgente"}`,
			codigo: http.StatusUnprocessableEntity,
		},
		{
			nombre: "estado vacio con JSON valido responde 422",
			ruta:   "/tareas",
			cuerpo: `{"titulo": "Revisar informe", "estado": ""}`,
			codigo: http.StatusUnprocessableEntity,
		},
		{
			nombre: "cuerpo vacio responde 400",
			ruta:   "/tareas",
			cuerpo: "",
			codigo: http.StatusBadRequest,
		},
		{
			nombre: "JSON con estado no permitido responde 422",
			ruta:   "/tareas",
			cuerpo: `{"titulo": "Revisar informe", "estado": "cancelada"}`,
			codigo: http.StatusUnprocessableEntity,
		},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			peticion := peticionAutenticada(http.MethodPost, caso.ruta, caso.cuerpo)
			grabadora := httptest.NewRecorder()
			enrutador(t).ServeHTTP(grabadora, peticion)

			if grabadora.Code != caso.codigo {
				t.Fatalf("se esperaba %d y llegó %d con cuerpo %s",
					caso.codigo, grabadora.Code, grabadora.Body.String())
			}
			if caso.nombre == "estado inventado responde 422" &&
				!strings.Contains(grabadora.Body.String(), "estado no permitido: urgente") {
				t.Fatalf("la respuesta no explica el estado rechazado: %s", grabadora.Body.String())
			}
		})
	}
}

func TestVerUnoConIDInvalidoResponde400(t *testing.T) {
	for _, ruta := range []string{"/tareas/abc", "/tareas/0", "/tareas/-3"} {
		t.Run(ruta, func(t *testing.T) {
			peticion := peticionAutenticada(http.MethodGet, ruta, "")
			grabadora := httptest.NewRecorder()
			enrutador(t).ServeHTTP(grabadora, peticion)

			if grabadora.Code != http.StatusBadRequest {
				t.Fatalf("se esperaba 400 y llegó %d", grabadora.Code)
			}
		})
	}
}

func TestListarConEstadoInvalidoResponde422(t *testing.T) {
	peticion := peticionAutenticada(http.MethodGet, "/tareas?estado=urgente", "")
	grabadora := httptest.NewRecorder()
	enrutador(t).ServeHTTP(grabadora, peticion)

	if grabadora.Code != http.StatusUnprocessableEntity {
		t.Fatalf("se esperaba 422 y llegó %d", grabadora.Code)
	}
}

func TestCambioEstadoInventadoResponde422(t *testing.T) {
	peticion := peticionAutenticada(http.MethodPatch, "/tareas/1", `{"estado":"inventado"}`)
	grabadora := httptest.NewRecorder()
	enrutador(t).ServeHTTP(grabadora, peticion)

	if grabadora.Code != http.StatusUnprocessableEntity {
		t.Fatalf("se esperaba 422 y llegó %d", grabadora.Code)
	}
	if !strings.Contains(grabadora.Body.String(), "estado no permitido: inventado") {
		t.Fatalf("la respuesta no explica el estado rechazado: %s", grabadora.Body.String())
	}
}

func TestColaboradorNoPuedeListarUsuarios(t *testing.T) {
	peticion := httptest.NewRequest(http.MethodGet, "/usuarios", nil)
	grabadora := httptest.NewRecorder()
	enrutadorCon(sesion.RolColaborador).ServeHTTP(grabadora, peticion)

	if grabadora.Code != http.StatusForbidden {
		t.Fatalf("se esperaba 403 y llegó %d", grabadora.Code)
	}
}

func TestColaboradorSoloPuedeModificarSusTareas(t *testing.T) {
	casos := []struct {
		nombre    string
		usuarioID uint
		esperado  int
	}{
		{nombre: "tarea propia", usuarioID: 2, esperado: http.StatusNoContent},
		{nombre: "tarea de otro usuario", usuarioID: 3, esperado: http.StatusForbidden},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			siguiente := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if puedeModificar(Tarea{UsuarioID: 2}, r) {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				w.WriteHeader(http.StatusForbidden)
			})
			manejador := tokensDePrueba.Autenticar(siguiente)
			token, err := tokensDePrueba.Emitir(sesion.Usuario{
				ID: caso.usuarioID, Nombre: "prueba", Rol: sesion.RolColaborador,
			})
			if err != nil {
				t.Fatal(err)
			}

			peticion := httptest.NewRequest(http.MethodPatch, "/", nil)
			peticion.Header.Set("Authorization", "Bearer "+token)
			grabadora := httptest.NewRecorder()
			manejador.ServeHTTP(grabadora, peticion)

			if grabadora.Code != caso.esperado {
				t.Fatalf("se esperaba %d y llegó %d", caso.esperado, grabadora.Code)
			}
		})
	}
}

func TestTransicionesDeTarea(t *testing.T) {
	casos := []struct {
		actual    string
		destino   string
		permitida bool
	}{
		{actual: "pendiente", destino: "en_proceso", permitida: true},
		{actual: "en_proceso", destino: "completada", permitida: true},
		{actual: "pendiente", destino: "completada", permitida: false},
		{actual: "completada", destino: "pendiente", permitida: false},
		{actual: "completada", destino: "completada", permitida: false},
	}
	for _, caso := range casos {
		t.Run(caso.actual+" a "+caso.destino, func(t *testing.T) {
			if actual := puedePasar(caso.actual, caso.destino); actual != caso.permitida {
				t.Fatalf("puedePasar(%q, %q) = %v, se esperaba %v",
					caso.actual, caso.destino, actual, caso.permitida)
			}
		})
	}
}
