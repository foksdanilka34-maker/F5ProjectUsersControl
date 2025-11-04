package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/internal/config"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/internal/handlers"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/internal/middleware"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("Starting API Gateway")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	_ = godotenv.Load()

	cfg := config.NewConfig()

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	loginServiceClient := service.NewLoginServiceClient(
		cfg.LoginServiceHost,
		cfg.LoginServicePort,
		cfg.JWTSecret,
	)
	defer func() {
		if err := loginServiceClient.Close(); err != nil {
			log.Printf("failed to close login service client: %v", err)
		}
	}()

	employeeServiceClient := service.NewEmployeeServiceClient(
		cfg.EmployeeServiceHost,
		cfg.EmployeeServicePort,
	)
	defer func() {
		if err := employeeServiceClient.Close(); err != nil {
			log.Printf("failed to close employee service client: %v", err)
		}
	}()

	projectServiceClient := service.NewProjectServiceClient(
		cfg.ProjectServiceHost,
		cfg.ProjectServicePort,
	)
	defer func() {
		if err := projectServiceClient.Close(); err != nil {
			log.Printf("failed to close project service client: %v", err)
		}
	}()

	router := gin.Default()

	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.ErrorHandlerMiddleware())
	router.Use(middleware.CORSMiddleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	authHandler := handlers.NewAuthHandler(loginServiceClient, cfg)
	employeeHandler := handlers.NewEmployeeHandler(employeeServiceClient)
	projectHandler := handlers.NewProjectHandler(projectServiceClient)

	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)

			protected := auth.Group("")
			protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
			{
				protected.POST("/logout", authHandler.Logout)
			}

			admin := auth.Group("")
			admin.Use(middleware.AuthMiddleware(cfg.JWTSecret))
			admin.Use(middleware.RoleMiddleware("admin", "director"))
			{
				// admin.POST("/credentials", authHandler.CreateCredentials)
				// admin.PATCH("/users/:userID/status", authHandler.ChangeUserStatus)
			}
		}

		// Employee routes - все требуют роль admin
		employees := v1.Group("/employees")
		employees.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		employees.Use(middleware.RoleMiddleware("admin"))
		{
			// Profile endpoints
			employees.POST("/profiles", employeeHandler.CreateProfile)
			employees.GET("/profiles", employeeHandler.ListProfiles)
			employees.GET("/profiles/:id", employeeHandler.GetProfile)
			employees.PATCH("/profiles/:id", employeeHandler.UpdateProfile)
			employees.PATCH("/profiles/:id/status", employeeHandler.ChangeUserStatus)

			// Department endpoints
			employees.POST("/departments", employeeHandler.CreateDepartment)
			employees.GET("/departments", employeeHandler.ListDepartments)
			employees.GET("/departments/:id", employeeHandler.GetDepartment)
			employees.PUT("/departments/:id", employeeHandler.UpdateDepartment)
			employees.DELETE("/departments/:id", employeeHandler.DeleteDepartment)

			// Position endpoints
			employees.POST("/positions", employeeHandler.CreatePosition)
			employees.GET("/positions", employeeHandler.ListPositions)
			employees.GET("/positions/:id", employeeHandler.GetPosition)
			employees.PUT("/positions/:id", employeeHandler.UpdatePosition)
			employees.DELETE("/positions/:id", employeeHandler.DeletePosition)

			// Skill endpoints
			employees.POST("/skills", employeeHandler.CreateSkill)
			employees.GET("/skills", employeeHandler.ListSkills)
			employees.POST("/profiles/:id/skills", employeeHandler.AddSkillToEmployee)
			employees.DELETE("/profiles/:id/skills/:skillId", employeeHandler.RemoveSkillFromEmployee)
		}

		// Project routes - разные уровни доступа
		projects := v1.Group("/projects")
		projects.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// Читать проекты могут все авторизованные пользователи
			projects.GET("", projectHandler.ListProjects)
			projects.GET("/:id", projectHandler.GetProject)
			projects.GET("/:projectId/members", projectHandler.ListProjectMembers)
			projects.GET("/:projectId/tasks", projectHandler.ListTasksByProject)
			projects.GET("/:projectId/metrics", projectHandler.GetProjectMetrics)

			// Создавать проекты могут только manager и director
			managerOrDirector := projects.Group("")
			managerOrDirector.Use(middleware.RoleMiddleware("manager", "director"))
			{
				managerOrDirector.POST("", projectHandler.CreateProject)
			}

			// Обновлять и удалять могут manager, director и admin
			managerDirectorAdmin := projects.Group("")
			managerDirectorAdmin.Use(middleware.RoleMiddleware("manager", "director", "admin"))
			{
				managerDirectorAdmin.PATCH("/:id", projectHandler.UpdateProject)
				managerDirectorAdmin.DELETE("/:id", projectHandler.DeleteProject)
			}

			// Управление участниками проекта - только manager (создатель проекта)
			// TODO: добавить проверку, что пользователь является создателем проекта
			managerOnly := projects.Group("")
			managerOnly.Use(middleware.RoleMiddleware("manager", "director"))
			{
				managerOnly.POST("/:projectId/members", projectHandler.AddMemberToProject)
				managerOnly.DELETE("/:projectId/members/:memberId", projectHandler.RemoveMemberFromProject)
			}
		}

		// Task routes - разные уровни доступа
		tasks := v1.Group("/tasks")
		tasks.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// Читать таски могут все авторизованные
			tasks.GET("/:id", projectHandler.GetTask)
			tasks.GET("/:id/history", projectHandler.GetTaskStatusHistory)

			// Обновлять и перемещать таски могут все авторизованные (участники проекта)
			tasks.PATCH("/:id", projectHandler.UpdateTask)
			tasks.POST("/:id/move", projectHandler.MoveTask)
			tasks.DELETE("/:id", projectHandler.DeleteTask)

			// Создавать и назначать таски - только manager и director (создатели проекта)
			managerDirectorTasks := tasks.Group("")
			managerDirectorTasks.Use(middleware.RoleMiddleware("manager", "director"))
			{
				// POST /tasks будет создаваться через /projects/:projectId/tasks
			}
		}

		// Создание тасков через проект
		v1.POST("/projects/:projectId/tasks",
			middleware.AuthMiddleware(cfg.JWTSecret),
			middleware.RoleMiddleware("manager", "director"),
			projectHandler.CreateTask,
		)
		// Назначение таска
		v1.POST("/tasks/:id/assign",
			middleware.AuthMiddleware(cfg.JWTSecret),
			middleware.RoleMiddleware("manager", "director"),
			projectHandler.AssignTask,
		)
	}

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	log.Printf("Starting API Gateway on %s", addr)

	go func() {
		if err := router.Run(addr); err != nil && err.Error() != "http: Server closed" {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down API Gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = ctx
	log.Println("API Gateway gracefully stopped")
}
