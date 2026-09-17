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

// Lista de estados válidos exigida por el taller
var estadosValidos = map[string]bool{
	"pendiente":  true,
	"en_proceso": true,
	"completada": true,
}
