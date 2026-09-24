package tareas

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"webA/internal/respuesta"
)

func TestErrorTieneLaFormaDelContrato(t *testing.T) {
	grabadora := httptest.NewRecorder()
	respuesta.Error(grabadora, 422, "estado no permitido", "Estado no válido")
	if grabadora.Code != 422 {
		t.Fatalf("código: se esperaba 422 y llegó %d", grabadora.Code)
	}
	var cuerpo map[string]any
	if err := json.Unmarshal(grabadora.Body.Bytes(), &cuerpo); err != nil {
		t.Fatalf("el cuerpo no es JSON: %v", err)
	}
	if cuerpo["ok"] != false {
		t.Fatalf("cuerpo: se esperaba ok=false y llegó %v", cuerpo)
	}
	if errorData, ok := cuerpo["error"].(map[string]any); ok {
		if errorData["codigo"] != "estado no permitido" {
			t.Fatalf("cuerpo: se esperaba error.codigo con el mensaje y llegó %v", cuerpo)
		}
	} else {
		t.Fatalf("cuerpo: no hay campo error con la estructura esperada: %v", cuerpo)
	}
	tipo := grabadora.Header().Get("Content-Type")
	if tipo != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type: llegó %q", tipo)
	}
}
