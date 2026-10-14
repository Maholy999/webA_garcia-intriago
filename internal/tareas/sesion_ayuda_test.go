package tareas

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"

	"webA/internal/sesion"
)

// rolDePrueba es el rol con el que entran las pruebas de la semana 4:
// el que tiene permiso en MÁS rutas de su matriz.
const rolDePrueba = sesion.RolPropietario

// tokensDePrueba firma con un secreto que solo existe en las pruebas.
var tokensDePrueba = sesion.NuevosTokens("secreto-solo-de-las-pruebas")

// enrutadorCon arma el servidor de prueba, SIN base de datos (DB: nil),
// con las mismas rutas y los mismos niveles que usa main.go.
// Con un rol, todas las peticiones entran con el token de un usuario de
// ese rol. Con "" entran sin token.
func enrutadorCon(rol string) http.Handler {
	r := chi.NewRouter()
	(&Manejador{DB: nil}).Rutas(r, tokensDePrueba) // el mismo Rutas de main.go
	if rol == "" {
		return r
	}
	token, err := tokensDePrueba.Emitir(
		sesion.Usuario{ID: 1, Nombre: "prueba", Rol: rol})
	if err != nil {
		panic(err)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, p *http.Request) {
		p.Header.Set("Authorization", "Bearer "+token)
		r.ServeHTTP(w, p)
	})
}

// seCayo dice si el manejador llegó a usar la base, que en las pruebas
// no existe. Sirve para dar un mensaje claro en lugar de un pánico.
func seCayo(t *testing.T, f func()) (cayo bool) {
	t.Helper()
	defer func() {
		if recover() != nil {
			cayo = true
		}
	}()
	f()
	return false
}
