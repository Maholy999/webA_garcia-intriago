package tareas

var transicionesPermitidas = map[string]map[string]bool{
	"pendiente":  {"en_proceso": true},
	"en_proceso": {"completada": true},
	"completada": {},
}

// TransicionPermitida indica si una tarea puede pasar del estado actual al destino.
func TransicionPermitida(actual, destino string) bool {
	return transicionesPermitidas[actual][destino]
}
