package sesion

import (
	"context"
	"net/http"
	"strings"

	"webA/internal/respuesta"
)

type claveIdentidad struct{}

// IdentidadDe obtiene la identidad que Guardia guardó en la petición.
func IdentidadDe(r *http.Request) (Identidad, bool) {
	identidad, ok := r.Context().Value(claveIdentidad{}).(Identidad)
	return identidad, ok
}

// Guardia exige un token válido y, si se indican, uno de los roles permitidos.
func (t *Tokens) Guardia(roles ...string) func(http.Handler) http.Handler {
	return func(siguiente http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			esquema, cadena, ok := strings.Cut(r.Header.Get("Authorization"), " ")
			if !ok || !strings.EqualFold(esquema, "Bearer") || strings.TrimSpace(cadena) == "" {
				respuesta.Error(w, http.StatusUnauthorized, "falta_token",
					"Falta el token: inicie sesión en POST /login")
				return
			}

			identidad, err := t.Verificar(strings.TrimSpace(cadena))
			if err != nil {
				respuesta.Error(w, http.StatusUnauthorized, "token_invalido", "El token no es válido")
				return
			}

			if len(roles) > 0 && !incluye(roles, identidad.Rol) {
				respuesta.Error(w, http.StatusForbidden, "sin_permiso",
					"No tiene permiso para esta operación")
				return
			}

			ctx := context.WithValue(r.Context(), claveIdentidad{}, identidad)
			siguiente.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func incluye(valores []string, buscado string) bool {
	for _, valor := range valores {
		if valor == buscado {
			return true
		}
	}
	return false
}
