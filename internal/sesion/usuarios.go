package sesion

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Cuenta almacena credenciales separadas de los usuarios del dominio.
type Cuenta struct {
	ID        uint   `gorm:"primaryKey"`
	Usuario   string `gorm:"uniqueIndex;not null"`
	ClaveHash string `gorm:"not null"`
	Rol       string `gorm:"not null"`
	UsuarioID uint   `gorm:"not null;index"`
}

func (Cuenta) TableName() string { return "cuentas" }

type perfil struct {
	ID     uint   `gorm:"primaryKey"`
	Nombre string `gorm:"not null"`
}

func (perfil) TableName() string { return "usuarios" }

type cuentaInicial struct {
	usuario string
	clave   string
	rol     string
	nombre  string
}

// Sembrar crea las cuentas de demostración y las relaciona con usuarios existentes.
func Sembrar(db *gorm.DB) error {
	iniciales := []cuentaInicial{
		{usuario: "ana", clave: "ana-2026", rol: RolPropietario, nombre: "Juan Pérez"},
		{usuario: "luis", clave: "luis-2026", rol: RolColaborador, nombre: "María López"},
	}

	for _, inicial := range iniciales {
		var usuario perfil
		if err := db.Where("nombre = ?", inicial.nombre).FirstOrCreate(
			&usuario, perfil{Nombre: inicial.nombre},
		).Error; err != nil {
			return err
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(inicial.clave), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		cuenta := Cuenta{
			Usuario:   inicial.usuario,
			ClaveHash: string(hash),
			Rol:       inicial.rol,
			UsuarioID: usuario.ID,
		}
		if err := db.Where("usuario = ?", inicial.usuario).FirstOrCreate(&Cuenta{}, cuenta).Error; err != nil {
			return err
		}
	}
	return nil
}

// Autenticar valida una contraseña sin revelar si existe la cuenta.
func Autenticar(db *gorm.DB, usuario, clave string) (Cuenta, bool, error) {
	var cuenta Cuenta
	if err := db.Where("usuario = ?", usuario).First(&cuenta).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Cuenta{}, false, nil
		}
		return Cuenta{}, false, err
	}
	if bcrypt.CompareHashAndPassword([]byte(cuenta.ClaveHash), []byte(clave)) != nil {
		return Cuenta{}, false, nil
	}
	return cuenta, true, nil
}
