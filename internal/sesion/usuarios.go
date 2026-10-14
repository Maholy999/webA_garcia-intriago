package sesion

// En este archivo cambian SOLO las líneas marcadas con «CAMBIEN»:
// los dos roles de su negocio y, si quieren, los dos usuarios de ejemplo.

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Los dos roles de su negocio, escritos como en la sección 6 de su ficha.
// El nombre de la constante empieza con Rol; el texto va en minúsculas,
// sin espacios y sin tildes.
const (
	RolPropietario = "propietario"
	RolColaborador = "colaborador"
)

// Usuario es una persona que puede iniciar sesión.
// De la contraseña se guarda un resumen (bcrypt), nunca la contraseña.
type Usuario struct {
	ID        uint   `json:"id"`
	Nombre    string `json:"nombre" gorm:"uniqueIndex"`
	Rol       string `json:"rol"`
	ClaveHash string `json:"-"`
}

// El dominio ya usa la tabla usuarios; las cuentas de acceso se guardan aparte.
func (Usuario) TableName() string {
	return "cuentas"
}

// Sembrar deja dos usuarios de ejemplo, uno por rol. Corre en cada
// arranque: si cambian un rol o una contraseña aquí, al reiniciar el
// servidor el usuario queda actualizado. Son usuarios y contraseñas
// INVENTADOS, solo para probar: nunca una contraseña real.
func Sembrar(db *gorm.DB) error {
	ejemplos := []struct{ nombre, rol, clave, usuarioDominio string }{
		{"ana", RolPropietario, "ana-2026", "Juan Pérez"},
		{"luis", RolColaborador, "luis-2026", "María López"},
	}
	tieneUsuarioAntiguo := db.Migrator().HasColumn(&Usuario{}, "usuario")
	tieneUsuarioIDAntiguo := db.Migrator().HasColumn(&Usuario{}, "usuario_id")

	for _, e := range ejemplos {
		// bcrypt calcula el resumen: eso es lo que se guarda.
		resumen, err := bcrypt.GenerateFromPassword(
			[]byte(e.clave), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		var perfil struct {
			ID uint
		}
		if err := db.Table("usuarios").Select("id").
			Where("nombre = ?", e.usuarioDominio).Take(&perfil).Error; err != nil {
			return fmt.Errorf("no se encontró el usuario del dominio %q: %w", e.usuarioDominio, err)
		}

		var u Usuario
		busqueda := db.Where("nombre = ?", e.nombre)
		if tieneUsuarioAntiguo {
			busqueda = db.Where("nombre = ? OR usuario = ?", e.nombre, e.nombre)
		}
		err = busqueda.First(&u).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		campos := map[string]any{
			"id":         perfil.ID,
			"nombre":     e.nombre,
			"rol":        e.rol,
			"clave_hash": string(resumen),
		}
		if tieneUsuarioAntiguo {
			campos["usuario"] = e.nombre
		}
		if tieneUsuarioIDAntiguo {
			campos["usuario_id"] = perfil.ID
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.Table("cuentas").Create(campos).Error; err != nil {
				return err
			}
			continue
		}

		if err := db.Table("cuentas").Where("id = ?", u.ID).Updates(campos).Error; err != nil {
			return err
		}
	}
	return nil
}

// Manejador tiene la única ruta de este paquete: iniciar sesión.
type Manejador struct {
	DB     *gorm.DB
	Tokens *Tokens
}

// Login: POST /login con {"usuario": "...", "clave": "..."}.
// 400 si el JSON está roto · 401 si no coincide · 200 con el token.
func (m *Manejador) Login(w http.ResponseWriter, r *http.Request) {
	var pedido struct {
		Usuario string `json:"usuario"`
		Clave   string `json:"clave"`
	}
	if err := json.NewDecoder(r.Body).Decode(&pedido); err != nil {
		responder(w, http.StatusBadRequest, "el cuerpo no es JSON válido")
		return
	}
	// El mismo mensaje para «no existe» y para «contraseña incorrecta»:
	// quien prueba nombres al azar no averigua cuáles existen.
	const noCoincide = "usuario o contraseña incorrectos"
	// ---- desde aquí hace falta la base de datos ----
	var u Usuario
	err := m.DB.Where("nombre = ?", pedido.Usuario).Limit(1).Find(&u).Error
	if err != nil {
		responder(w, http.StatusInternalServerError, "no se pudo leer")
		return
	}
	// bcrypt compara la contraseña que llegó con el resumen guardado.
	// Si el usuario no existe, u.ID es 0 y no hay nada que comparar.
	if u.ID == 0 || bcrypt.CompareHashAndPassword(
		[]byte(u.ClaveHash), []byte(pedido.Clave)) != nil {
		responder(w, http.StatusUnauthorized, noCoincide)
		return
	}
	token, err := m.Tokens.Emitir(u)
	if err != nil {
		responder(w, http.StatusInternalServerError,
			"no se pudo emitir el token")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
