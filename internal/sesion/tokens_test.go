package sesion

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const secretoTokensPrueba = "secreto-solo-de-las-pruebas"

func TestEmitirYLeerToken(t *testing.T) {
	tokens := NuevosTokens(secretoTokensPrueba)
	usuario := Usuario{ID: 17, Nombre: "ana", Rol: RolPropietario}

	cadena, err := tokens.Emitir(usuario)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := tokens.Leer(cadena)
	if err != nil {
		t.Fatal(err)
	}
	if actual.UsuarioID != usuario.ID || actual.Nombre != usuario.Nombre || actual.Rol != usuario.Rol {
		t.Fatalf("sesión: se esperaba %+v y llegó %+v", usuario, actual)
	}
}

func TestLeerRechazaTokenModificado(t *testing.T) {
	tokens := NuevosTokens(secretoTokensPrueba)
	cadena, err := tokens.Emitir(Usuario{ID: 1, Nombre: "ana", Rol: RolPropietario})
	if err != nil {
		t.Fatal(err)
	}
	partes := strings.Split(cadena, ".")
	if len(partes) != 3 {
		t.Fatalf("formato inesperado del token: %q", cadena)
	}
	partes[2] = "x" + partes[2][1:]
	modificado := strings.Join(partes, ".")
	if _, err := tokens.Leer(modificado); err == nil {
		t.Fatal("se esperaba que el token modificado fuera rechazado")
	}
}

func TestAutenticarRechazaFaltaToken(t *testing.T) {
	tokens := NuevosTokens(secretoTokensPrueba)
	manejador := tokens.Autenticar(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	grabadora := httptest.NewRecorder()
	manejador.ServeHTTP(grabadora, httptest.NewRequest(http.MethodGet, "/", nil))
	if grabadora.Code != http.StatusUnauthorized {
		t.Fatalf("se esperaba 401 y llegó %d", grabadora.Code)
	}
}

func TestSoloRolRechazaRolDistinto(t *testing.T) {
	tokens := NuevosTokens(secretoTokensPrueba)
	usuario := Usuario{ID: 2, Nombre: "luis", Rol: RolColaborador}
	cadena, err := tokens.Emitir(usuario)
	if err != nil {
		t.Fatal(err)
	}

	final := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	manejador := tokens.Autenticar(SoloRol(RolPropietario)(final))
	peticion := httptest.NewRequest(http.MethodGet, "/", nil)
	peticion.Header.Set("Authorization", "Bearer "+cadena)
	grabadora := httptest.NewRecorder()
	manejador.ServeHTTP(grabadora, peticion)
	if grabadora.Code != http.StatusForbidden {
		t.Fatalf("se esperaba 403 y llegó %d", grabadora.Code)
	}
}
