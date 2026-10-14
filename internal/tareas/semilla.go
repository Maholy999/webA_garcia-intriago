package tareas

import "gorm.io/gorm"

// Sembrar carga datos de ejemplo si la tabla de usuarios está vacía
func Sembrar(db *gorm.DB) error {
	var total int64
	if err := db.Model(&Usuario{}).Count(&total).Error; err != nil {
		return err
	}
	if total > 0 {
		return nil
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

	return db.Create(&datos).Error
}
