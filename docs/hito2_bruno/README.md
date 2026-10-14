# Colección Bruno · Hito 2

Importe esta carpeta como colección en Bruno y cree un entorno con estas
variables:

| Variable | Valor inicial |
| --- | --- |
| `baseUrl` | `http://localhost:8080` |
| `tokenPropietario` | Vacío; lo guarda la petición 2 |
| `tokenColaborador` | Vacío; lo guarda `00-token-colaborador.bru` |
| `usuarioPropietario` | ID que corresponde a Ana en `GET /usuarios` |
| `tareaDelPropietario` | ID de la primera tarea de prueba |
| `tareaProhibida` | ID de la segunda tarea de prueba |

## Preparar los registros y tokens

1. Arranque el servidor con la base de datos configurada.
2. Ejecute `00-token-colaborador.bru`; verifica el acceso con el usuario `luis`
   y guarda su token en `tokenColaborador`.
3. Ejecute `00-token-propietario.bru` y luego las dos peticiones
   `00-crear-tarea-*.bru`. Las tareas nuevas comienzan en `pendiente`; ambas
   deben quedar asignadas al usuario propietario. Copie los IDs devueltos a
   `tareaDelPropietario` y `tareaProhibida`.
4. Ejecute las peticiones `01` a `06` en orden. No reutilice los registros si
   repite la demostración: la petición 05 cambia el primero a `en_proceso`.

## Resultado esperado

| Petición | Resultado |
| --- | --- |
| 01 — listar sin sesión | 401 |
| 02 — iniciar sesión como Ana | 200 y token |
| 03 — listar con sesión de Ana | 200 |
| 04 — Luis intenta cambiar la tarea de Ana | 403 |
| 05 — Ana cambia la misma tarea a `en_proceso` | 200 |
| 06 — Ana intenta pasar otra tarea de `pendiente` a `completada` | 422 |

Para la fase 1.4, use una tarea nueva y haga tres PATCH: `inventado` debe
responder 422 con `estado no permitido: inventado`; `completada` desde
`pendiente` debe responder 422 con `de pendiente no se puede pasar a
completada`; y `en_proceso` desde `pendiente` debe responder 200.
