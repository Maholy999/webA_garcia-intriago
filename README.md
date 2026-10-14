# Servidor Garcia

> Reemplacen todo lo que está entre corchetes en su primer commit.
> Este README es la puerta de entrada del repositorio: en la semana 1 otra pareja debe poder levantar el servidor siguiendo solo lo que dice aquí, y desde la semana 4 es la base de la integración continua.

**Hilo Servidor · ULEAM · Período 2026-2**
Aplicaciones Web II (TDI-610) · Aplicación para el Servidor Web (IS-503)

## Integrantes

| Integrante                      | Usuario de GitHub      | Paralelo       |
| ------------------------------- | ---------------------- | -------------- |
| Garcia Intriago Maholy Maria    | Maholy999              | Servidor Web A |
| Garcia Peñarrieta Brittany Rubi | brittanygarcia2406-eng | Servidor Web A |

## El producto

[Una o dos líneas: qué negocio es y qué vende. Se completa cuando la ficha del negocio pase la compuerta de la semana 3. Mientras tanto, el dominio de trabajo es la mesa de ayuda.]

## Cómo levantar el servidor

```bash
go run .
```

El servidor responde en `http://localhost:8080`.

Requisitos: Go (versión estable) y PostgreSQL.

## Base de datos

- Opción usada por la pareja: [nativa / contenedor]
- Con contenedor: `docker run --name pg -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres:16`
- Base de datos del proyecto: `[nombre]`

## Configuración de sesión

Copie `.env.example` a `.env`, complete `DATABASE_URL` y cambie `CLAVE_FIRMA`
por una clave propia de al menos 32 caracteres. El archivo `.env` no se debe subir.

Para pruebas locales, el servidor crea estas cuentas de demostración:

| Usuario | Contraseña | Rol |
| --- | --- | --- |
| ana | ana-2026 | propietario |
| luis | luis-2026 | colaborador |

`POST /login` devuelve un objeto `{"token":"..."}`. En las demás rutas envíe
el encabezado `Authorization: Bearer <token>`.

| Ruta | Propietario | Colaborador |
| --- | --- | --- |
| `GET /tareas`, `GET /tareas/{id}` | Ver tareas | Ver tareas |
| `POST /tareas` | Crear y asignar tarea | Crear una tarea propia |
| `PUT`, `PATCH`, `DELETE /tareas/{id}` | Cualquier tarea | Solo sus tareas |
| `GET /usuarios` | Permitido | Prohibido |

El estado inicial siempre es `pendiente`; solo `PATCH` lo cambia, y las únicas
transiciones son `pendiente` → `en_proceso` → `completada`.

## Cómo correr las pruebas

```bash
go test ./...
```

Las pruebas corren también en la integración continua (pestaña Actions). Desde la semana 4, un entregable cuyas pruebas no pasan en la integración no se recibe.

## Convenciones del repositorio

- Un commit de cada integrante como mínimo por taller; el commit de cierre se hace en clase.
- Mensajes de commit: qué cambió y por qué, entendibles sin el autor presente.
- Uso de IA declarado en el cuerpo del commit: una línea con qué herramienta y para qué parte.
- Ningún secreto en el código ni en el historial: la configuración se externaliza (semana 4).

## Estructura

```
docs/ficha_negocio.md   → ficha del negocio (se entrega antes del día A de la semana 3)
main.go                 → punto de entrada del servidor
```
