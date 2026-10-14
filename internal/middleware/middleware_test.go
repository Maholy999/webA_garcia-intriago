package middleware

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecuperacionDevuelve500YRegistroLoAnota(t *testing.T) {
	var salida bytes.Buffer
	original := log.Writer()
	log.SetOutput(&salida)
	defer log.SetOutput(original)

	final := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("fallo de prueba")
	})
	manejador := Registro(Recuperacion(final))
	grabadora := httptest.NewRecorder()
	manejador.ServeHTTP(grabadora, httptest.NewRequest(http.MethodGet, "/fallo", nil))

	if grabadora.Code != http.StatusInternalServerError {
		t.Fatalf("se esperaba 500 y llegó %d", grabadora.Code)
	}
	if !strings.Contains(salida.String(), "GET /fallo → 500") {
		t.Fatalf("el registro no anotó el 500: %q", salida.String())
	}

	var cuerpo struct {
		OK    bool `json:"ok"`
		Error struct {
			Codigo string `json:"codigo"`
		} `json:"error"`
	}
	if err := json.Unmarshal(grabadora.Body.Bytes(), &cuerpo); err != nil {
		t.Fatalf("la respuesta no es JSON: %v", err)
	}
	if cuerpo.OK || cuerpo.Error.Codigo != "error_interno" {
		t.Fatalf("la respuesta de recuperación es incorrecta: %+v", cuerpo)
	}
}
