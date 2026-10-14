package tareas

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"webA/internal/sesion"
)

const secretoPrueba = "una-clave-de-prueba-de-al-menos-32-caracteres"

// enrutador arma el servidor de prueba SIN base de datos (DB: nil).
// Sirve porque todas las pruebas de este archivo responden ANTES
// de la línea "desde aquí hace falta la base de datos".
func enrutador(t *testing.T) http.Handler {
	t.Helper()
	tokens, err := sesion.NuevoTokens(secretoPrueba)
	if err != nil {
		t.Fatal(err)
	}
	r := chi.NewRouter()
	(&Manejador{DB: nil}).Rutas(r, tokens)
	return r
}

func peticionAutenticada(t *testing.T, metodo, ruta, cuerpo string) *http.Request {
	t.Helper()
	return peticionConRol(t, metodo, ruta, cuerpo, sesion.RolPropietario, 1)
}

func peticionConRol(t *testing.T, metodo, ruta, cuerpo, rol string, usuarioID uint) *http.Request {
	t.Helper()
	tokens, err := sesion.NuevoTokens(secretoPrueba)
	if err != nil {
		t.Fatal(err)
	}
	token, err := tokens.Crear(sesion.Identidad{
		Usuario: "prueba", Rol: rol, UsuarioID: usuarioID,
	})
	if err != nil {
		t.Fatal(err)
	}
	peticion := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	peticion.Header.Set("Authorization", "Bearer "+token)
	return peticion
}

// TestCrear valida solo los casos que se comprueban antes de tocar la BD.
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
			peticion := peticionAutenticada(t, http.MethodPost, caso.ruta, caso.cuerpo)
			grabadora := httptest.NewRecorder()

			enrutador(t).ServeHTTP(grabadora, peticion)

			if grabadora.Code != caso.codigo {
				t.Fatalf("se esperaba %d y llegó %d con cuerpo %s", caso.codigo, grabadora.Code, grabadora.Body.String())
			}
			if caso.nombre == "estado inventado responde 422" &&
				!strings.Contains(grabadora.Body.String(), "estado no permitido: urgente") {
				t.Fatalf("la respuesta no explica el estado rechazado: %s", grabadora.Body.String())
			}
		})
	}
}

func TestVerUnoConIDInvalidoResponde400(t *testing.T) {
	casos := []struct {
		ruta string
	}{
		{ruta: "/tareas/abc"},
		{ruta: "/tareas/0"},
		{ruta: "/tareas/-3"},
	}

	for _, caso := range casos {
		t.Run(caso.ruta, func(t *testing.T) {
			peticion := peticionAutenticada(t, http.MethodGet, caso.ruta, "")
			grabadora := httptest.NewRecorder()
			enrutador(t).ServeHTTP(grabadora, peticion)

			if grabadora.Code != http.StatusBadRequest {
				t.Fatalf("se esperaba 400 y llegó %d", grabadora.Code)
			}
		})
	}
}

func TestListarConEstadoInvalidoResponde422(t *testing.T) {
	peticion := peticionAutenticada(t, http.MethodGet, "/tareas?estado=urgente", "")
	grabadora := httptest.NewRecorder()
	enrutador(t).ServeHTTP(grabadora, peticion)

	if grabadora.Code != http.StatusUnprocessableEntity {
		t.Fatalf("se esperaba 422 y llegó %d", grabadora.Code)
	}
}

func TestSinTokenResponde401(t *testing.T) {
	peticion := httptest.NewRequest(http.MethodGet, "/tareas", nil)
	grabadora := httptest.NewRecorder()
	enrutador(t).ServeHTTP(grabadora, peticion)

	if grabadora.Code != http.StatusUnauthorized {
		t.Fatalf("se esperaba 401 y llegó %d", grabadora.Code)
	}
}

func TestColaboradorNoPuedeListarUsuarios(t *testing.T) {
	peticion := peticionConRol(t, http.MethodGet, "/usuarios", "", sesion.RolColaborador, 2)
	grabadora := httptest.NewRecorder()
	enrutador(t).ServeHTTP(grabadora, peticion)

	if grabadora.Code != http.StatusForbidden {
		t.Fatalf("se esperaba 403 y llegó %d", grabadora.Code)
	}
}

func TestColaboradorSoloPuedeModificarSusTareas(t *testing.T) {
	tokens, err := sesion.NuevoTokens(secretoPrueba)
	if err != nil {
		t.Fatal(err)
	}
	manejador := tokens.Guardia()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if puedeModificar(Tarea{UsuarioID: 2}, r) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusForbidden)
	}))

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
			peticion := peticionConRol(t, http.MethodPatch, "/", "", sesion.RolColaborador, caso.usuarioID)
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
			if got := TransicionPermitida(caso.actual, caso.destino); got != caso.permitida {
				t.Fatalf("TransicionPermitida(%q, %q) = %v, se esperaba %v",
					caso.actual, caso.destino, got, caso.permitida)
			}
		})
	}
}
