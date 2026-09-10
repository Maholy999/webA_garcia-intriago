// internal/middleware/middleware.go
package middleware

import (
	"log"
	"net/http"
	"time"

	"webA/internal/respuesta"
)

// grabador envuelve al ResponseWriter para saber qué código escribió el handler.
type grabador struct {
	http.ResponseWriter // embebido: Header, Write, etc. siguen igual
	estado              int
}

func (g *grabador) WriteHeader(codigo int) {
	g.estado = codigo                    // 1. anotar
	g.ResponseWriter.WriteHeader(codigo) // 2. pasar al original
}

// Registro escribe una línea por petición: método, ruta, estado y duración.
func Registro(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		// 200 por defecto: un handler que no llama a WriteHeader envía 200
		g := &grabador{ResponseWriter: w, estado: http.StatusOK}
		next.ServeHTTP(g, r) // ← se pasa g, NO w
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, g.estado, time.Since(inicio))
	})
}

// internal/middleware/middleware.go (continuación).
// Agregar "log" y "mesa-ayuda/internal/respuesta" al bloque import.
// Recuperacion atrapa un pánico y lo convierte en 500 con envoltura.
func Recuperacion(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				log.Printf("PÁNICO en %s %s: %v", r.Method, r.URL.Path, p)
				respuesta.Error(w, http.StatusInternalServerError, "error_interno", "ocurrió un error inesperado")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
