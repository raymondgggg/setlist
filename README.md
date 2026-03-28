# Setlist - Letterboxd for Music

## Getting started

To run the project, you'll need to install Docker on your machine. Once installed, navigate to the root of the project and run:

```sh
docker compose up
```

GQLGen Code Generation:

```sh
# sh needs to come after "api" if on macOS
docker compose exec api go generate ./...
```
