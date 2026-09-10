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

## Campo desconocido en el cuerpo

Rechazar el cuerpo con
`dec.DisallowUnknownFields()` si trae un campo que no es `titulo` o `prioridad`.
Explicación,
Un campo como `"estado":"cerrado"` en la creación sugiere que el cliente cree que puede fijar el estado del ticket,
cosa que no le corresponde. Ignorarlo en silencio esconde ese error del lado del cliente; rechazarlo lo hace visible de inmediato.

Alternativa descartada:
ignorarlo en silencio (comportamiento por defecto de `encoding/json`). Es más tolerante, pero permite
que un cliente mal hecho siga funcionando sin darse cuenta de su error.
Cómo lo comprobamos:
`curl -i -X POST localhost:8080/tickets -d '{"titulo":"Pantalla rota","prioridad":"alta","estado":"cerrado"}'`
El `Decode` falla y responde `cuerpo_malformado` (o el código que hayan definido para este caso).
