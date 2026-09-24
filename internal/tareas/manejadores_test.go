package tareas

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// enrutador arma el servidor de prueba SIN base de datos (DB: nil).
// Sirve porque todas las pruebas de este archivo responden ANTES
// de la línea "desde aquí hace falta la base de datos".
func enrutador() http.Handler {
	r := chi.NewRouter()
	(&Manejador{DB: nil}).Rutas(r)
	return r
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
			peticion := httptest.NewRequest(http.MethodPost, caso.ruta, strings.NewReader(caso.cuerpo))
			grabadora := httptest.NewRecorder()

			enrutador().ServeHTTP(grabadora, peticion)

			if grabadora.Code != caso.codigo {
				t.Fatalf("se esperaba %d y llegó %d con cuerpo %s", caso.codigo, grabadora.Code, grabadora.Body.String())
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
			peticion := httptest.NewRequest(http.MethodGet, caso.ruta, nil)
			grabadora := httptest.NewRecorder()
			enrutador().ServeHTTP(grabadora, peticion)

			if grabadora.Code != http.StatusBadRequest {
				t.Fatalf("se esperaba 400 y llegó %d", grabadora.Code)
			}
		})
	}
}

func TestListarConEstadoInvalidoResponde422(t *testing.T) {
	peticion := httptest.NewRequest(http.MethodGet, "/tareas?estado=urgente", nil)
	grabadora := httptest.NewRecorder()
	enrutador().ServeHTTP(grabadora, peticion)

	if grabadora.Code != http.StatusUnprocessableEntity {
		t.Fatalf("se esperaba 422 y llegó %d", grabadora.Code)
	}
}
