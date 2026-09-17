package tareas

import "gorm.io/gorm"

// Sembrar carga datos de ejemplo si la tabla de usuarios está vacía
func Sembrar(db *gorm.DB) {
	var total int64
	db.Model(&Usuario{}).Count(&total)
	if total > 0 {
		return
	}

	datos := []Usuario{
		{
			Nombre: "Juan Pérez",
			Tareas: []Tarea{
				{
					Titulo: "Revisar proyector sin señal",
					Estado: "pendiente",
				},
				{
					Titulo: "Configurar aula virtual",
					Estado: "en_proceso",
				},
			},
		},
		{
			Nombre: "María López",
			Tareas: []Tarea{
				{
					Titulo: "Restablecer contraseña de acceso",
					Estado: "completada",
				},
			},
		},
	}

	db.Create(&datos)
}
