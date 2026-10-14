package tareas

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Prueba de integración: arma el servidor entero (rutas, guardia y
// manejadores) y le envía cada petición SIN token.
//
// Es la prueba que cuida la puerta: ninguna ruta protegida responde sin
// token. Si mañana alguien saca una ruta del grupo protegido, esta
// prueba lo dice antes de que el cambio llegue a los usuarios.
func TestSinTokenLasRutasProtegidasResponden401(t *testing.T) {
	// CAMBIEN la lista: una fila por cada ruta de su matriz que NO sea
	// pública. Método, ruta con un id cualquiera, y cuerpo si lo lleva.
	protegidas := []struct{ metodo, ruta, cuerpo string }{
		{"GET", "/tareas", ""},
		{"POST", "/tareas", `{"titulo":"Sin autorización"}`},
		{"GET", "/tareas/1", ""},
		{"PUT", "/tareas/1", `{"titulo":"Cambio no autorizado"}`},
		{"PATCH", "/tareas/1", `{"estado":"en_proceso"}`},
		{"DELETE", "/tareas/1", ""},
		{"GET", "/usuarios", ""},
	}

	servidor := enrutadorCon("") // "" = nadie inició sesión

	for _, p := range protegidas {
		peticion := httptest.NewRequest(p.metodo, p.ruta,
			strings.NewReader(p.cuerpo))
		grabadora := httptest.NewRecorder()

		cayo := seCayo(t, func() { servidor.ServeHTTP(grabadora, peticion) })

		if cayo {
			t.Errorf("%s %s sin token: el guardia no la detuvo y la "+
				"petición llegó hasta la base de datos", p.metodo, p.ruta)
			continue
		}
		if grabadora.Code != http.StatusUnauthorized {
			t.Errorf("%s %s sin token: se esperaba 401 y llegó %d",
				p.metodo, p.ruta, grabadora.Code)
		}
	}
}
