# Postboard

Gestor de publicaciones programadas para redes sociales.
**Stack:** Go (API REST) · PostgreSQL · React + TypeScript · Docker · GitHub Actions

## Estructura

```
postboard/
├── api/
│   ├── cmd/api/main.go          # Arranque, middleware (CORS, logs), graceful shutdown
│   ├── internal/
│   │   ├── models/post.go       # Modelo Post + validación
│   │   ├── store/               # Interfaz PostStore + Postgres + memoria
│   │   └── handlers/            # Endpoints HTTP + tests
│   ├── migrations/              # SQL de la tabla posts
│   ├── Dockerfile               # Build multi-stage (imagen final ~10 MB)
│   └── Makefile
├── web/                         # (Día 8) Frontend React
├── docker-compose.yml
└── .github/workflows/ci.yml
```

## Primeros pasos

Requisitos: Go 1.24+, Docker Desktop.

```bash
cd api
go mod tidy        # descarga dependencias y crea go.sum (solo la primera vez)
make test          # ejecuta los tests
make run           # arranca la API en memoria, sin base de datos -> http://localhost:8080
```

Con PostgreSQL (desde la raíz del proyecto):

```bash
docker compose up --build
```

## Endpoints

| Método | Ruta          | Descripción          |
|--------|---------------|----------------------|
| GET    | /health       | Estado de la API     |
| GET    | /posts        | Lista de posts       |
| GET    | /posts/{id}   | Un post              |
| POST   | /posts        | Crear post           |
| PUT    | /posts/{id}   | Editar post          |
| DELETE | /posts/{id}   | Borrar post          |

Ejemplo:

```bash
curl -X POST localhost:8080/posts \
  -H "Content-Type: application/json" \
  -d '{"title":"Lanzamiento","content":"Hola","platform":"instagram","status":"scheduled","scheduled_at":"2026-12-01T10:00:00Z"}'
```

Valores válidos: `platform` = instagram | facebook | linkedin · `status` = draft | scheduled | published

## Decisiones técnicas

- **`net/http` sin framework:** desde Go 1.22 el router nativo admite métodos y parámetros en la ruta (`GET /posts/{id}`).
- **Interfaz `PostStore`:** los handlers no dependen de Postgres; en los tests se usa una implementación en memoria.
- **Table-driven tests** con `httptest`.
- **Graceful shutdown:** al recibir SIGTERM, el servidor termina las peticiones en curso antes de cerrar.
- **Imagen distroless** sin shell y con usuario no root.

## Próximos pasos

- [ ] Día 6: migraciones con golang-migrate
- [ ] Días 8–9: frontend React (`npm create vite@latest web -- --template react-ts`)
- [ ] Día 10: autenticación JWT (opcional)
- [ ] Día 11: servicio web en docker-compose
- [ ] Día 12: job de frontend en CI
- [ ] Días 13–14: deploy + capturas en este README
