# Setlist API - Letterboxd for Music

## Stack

- Go v1.26.1
- GraphQL (GQLGen v0.17.88)
- Docker + Docker Compose

## Requirements

- Docker
- Docker Compose
- Go (for the Go extension)

## Getting Started

To run the project, you'll need to install Docker on your machine. Once installed, navigate to the root of the project and run:

```sh
docker compose up
```

Then open: http://localhost:8000 to view the GraphQL playground.

GQLGen Code Generation:

```sh
# sh needs to come after "api" if on macOS
docker compose exec api go generate ./...
```
