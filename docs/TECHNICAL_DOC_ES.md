# Documentación Técnica de Zaptools-go

## Visión General
`zaptools-go` es un kit de herramientas para la gestión de WebSockets orientados a eventos en Go. Abstrae la implementación subyacente del WebSocket (como Gorilla WebSocket, Echo o Fiber) a través de una interfaz estándar, permitiendo a los desarrolladores construir aplicaciones en tiempo real basadas en eventos de forma similar a Socket.io.

## Arquitectura

### 1. Interfaz `StdConn`
La abstracción central de la biblioteca. Cualquier framework de WebSocket que implemente esta interfaz puede ser utilizado nativamente con Zaptools.
```go
type StdConn interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	Close() error
}
```

### 2. Registro de Eventos (`EventRegister` & `EventBook`)
- **`EventBook`**: Un mapa concurrente seguro para hilos (usando `sync.RWMutex`) que almacena los eventos registrados evitando colisiones (Data Races).
- **`EventRegister`**: Expone el método `OnEvent` para vincular nombres de eventos (cadenas de texto) a funciones *callback* (`func(ctx *EventContext)`).

### 3. Ciclo de Vida de la Conexión (`ZapConnector` & `EventProcessor`)
Cuando ocurre una nueva conexión WebSocket, se crea un `ZapConnector` con `ConnectorOptions`. Al llamar a `Start()`, se inicializa un `EventProcessor` que coordina el ciclo de vida:
1. Dispara el evento `"connected"`.
2. Inicia una *goroutine* dedicada exclusivamente a escribir mensajes salientes (el *Write Stream*).
3. Entra en un bucle bloqueante para leer los mensajes entrantes desde el socket.
4. Analiza el JSON entrante convirtiéndolo en un objeto `EventData` y dispara el *callback* correspondiente.
5. Al desconectarse o fallar la conexión, cierra de forma segura los flujos de datos mediante un canal `done` y dispara el evento `"disconnected"`.

### 4. Concurrencia y Seguridad (Thread-Safety)
- **Flujo de Escritura**: Todos los mensajes salientes se envían a un canal interno (`writeChannel`). Una goroutine dedicada escucha este canal y escribe en el socket TCP, evitando que clientes con conexión lenta bloqueen el hilo principal u otras conexiones.
- **Sincronización de Canales**: La biblioteca utiliza el patrón de canal `done` y `sync.Once` para prevenir escenarios de `panic` (como intentar escribir en un canal cerrado) durante desconexiones concurrentes.
- **Mapas Seguros**: El almacenamiento de eventos y los estados de conexión (`isConnected`) están fuertemente protegidos por candados de lectura y escritura (`sync.RWMutex`).

## Referencia de la API

### `EventData`
El payload JSON estándar esperado y enviado por la biblioteca.
```json
{
  "eventName": "string",
  "payload": "any",
  "headers": { "key": "value" }
}
```

### `EventContext`
Se pasa como argumento a cada callback de evento. Proporciona acceso a los datos entrantes (`EventData`) y al objeto `WebSocketConnection` para poder enviar respuestas directamente a ese cliente.
```go
func(ctx *zap.EventContext) {
    // Acceder al contenido del evento
    data := ctx.EventData.Payload
    
    // Enviar una respuesta al cliente
    ctx.Connection.SendEvent(&zap.EventData{...})
}
```

### Eventos de Sistema Integrados
Zaptools reserva ciertos nombres de eventos para el ciclo de vida de la conexión:
- `"connected"`: Se dispara en el momento en que un `ZapConnector` inicia exitosamente la conexión.
- `"disconnected"`: Se dispara cuando el socket se cierra limpiamente o la lectura de red falla.
- `"error"`: Se dispara por errores de sintaxis (parsing) del JSON u origen de datos corruptos.
- `"send-error"`: Se dispara si ocurre un fallo interno al intentar escribir un mensaje en el socket TCP subyacente.
