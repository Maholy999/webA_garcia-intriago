package sesion

// Este archivo se copia tal cual: no tiene ninguna línea que cambiar.

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// clave es el tipo de la llave con la que la sesión se guarda en el
// contexto de la petición. Es propio de este paquete: nadie más puede
// escribir ahí.
type clave struct{}

// De devuelve la sesión que el guardia dejó en la petición.
func De(r *http.Request) (Sesion, bool) {
	s, ok := r.Context().Value(clave{}).(Sesion)
	return s, ok
}

// Autenticar es el guardia: corre ANTES que el manejador.
// Sin token, con un token cambiado o con un token vencido responde 401,
// y el manejador que sigue nunca se entera.
func (t *Tokens) Autenticar(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cabecera := r.Header.Get("Authorization")
		texto, ok := strings.CutPrefix(cabecera, "Bearer ")
		if !ok {
			responder(w, http.StatusUnauthorized,
				"falta el token: inicie sesión en POST /login")
			return
		}
		s, err := t.Leer(texto)
		if errors.Is(err, ErrVencido) {
			responder(w, http.StatusUnauthorized,
				"el token está vencido: inicie sesión otra vez")
			return
		}
		if err != nil {
			responder(w, http.StatusUnauthorized, "el token no es válido")
			return
		}
		// Ya se sabe quién es: se anota en la petición y se sigue.
		ctx := context.WithValue(r.Context(), clave{}, s)
		siguiente.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SoloRol responde a la segunda pregunta: ya sé quién es, ¿puede hacer
// esto? Si su rol no es el pedido responde 403. Va DESPUÉS de Autenticar.
func SoloRol(rol string) func(http.Handler) http.Handler {
	return func(siguiente http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			s, ok := De(r)
			if !ok || s.Rol != rol {
				responder(w, http.StatusForbidden,
					"solo el rol "+rol+" puede hacer esto")
				return
			}
			siguiente.ServeHTTP(w, r)
		})
	}
}

// responder escribe {"error": mensaje} con el código indicado: la misma
// forma de error que usa todo el servidor (contrato del curso).
// Está aquí para que este paquete no dependa de ningún otro del proyecto.
func responder(w http.ResponseWriter, codigo int, mensaje string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codigo)
	json.NewEncoder(w).Encode(map[string]string{"error": mensaje})
}
