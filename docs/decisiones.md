# Semana 7 · D1 — Máquina de estados de Tarea

| Desde | Se puede pasar a |
| --- | --- |
| `pendiente` | `en_proceso` |
| `en_proceso` | `completada` |
| `completada` | Ningún estado (final) |

**Transición prohibida elegida:** `pendiente` → `completada`. No se marca una
tarea como terminada sin que haya pasado por `en_proceso`. El servidor responde
422 porque entiende el estado solicitado, pero la regla del negocio no permite
esa transición.

## Semana 7 · D2 — Roles y niveles de acceso

| Ruta | Nivel | Propietario | Colaborador |
| --- | --- | --- | --- |
| `POST /login` | Pública | Inicia sesión | Inicia sesión |
| `GET /tareas`, `GET /tareas/{id}` | Con sesión | Ve las tareas del tablero | Ve las tareas del tablero |
| `POST /tareas` | Con sesión | Crea y asigna a cualquier usuario | Crea tareas para sí mismo |
| `PUT /tareas/{id}`, `PATCH /tareas/{id}`, `DELETE /tareas/{id}` | Con sesión | Gestiona cualquier tarea | Gestiona solo las tareas propias |
| `GET /usuarios` | Solo propietario | Permitido | 403 |

Las rutas con sesión exigen un JWT vigente y admiten ambos roles. La ruta de
usuarios se agrupa aparte para exigir el rol `propietario`. El propietario puede
actuar sobre las tareas del tablero; al colaborador se le compara el ID de la
sesión con `tarea.UsuarioID`. Una sesión ausente o inválida responde 401; un rol
o propietario incorrecto responde 403; una transición de negocio prohibida
responde 422.

## Semana 7 · D3 — Prueba que protege las rutas

`TestSinTokenLasRutasProtegidasResponden401` enumera las rutas no públicas
actuales, incluidas las de sesión y la ruta de un solo rol. Si una ruta se saca
accidentalmente de su grupo de autenticación, la prueba falla e identifica el
método y la ruta. Para la demostración se comprueba además el mismo cambio de
estado con un colaborador (403), con el propietario (200) y con una transición
prohibida (422).

## orden de middleware

r.Use(middleware.Registro)
r.Use(middleware.Recuperacion)

Explicación
sí Registro sigue vivo aunque el handler entre en pánico, porque el `recover()` de Recuperación detiene el pánico
antes de que suba y rompa la capa de Registro. Esto permite que cada petición, incluso las que fallan por dentro, quede anotada en el log con su código real.

Descartada.
Recuperación afuera y Registro adentro. Con ese orden el pánico salta por encima de Registro (que no tiene `recover`) y la línea de log de esa petición nunca se escribe, aunque el cliente sí reciba el 500.

## 2.5 Código para JSON roto frente a datos inválidos

cuerpo_malformado - 400
datos_invalidos - 422

Explicación.
un 400 comunica no entendí lo que me mandaste, mientras que un 422 comunica te entendí,
pero el contenido no cumple las reglas.

## Estado inicial de una tarea

Una tarea nueva empieza en `pendiente`. El cliente puede omitir `estado` o
enviarlo como `pendiente`; cualquier otro estado inicial se rechaza con 422.
Los cambios posteriores se hacen con `PATCH /tareas/{id}`, donde se valida la
máquina de estados de D1.

Cómo lo comprobamos:
`curl -i -X POST localhost:8080/tareas -H "Content-Type: application/json" -d '{"titulo":"Revisar proyector","estado":"en_proceso"}'`
La petición responde `422` con el código `estado_inicial`.

## Semana 4 · D3 · La prueba que no escribimos
**Decisión:** La fila estrella de nuestra E4 es la prueba `estado inventado responde 422` del `TestCrear`, porque valida la regla más importante del flujo antes de tocar la base: un `estado` que no está en el mapa debe rechazarse con 422.
**Por qué:** En el manejador de creación, la validación ocurre antes de cualquier consulta a la DB; si el JSON llega roto o con un estado no permitido, el controlador responde inmediatamente sin consultar `m.DB`. Eso es exactamente lo que queríamos probar para asegurar que la lógica de validación no cruza la línea de la base.
**Qué descartamos:** Descartamos la prueba que cruza la base con `m.DB.Debug()` en un caso inválido, porque esa prueba termina en un `nil pointer dereference` y no comprueba la regla de negocio; lo correcto es dejarla fuera o reordenar el código para validar primero. La línea de inicio real de la base en el archivo terminado es `internal/tareas/manejadores.go`, línea 80, donde aparece `query := m.DB.Debug()`. Qué haría falta para probarla: un caso de `GET /tareas?estado=urgente` con `DB: nil` y una expectativa de 422, que confirme que el filtro del query string se valida antes de tocar la base, sin implementarlo aquí.
