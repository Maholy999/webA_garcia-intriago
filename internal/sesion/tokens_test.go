package sesion

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const secretoTokensPrueba = "una-clave-de-prueba-de-al-menos-32-caracteres"

func TestTokensCrearYVerificar(t *testing.T) {
	tokens, err := NuevoTokens(secretoTokensPrueba)
	if err != nil {
		t.Fatal(err)
	}

	esperada := Identidad{Usuario: "ana", Rol: RolPropietario, UsuarioID: 1}
	cadena, err := tokens.Crear(esperada)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := tokens.Verificar(cadena)
	if err != nil {
		t.Fatal(err)
	}
	if actual != esperada {
		t.Fatalf("identidad: se esperaba %+v y llegó %+v", esperada, actual)
	}
}

func TestGuardiaRechazaFaltaToken(t *testing.T) {
	tokens, err := NuevoTokens(secretoTokensPrueba)
	if err != nil {
		t.Fatal(err)
	}

	manejador := tokens.Guardia()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	grabadora := httptest.NewRecorder()
	manejador.ServeHTTP(grabadora, httptest.NewRequest(http.MethodGet, "/", nil))
	if grabadora.Code != http.StatusUnauthorized {
		t.Fatalf("se esperaba 401 y llegó %d", grabadora.Code)
	}
}

func TestGuardiaRechazaRolDistinto(t *testing.T) {
	tokens, err := NuevoTokens(secretoTokensPrueba)
	if err != nil {
		t.Fatal(err)
	}
	cadena, err := tokens.Crear(Identidad{
		Usuario: "luis", Rol: RolColaborador, UsuarioID: 2,
	})
	if err != nil {
		t.Fatal(err)
	}

	manejador := tokens.Guardia(RolPropietario)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	peticion := httptest.NewRequest(http.MethodGet, "/", nil)
	peticion.Header.Set("Authorization", "Bearer "+cadena)
	grabadora := httptest.NewRecorder()
	manejador.ServeHTTP(grabadora, peticion)
	if grabadora.Code != http.StatusForbidden {
		t.Fatalf("se esperaba 403 y llegó %d", grabadora.Code)
	}
}

func TestNuevoTokensRequiereSecretoLargo(t *testing.T) {
	if _, err := NuevoTokens("corto"); err == nil {
		t.Fatal("se esperaba un error para un secreto corto")
	}
}
