# API de Contactos - Documentación de Endpoints

## Base URL
```
http://localhost:8080
```

## Endpoints

### 1. Health Check
Verifica el estado del servidor.

**Endpoint:** `GET /health`

**Respuesta:**
```json
{
  "status": "ok",
  "timestamp": "2024-12-07T10:30:00Z"
}
```

**Códigos de Estado:**
- `200 OK` - Servidor funcionando correctamente

---

### 2. Crear Contacto
Crea un nuevo contacto en la base de datos.

**Endpoint:** `POST /contact`

**Headers:**
```
Content-Type: application/json
```

**Body:**
```json
{
  "name": "Juan Pérez",
  "email": "juan@ejemplo.com",
  "message": "Hola, me interesa más información sobre sus servicios."
}
```

**Respuesta Exitosa:**
```json
{
  "message": "Contacto guardado exitosamente",
  "id": 1
}
```

**Respuesta de Error:**
```json
{
  "error": "Todos los campos son requeridos"
}
```

**Códigos de Estado:**
- `201 Created` - Contacto creado exitosamente
- `400 Bad Request` - Datos inválidos o campos faltantes
- `500 Internal Server Error` - Error del servidor

**Validaciones:**
- `name`: Requerido, no puede estar vacío
- `email`: Requerido, no puede estar vacío
- `message`: Requerido, no puede estar vacío

---

### 3. Listar Contactos
Obtiene una lista paginada de contactos ordenados por fecha de creación (más recientes primero).

**Endpoint:** `GET /contacts`

**Parámetros de Query (opcionales):**
- `page`: Número de página (default: 1)
- `limit`: Cantidad de elementos por página (default: 10, máximo: 100)

**Ejemplos de Uso:**
```
GET /contacts
GET /contacts?page=2
GET /contacts?page=1&limit=5
GET /contacts?limit=20
```

**Respuesta:**
```json
{
  "contacts": [
    {
      "id": 2,
      "name": "María González",
      "email": "maria@ejemplo.com",
      "message": "Necesito información sobre precios.",
      "created_at": "2024-12-07T10:25:00Z"
    },
    {
      "id": 1,
      "name": "Juan Pérez",
      "email": "juan@ejemplo.com",
      "message": "Hola, me interesa más información sobre sus servicios.",
      "created_at": "2024-12-07T10:20:00Z"
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 1,
    "total_items": 2,
    "items_per_page": 10
  }
}
```

**Códigos de Estado:**
- `200 OK` - Lista obtenida exitosamente
- `500 Internal Server Error` - Error del servidor

**Notas:**
- Los contactos se ordenan por fecha de creación descendente (más recientes primero)
- Si no hay contactos, se devuelve un array vacío
- Los parámetros de paginación inválidos se corrigen automáticamente

---

## Ejemplos de Uso con cURL

### Crear un contacto:
```bash
curl -X POST http://localhost:8080/contact \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Ana López",
    "email": "ana@ejemplo.com",
    "message": "Me gustaría agendar una reunión."
  }'
```

### Obtener contactos:
```bash
curl http://localhost:8080/contacts
```

### Obtener contactos con paginación:
```bash
curl "http://localhost:8080/contacts?page=2&limit=5"
```

### Verificar estado del servidor:
```bash
curl http://localhost:8080/health
```

---

## Códigos de Error Comunes

| Código | Descripción |
|--------|-------------|
| 200 | OK - Operación exitosa |
| 201 | Created - Recurso creado exitosamente |
| 400 | Bad Request - Datos inválidos o faltantes |
| 404 | Not Found - Endpoint no encontrado |
| 500 | Internal Server Error - Error del servidor |

---

## Notas Adicionales

- La API soporta CORS para desarrollo local
- Todos los timestamps están en formato ISO 8601 (UTC)
- Los campos de texto no tienen límite de caracteres definido
- La paginación tiene un límite máximo de 100 elementos por página