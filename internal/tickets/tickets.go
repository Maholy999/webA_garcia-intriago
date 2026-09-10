package tickets

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"

	"webA/internal/respuesta"
)

type Ticket struct {
	ID        int    `json:"id"`
	Titulo    string `json:"titulo"`
	Prioridad string `json:"prioridad"` // baja | media | alta
	Estado    string `json:"estado"`    // abierto al crear
}

type Almacen struct {
	mu     sync.Mutex
	ultimo int
	items  map[int]Ticket
}

func NuevoAlmacen() *Almacen {
	return &Almacen{items: map[int]Ticket{}}
}

func (a *Almacen) crear(titulo, prioridad string) Ticket {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ultimo++
	t := Ticket{
		ID:        a.ultimo,
		Titulo:    titulo,
		Prioridad: prioridad,
		Estado:    "abierto",
	}
	a.items[t.ID] = t
	return t
}

func (a *Almacen) obtener(id int) (Ticket, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	t, ok := a.items[id]
	return t, ok
}

func (a *Almacen) listar() []Ticket {
	a.mu.Lock()
	defer a.mu.Unlock()
	lista := make([]Ticket, 0, len(a.items))
	for i := 1; i <= a.ultimo; i++ {
		if t, ok := a.items[i]; ok {
			lista = append(lista, t)
		}
	}
	return lista
}

type entradaCrear struct {
	Titulo    string `json:"titulo"`
	Prioridad string `json:"prioridad"`
}

var prioridadesValidas = map[string]bool{
	"baja":  true,
	"media": true,
	"alta":  true,
}

// GET /tickets
func (a *Almacen) Listar(w http.ResponseWriter, r *http.Request) {
	respuesta.Exito(w, http.StatusOK, a.listar())
}

// GET /tickets/{id}
func (a *Almacen) Obtener(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		respuesta.Error(w, http.StatusBadRequest, "id_invalido",
			"el id debe ser un entero positivo")
		return
	}
	t, ok := a.obtener(id)
	if !ok {
		respuesta.Error(w, http.StatusNotFound, "no_encontrado",
			"no existe un ticket con ese id")
		return
	}
	respuesta.Exito(w, http.StatusOK, t)
}

// POST /tickets — Fase 2: validación de JSON y reglas de negocio.
func (a *Almacen) Crear(w http.ResponseWriter, r *http.Request) {
	var entrada entradaCrear

	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))

	if err := dec.Decode(&entrada); err != nil {
		respuesta.Error(
			w,
			http.StatusBadRequest,
			"cuerpo_malformado",
			"el cuerpo debe ser JSON con los campos titulo y prioridad",
		)
		return
	}

	titulo := strings.TrimSpace(entrada.Titulo)
	if len(titulo) < 5 {
		respuesta.Error(w, http.StatusUnprocessableEntity, "datos_invalidos",
			"el título debe tener al menos 5 caracteres")
		return
	}
	if !prioridadesValidas[entrada.Prioridad] {
		respuesta.Error(w, http.StatusUnprocessableEntity, "datos_invalidos",
			"la prioridad debe ser baja, media o alta")
		return
	}

	t := a.crear(titulo, entrada.Prioridad)
	respuesta.Exito(w, http.StatusCreated, t)
}

// Explotar existe solo para probar el middleware de recuperación.
func Explotar(w http.ResponseWriter, r *http.Request) {
	var t *Ticket
	_ = t.Titulo // puntero nulo → pánico
}
