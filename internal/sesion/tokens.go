package sesion

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const emisor = "webA"

// Roles define los permisos de las cuentas de la aplicación.
const (
	RolPropietario = "propietario"
	RolColaborador = "colaborador"
)

// Identidad contiene la identidad autenticada por un token válido.
type Identidad struct {
	Usuario   string
	Rol       string
	UsuarioID uint
}

type claims struct {
	Rol       string `json:"rol"`
	UsuarioID uint   `json:"usuario_id"`
	jwt.RegisteredClaims
}

// Tokens firma y valida tokens de acceso.
type Tokens struct {
	secreto []byte
}

func NuevoTokens(secreto string) (*Tokens, error) {
	if len(secreto) < 32 {
		return nil, errors.New("JWT_SECRET debe tener al menos 32 caracteres")
	}
	return &Tokens{secreto: []byte(secreto)}, nil
}

func (t *Tokens) Crear(identidad Identidad) (string, error) {
	if identidad.Usuario == "" || identidad.Rol == "" || identidad.UsuarioID == 0 {
		return "", errors.New("la identidad del token está incompleta")
	}

	ahora := time.Now()
	payload := claims{
		Rol:       identidad.Rol,
		UsuarioID: identidad.UsuarioID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    emisor,
			Subject:   identidad.Usuario,
			IssuedAt:  jwt.NewNumericDate(ahora),
			ExpiresAt: jwt.NewNumericDate(ahora.Add(8 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return token.SignedString(t.secreto)
}

func (t *Tokens) Verificar(cadena string) (Identidad, error) {
	payload := &claims{}
	token, err := jwt.ParseWithClaims(cadena, payload, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("método de firma no permitido: %s", token.Method.Alg())
		}
		return t.secreto, nil
	}, jwt.WithIssuer(emisor))
	if err != nil {
		return Identidad{}, err
	}
	if !token.Valid || payload.Subject == "" || payload.Rol == "" || payload.UsuarioID == 0 {
		return Identidad{}, errors.New("token inválido")
	}
	return Identidad{
		Usuario:   payload.Subject,
		Rol:       payload.Rol,
		UsuarioID: payload.UsuarioID,
	}, nil
}
