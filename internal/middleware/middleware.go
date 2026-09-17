package middleware

import (
	"log"
	"net/http"
	"time"

	"webA/internal/respuesta"
)

// grabador envuelve al ResponseWriter para saber qué código escribió el handler.
type grabador struct {
	http.ResponseWriter
	estado int
}

func (g *grabador) WriteHeader(codigo int) {
	g.estado = codigo
	g.ResponseWriter.WriteHeader(codigo)
}

// Registro escribe una línea por petición: método, ruta, estado y duración.
func Registro(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		g := &grabador{ResponseWriter: w, estado: http.StatusOK}
		next.ServeHTTP(g, r)
		// Después de next: el estado y la duración solo existen cuando el handler terminó.
		log.Printf("%s %s → %d (%s)", r.Method, r.URL.Path, g.estado, time.Since(inicio))
	})
}

// Recuperacion atrapa un pánico y lo convierte en 500 con envoltura.
func Recuperacion(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				log.Printf("PÁNICO en %s %s: %v", r.Method, r.URL.Path, p)
				respuesta.Error(w, http.StatusInternalServerError,
					"error_interno", "ocurrió un error inesperado")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
