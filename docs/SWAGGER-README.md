Swagger UI is available at http://localhost:8081 after running docker compose up.

How it works:
- The docker-compose mounts ai-forge/openapi.yaml into the swagger container at /openapi.yaml.
- The container uses SWAGGER_JSON=/openapi.yaml so the UI loads the spec.

Start:
  docker compose up -d
Open: http://localhost:8081

Notes:
- If port 8081 conflicts, change ports mapping in docker-compose.yml.
- The UI accesses your API endpoints directly (http://localhost:8080). For remote access, ensure CORS and network reachability.
