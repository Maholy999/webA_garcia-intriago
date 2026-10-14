package tareas

// Usuario representa la entidad del lado del uno
type Usuario struct {
	ID     uint    `json:"id"`
	Nombre string  `json:"nombre"`
	Tareas []Tarea `json:"tareas,omitempty"`
}

// Tarea representa la entidad con estados (lado de los muchos)
type Tarea struct {
	ID        uint   `json:"id"`
	Titulo    string `json:"titulo"`
	Estado    string `json:"estado"`
	UsuarioID uint   `json:"usuario_id"`
}

// NuevaTarea representa los campos que el cliente puede enviar al crear una tarea.
type NuevaTarea struct {
	Titulo    string  `json:"titulo"`
	UsuarioID uint    `json:"usuario_id"`
	Estado    *string `json:"estado,omitempty"`
}

// CambioEstado contiene la transición solicitada por PATCH.
type CambioEstado struct {
	Estado string `json:"estado"`
}

var estadosValidos = map[string]bool{
	"pendiente":  true,
	"en_proceso": true,
	"completada": true,
}
