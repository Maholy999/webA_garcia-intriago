// Package sesion responde a una sola pregunta: ¿quién hace esta petición?
// Emite los tokens al iniciar sesión y los revisa en cada petición.
//
// Este archivo se copia tal cual: no tiene ninguna línea que cambiar.
package sesion

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Duracion es cuánto vale un token desde que se emite.
const Duracion = time.Hour

// Sesion es lo que el servidor sabe de quien hace la petición.
// Sale del token: no se consulta la base de datos en cada petición.
type Sesion struct {
	UsuarioID uint
	Nombre    string
	Rol       string
}

// datos es lo que viaja dentro del token. Cualquiera puede leerlo:
// va codificado, no cifrado. Lo que nadie puede hacer es cambiarlo.
type datos struct {
	Nombre string `json:"nombre"`
	Rol    string `json:"rol"`
	jwt.RegisteredClaims
}

// Tokens emite y revisa tokens con el secreto que recibe.
type Tokens struct {
	secreto []byte
}

// NuevosTokens recibe el secreto de la firma. El secreto llega desde la
// configuración (la variable de entorno); nunca se escribe en el código.
func NuevosTokens(secreto string) *Tokens {
	return &Tokens{secreto: []byte(secreto)}
}

// Emitir firma un token para el usuario que acaba de probar quién es.
func (t *Tokens) Emitir(u Usuario) (string, error) {
	ahora := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, datos{
		Nombre: u.Nombre,
		Rol:    u.Rol,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(int(u.ID)),
			IssuedAt:  jwt.NewNumericDate(ahora),
			ExpiresAt: jwt.NewNumericDate(ahora.Add(Duracion)),
		},
	})
	return token.SignedString(t.secreto)
}

// ErrVencido distingue un token que fue bueno de uno que nunca lo fue.
var ErrVencido = errors.New("el token está vencido")

// Leer revisa la firma y la fecha de vencimiento. Solo si las dos
// pasan devuelve la sesión.
func (t *Tokens) Leer(texto string) (Sesion, error) {
	var d datos
	_, err := jwt.ParseWithClaims(texto, &d,
		func(*jwt.Token) (any, error) { return t.secreto, nil },
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
	)
	if errors.Is(err, jwt.ErrTokenExpired) {
		return Sesion{}, ErrVencido
	}
	if err != nil {
		return Sesion{}, err
	}
	id, err := strconv.Atoi(d.Subject)
	if err != nil {
		return Sesion{}, err
	}
	return Sesion{UsuarioID: uint(id), Nombre: d.Nombre, Rol: d.Rol}, nil
}
