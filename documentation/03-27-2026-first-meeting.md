# App Idea

Music Version of Letterboxd

## Tech Stack

### Core

- DB
  - Postgres
- React Native with Expo
- Functional or Object Oriented? (React)
  - Raymond prefers Functional
  - Brent prefers Functional
- Go
  - Gin
  - GORM
  - GQLGen
- Rest or GraphQL?
  - GraphQL
- Should we do automated testing?
  - Ray -> yes
  - Brent -> yes
  - Go: standard library, testify
- Should we do CI/CD?
  - Ray -> yes
  - Brent -> yes
  - Github Actions -> job to build & package the app, run automated tests, etc.
- Containerized?
  - Ray -> yes
  - Brent -> yes
  - Docker Compose with 3 containers: DB, UI, and API
- Cloud Platform
  - AWS
- Infrastructure as Code
  - Terraform (?)
  - Ansible (?)
  - Ray -> yes
  - Brent -> yes
- Domain
  - Buy one once we figure out the name

### Integrations

- Spotify API
- Apple Music API (future)

## Platforms to Target

- iOS
- Android
- Web
