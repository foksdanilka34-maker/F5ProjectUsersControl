package docs

import "net/http"

// @title F5 Project Users Control API
// @version 1.0
// @description API для управления проектами и сотрудниками
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

const SwaggerJSON = `{
  "openapi": "3.0.0",
  "info": {
    "title": "F5 Project Users Control API",
    "version": "1.0",
    "description": "API для управления проектами и сотрудниками"
  },
  "servers": [{"url": "/api/v1"}],
  "components": {
    "securitySchemes": {
      "BearerAuth": {"type": "http", "scheme": "bearer", "bearerFormat": "JWT"}
    },
    "schemas": {
      "Error": {"type": "object", "properties": {"error": {"type": "string"}}},
      "Profile": {"type": "object", "properties": {"id": {"type": "integer"}, "first_name": {"type": "string"}, "last_name": {"type": "string"}, "email": {"type": "string"}, "position": {"type": "string"}, "department": {"type": "string"}, "role": {"type": "string"}}},
      "Project": {"type": "object", "properties": {"id": {"type": "integer"}, "name": {"type": "string"}, "description": {"type": "string"}, "status": {"type": "string"}, "manager_id": {"type": "integer"}}},
      "Task": {"type": "object", "properties": {"id": {"type": "integer"}, "title": {"type": "string"}, "description": {"type": "string"}, "status": {"type": "string"}, "priority": {"type": "integer"}, "project_id": {"type": "integer"}, "assignee_id": {"type": "integer"}}}
    }
  },
  "paths": {
    "/auth/login": {
      "post": {
        "tags": ["Auth"],
        "summary": "Авторизация",
        "requestBody": {"content": {"application/json": {"schema": {"type": "object", "required": ["login", "password"], "properties": {"login": {"type": "string"}, "password": {"type": "string"}}}}}},
        "responses": {"200": {"description": "OK"}, "401": {"description": "Unauthorized"}}
      }
    },
    "/auth/logout": {
      "post": {
        "tags": ["Auth"],
        "summary": "Выход",
        "security": [{"BearerAuth": []}],
        "responses": {"200": {"description": "OK"}}
      }
    },
    "/auth/refresh": {
      "post": {
        "tags": ["Auth"],
        "summary": "Обновить токен",
        "security": [{"BearerAuth": []}],
        "responses": {"200": {"description": "OK"}}
      }
    },
    "/auth/me": {
      "get": {
        "tags": ["Auth"],
        "summary": "Текущий пользователь",
        "security": [{"BearerAuth": []}],
        "responses": {"200": {"description": "OK"}}
      }
    },
    "/profiles": {
      "get": {
        "tags": ["Profiles"],
        "summary": "Список профилей",
        "security": [{"BearerAuth": []}],
        "parameters": [
          {"name": "page_size", "in": "query", "schema": {"type": "integer"}},
          {"name": "page_number", "in": "query", "schema": {"type": "integer"}},
          {"name": "department_id", "in": "query", "schema": {"type": "integer"}},
          {"name": "search", "in": "query", "schema": {"type": "string"}}
        ],
        "responses": {"200": {"description": "OK"}}
      },
      "post": {
        "tags": ["Profiles"],
        "summary": "Создать профиль",
        "security": [{"BearerAuth": []}],
        "requestBody": {"content": {"application/json": {"schema": {"type": "object", "required": ["first_name", "last_name", "position_id", "email", "login", "password"], "properties": {"first_name": {"type": "string"}, "last_name": {"type": "string"}, "position_id": {"type": "integer"}, "department_id": {"type": "integer"}, "email": {"type": "string"}, "login": {"type": "string"}, "password": {"type": "string"}, "role": {"type": "string"}}}}}},
        "responses": {"201": {"description": "Created"}}
      }
    },
    "/profiles/{id}": {
      "get": {
        "tags": ["Profiles"],
        "summary": "Получить профиль",
        "security": [{"BearerAuth": []}],
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "integer"}}],
        "responses": {"200": {"description": "OK"}, "404": {"description": "Not Found"}}
      },
      "put": {
        "tags": ["Profiles"],
        "summary": "Обновить профиль",
        "security": [{"BearerAuth": []}],
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "integer"}}],
        "requestBody": {"content": {"application/json": {"schema": {"type": "object", "properties": {"first_name": {"type": "string"}, "last_name": {"type": "string"}, "position_id": {"type": "integer"}, "department_id": {"type": "integer"}, "email": {"type": "string"}}}}}},
        "responses": {"200": {"description": "OK"}}
      },
      "delete": {
        "tags": ["Profiles"],
        "summary": "Удалить профиль",
        "security": [{"BearerAuth": []}],
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "integer"}}],
        "responses": {"200": {"description": "OK"}}
      }
    },
    "/departments": {
      "get": {
        "tags": ["Departments"],
        "summary": "Список отделов",
        "security": [{"BearerAuth": []}],
        "responses": {"200": {"description": "OK"}}
      },
      "post": {
        "tags": ["Departments"],
        "summary": "Создать отдел",
        "security": [{"BearerAuth": []}],
        "requestBody": {"content": {"application/json": {"schema": {"type": "object", "required": ["name"], "properties": {"name": {"type": "string"}, "description": {"type": "string"}}}}}},
        "responses": {"201": {"description": "Created"}}
      }
    },
    "/departments/{id}": {
      "get": {
        "tags": ["Departments"],
        "summary": "Получить отдел",
        "security": [{"BearerAuth": []}],
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "integer"}}],
        "responses": {"200": {"description": "OK"}}
      },
      "put": {
        "tags": ["Departments"],
        "summary": "Обновить отдел",
        "security": [{"BearerAuth": []}],
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "integer"}}],
        "responses": {"200": {"description": "OK"}}
      },
      "delete": {
        "tags": ["Departments"],
        "summary": "Удалить отдел",
        "security": [{"BearerAuth": []}],
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "integer"}}],
        "responses": {"200": {"description": "OK"}}
      }
    },
    "/positions": {
      "get": {
        "tags": ["Positions"],
        "summary": "Список должностей",
        "security": [{"BearerAuth": []}],
        "responses": {"200": {"description": "OK"}}
      },
      "post": {
        "tags": ["Positions"],
        "summary": "Создать должность",
        "security": [{"BearerAuth": []}],
        "requestBody": {"content": {"application/json": {"schema": {"type": "object", "required": ["name"], "properties": {"name": {"type": "string"}}}}}},
        "responses": {"201": {"description": "Created"}}
      }
    },
    "/skills": {
      "get": {
        "tags": ["Skills"],
        "summary": "Список навыков",
        "security": [{"BearerAuth": []}],
        "responses": {"200": {"description": "OK"}}
      },
      "post": {
        "tags": ["Skills"],
        "summary": "Создать навык",
        "security": [{"BearerAuth": []}],
        "requestBody": {"content": {"application/json": {"schema": {"type": "object", "required": ["name"], "properties": {"name": {"type": "string"}}}}}},
        "responses": {"201": {"description": "Created"}}
      }
    },
    "/projects": {
      "get": {
        "tags": ["Projects"],
        "summary": "Список проектов",
        "security": [{"BearerAuth": []}],
        "parameters": [
          {"name": "page_size", "in": "query", "schema": {"type": "integer"}},
          {"name": "page_number", "in": "query", "schema": {"type": "integer"}},
          {"name": "manager_id", "in": "query", "schema": {"type": "integer"}},
          {"name": "status", "in": "query", "schema": {"type": "string", "enum": ["PLANNING", "ACTIVE", "ON_HOLD", "COMPLETED", "CANCELLED"]}}
        ],
        "responses": {"200": {"description": "OK"}}
      },
      "post": {
        "tags": ["Projects"],
        "summary": "Создать проект",
        "security": [{"BearerAuth": []}],
        "requestBody": {"content": {"application/json": {"schema": {"type": "object", "required": ["name", "manager_id"], "properties": {"name": {"type": "string"}, "description": {"type": "string"}, "manager_id": {"type": "integer"}}}}}},
        "responses": {"201": {"description": "Created"}}
      }
    },
    "/projects/{id}": {
      "get": {
        "tags": ["Projects"],
        "summary": "Получить проект",
        "security": [{"BearerAuth": []}],
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "integer"}}],
        "responses": {"200": {"description": "OK"}}
      },
      "put": {
        "tags": ["Projects"],
        "summary": "Обновить проект",
        "security": [{"BearerAuth": []}],
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "integer"}}],
        "responses": {"200": {"description": "OK"}}
      },
      "delete": {
        "tags": ["Projects"],
        "summary": "Удалить проект",
        "security": [{"BearerAuth": []}],
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "integer"}}],
        "responses": {"200": {"description": "OK"}}
      }
    },
    "/projects/{id}/members": {
      "get": {
        "tags": ["Projects"],
        "summary": "Участники проекта",
        "security": [{"BearerAuth": []}],
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "integer"}}],
        "responses": {"200": {"description": "OK"}}
      },
      "post": {
        "tags": ["Projects"],
        "summary": "Добавить участника",
        "security": [{"BearerAuth": []}],
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "integer"}}],
        "requestBody": {"content": {"application/json": {"schema": {"type": "object", "required": ["user_id", "role"], "properties": {"user_id": {"type": "integer"}, "role": {"type": "string"}}}}}},
        "responses": {"200": {"description": "OK"}}
      }
    },
    "/tasks": {
      "get": {
        "tags": ["Tasks"],
        "summary": "Список задач",
        "security": [{"BearerAuth": []}],
        "parameters": [
          {"name": "project_id", "in": "query", "schema": {"type": "integer"}},
          {"name": "assignee_id", "in": "query", "schema": {"type": "integer"}},
          {"name": "status", "in": "query", "schema": {"type": "string", "enum": ["TODO", "IN_PROGRESS", "REVIEW", "DONE"]}},
          {"name": "priority", "in": "query", "schema": {"type": "string", "enum": ["LOW", "MEDIUM", "HIGH", "CRITICAL"]}}
        ],
        "responses": {"200": {"description": "OK"}}
      },
      "post": {
        "tags": ["Tasks"],
        "summary": "Создать задачу",
        "security": [{"BearerAuth": []}],
        "requestBody": {"content": {"application/json": {"schema": {"type": "object", "required": ["project_id", "title"], "properties": {"project_id": {"type": "integer"}, "title": {"type": "string"}, "description": {"type": "string"}, "assignee_id": {"type": "integer"}, "priority": {"type": "integer"}}}}}},
        "responses": {"201": {"description": "Created"}}
      }
    },
    "/tasks/{id}": {
      "get": {
        "tags": ["Tasks"],
        "summary": "Получить задачу",
        "security": [{"BearerAuth": []}],
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "integer"}}],
        "responses": {"200": {"description": "OK"}}
      },
      "put": {
        "tags": ["Tasks"],
        "summary": "Обновить задачу",
        "security": [{"BearerAuth": []}],
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "integer"}}],
        "responses": {"200": {"description": "OK"}}
      },
      "delete": {
        "tags": ["Tasks"],
        "summary": "Удалить задачу",
        "security": [{"BearerAuth": []}],
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "integer"}}],
        "responses": {"200": {"description": "OK"}}
      }
    },
    "/tasks/{id}/move": {
      "post": {
        "tags": ["Tasks"],
        "summary": "Переместить задачу (статус)",
        "security": [{"BearerAuth": []}],
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "integer"}}],
        "requestBody": {"content": {"application/json": {"schema": {"type": "object", "required": ["status"], "properties": {"status": {"type": "string"}}}}}},
        "responses": {"200": {"description": "OK"}}
      }
    },
    "/tasks/{id}/assign": {
      "post": {
        "tags": ["Tasks"],
        "summary": "Назначить исполнителя",
        "security": [{"BearerAuth": []}],
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "integer"}}],
        "requestBody": {"content": {"application/json": {"schema": {"type": "object", "required": ["assignee_id"], "properties": {"assignee_id": {"type": "integer"}}}}}},
        "responses": {"200": {"description": "OK"}}
      }
    },
    "/analytics/summary": {
      "get": {
        "tags": ["Analytics"],
        "summary": "Общая статистика",
        "security": [{"BearerAuth": []}],
        "responses": {"200": {"description": "OK"}}
      }
    },
    "/analytics/project/{id}": {
      "get": {
        "tags": ["Analytics"],
        "summary": "Аналитика проекта",
        "security": [{"BearerAuth": []}],
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "integer"}}],
        "responses": {"200": {"description": "OK"}}
      }
    },
    "/analytics/employee/{id}": {
      "get": {
        "tags": ["Analytics"],
        "summary": "Аналитика сотрудника",
        "security": [{"BearerAuth": []}],
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "integer"}}],
        "responses": {"200": {"description": "OK"}}
      }
    },
    "/analytics/dashboard": {
      "get": {
        "tags": ["Analytics"],
        "summary": "Дашборд",
        "security": [{"BearerAuth": []}],
        "responses": {"200": {"description": "OK"}}
      }
    },
    "/health": {
      "get": {
        "tags": ["System"],
        "summary": "Health check",
        "responses": {"200": {"description": "OK"}}
      }
    }
  }
}`

const SwaggerUIHTML = `<!DOCTYPE html>
<html><head>
<title>API Docs</title>
<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head><body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>SwaggerUIBundle({url:"/swagger/doc.json",dom_id:"#swagger-ui"});</script>
</body></html>`

func SwaggerHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(SwaggerJSON))
	})
	mux.HandleFunc("GET /swagger/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(SwaggerUIHTML))
	})
	mux.HandleFunc("GET /swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})
	return mux
}
