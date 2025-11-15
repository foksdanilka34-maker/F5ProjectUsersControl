// @title            F5 API Gateway
// @version          1.0
// @description      HTTP gateway for employee, project, login, and analytics services.
// @contact.name     F5 Project Team
// @contact.url      https://f5.example.com
// @host             localhost:8080
// @BasePath         /api/v1
// @schemes          http https
// @securityDefinitions.apikey ApiKeyAuth
// @in               header
// @name             Authorization
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	docs "github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/docs"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/internal/config"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/internal/handlers"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/internal/middleware"
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/ApiGateway/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	log.Println("Starting API Gateway")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	_ = godotenv.Load()

	cfg := config.NewConfig()

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	configureSwagger(cfg)

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

	analyticsServiceClient := service.NewAnalyticsServiceClient(
		cfg.AnalyticsServiceHost,
		cfg.AnalyticsServicePort,
	)
	defer func() {
		if err := analyticsServiceClient.Close(); err != nil {
			log.Printf("failed to close analytics service client: %v", err)
		}
	}()

	authHandler := handlers.NewAuthHandler(loginServiceClient, cfg)
	employeeHandler := handlers.NewEmployeeHandler(employeeServiceClient)
	projectHandler := handlers.NewProjectHandler(projectServiceClient)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsServiceClient, projectServiceClient)

	router := buildRouter(cfg, handlerBundle{
		auth:      authHandler,
		employee:  employeeHandler,
		project:   projectHandler,
		analytics: analyticsHandler,
	})

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

type handlerBundle struct {
	auth      *handlers.AuthHandler
	employee  *handlers.EmployeeHandler
	project   *handlers.ProjectHandler
	analytics *handlers.AnalyticsHandler
}

func buildRouter(cfg *config.Config, bundle handlerBundle) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(
		middleware.LoggingMiddleware(),
		middleware.ErrorHandlerMiddleware(),
		middleware.CORSMiddleware(),
	)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	registerAuthRoutes(v1, bundle.auth, cfg.JWTSecret)
	registerEmployeeRoutes(v1, bundle.employee, cfg.JWTSecret)
	registerProjectRoutes(v1, bundle.project, cfg.JWTSecret)
	registerAnalyticsRoutes(v1, bundle.analytics, cfg.JWTSecret)

	return router
}

func registerAuthRoutes(api *gin.RouterGroup, handler *handlers.AuthHandler, jwtSecret string) {
	auth := api.Group("/auth")
	auth.POST("/login", handler.Login)
	auth.POST("/refresh", handler.Refresh)

	withAuth(auth, jwtSecret).POST("/logout", handler.Logout)
}

func registerEmployeeRoutes(api *gin.RouterGroup, handler *handlers.EmployeeHandler, jwtSecret string) {
	employees := withAuth(api.Group("/employees"), jwtSecret, "admin")

	employees.POST("/profiles", handler.CreateProfile)
	employees.GET("/profiles", handler.ListProfiles)
	employees.GET("/profiles/:id", handler.GetProfile)
	employees.PATCH("/profiles/:id", handler.UpdateProfile)
	employees.PATCH("/profiles/:id/status", handler.ChangeUserStatus)

	employees.POST("/departments", handler.CreateDepartment)
	employees.GET("/departments", handler.ListDepartments)
	employees.GET("/departments/:id", handler.GetDepartment)
	employees.PUT("/departments/:id", handler.UpdateDepartment)
	employees.DELETE("/departments/:id", handler.DeleteDepartment)

	employees.POST("/positions", handler.CreatePosition)
	employees.GET("/positions", handler.ListPositions)
	employees.GET("/positions/:id", handler.GetPosition)
	employees.PUT("/positions/:id", handler.UpdatePosition)
	employees.DELETE("/positions/:id", handler.DeletePosition)

	employees.POST("/skills", handler.CreateSkill)
	employees.GET("/skills", handler.ListSkills)
	employees.POST("/profiles/:id/skills", handler.AddSkillToEmployee)
	employees.DELETE("/profiles/:id/skills/:skillId", handler.RemoveSkillFromEmployee)
}

func registerProjectRoutes(api *gin.RouterGroup, handler *handlers.ProjectHandler, jwtSecret string) {
	projects := withAuth(api.Group("/projects"), jwtSecret)
	projects.GET("", handler.ListProjects)
	projects.GET("/:id/members", handler.ListProjectMembers)
	projects.GET("/:id", handler.GetProject)

	withRoles(projects, "manager", "director").POST("", handler.CreateProject)
	withRoles(projects, "manager", "director", "admin").PATCH("/:id", handler.UpdateProject)
	withRoles(projects, "manager", "director", "admin").DELETE("/:id", handler.DeleteProject)
	withRoles(projects, "manager", "director").POST("/:id/members", handler.AddMemberToProject)
	withRoles(projects, "manager", "director").DELETE("/:id/members/:memberId", handler.RemoveMemberFromProject)

	projectTasks := projects.Group("/:id/tasks")
	projectTasks.GET("", handler.ListTasksByProject)
	projectTasks.POST("", handler.CreateTask)
	projectTasks.GET("/:taskId", handler.GetTask)
	projectTasks.PATCH("/:taskId", handler.UpdateTask)
	projectTasks.DELETE("/:taskId", handler.DeleteTask)
	projectTasks.POST("/:taskId/move", handler.MoveTask)

	withRoles(projectTasks, "manager", "director").POST("/:taskId/assign", handler.AssignTask)
}

func registerAnalyticsRoutes(api *gin.RouterGroup, handler *handlers.AnalyticsHandler, jwtSecret string) {
	analytics := withAuth(api.Group("/analytics"), jwtSecret, "director")
	analytics.GET("/dashboard", handler.GetDashboardStats)
	analytics.GET("/employees/metrics", handler.ListEmployeeMetrics)
	analytics.GET("/employees/top-performers", handler.GetTopPerformers)
	analytics.GET("/employees/:id/metrics", handler.GetEmployeeMetrics)
	analytics.GET("/projects/metrics", handler.ListProjectMetrics)
	analytics.GET("/projects/:id/metrics", handler.GetProjectMetrics)
	analytics.GET("/trends/productivity", handler.GetProductivityTrends)
	analytics.GET("/trends/completion-rate", handler.GetCompletionRateTrends)
}

func withAuth(group *gin.RouterGroup, jwtSecret string, roles ...string) *gin.RouterGroup {
	protected := group.Group("")
	protected.Use(middleware.AuthMiddleware(jwtSecret))
	if len(roles) > 0 {
		protected.Use(middleware.RoleMiddleware(roles...))
	}
	return protected
}

func withRoles(group *gin.RouterGroup, roles ...string) *gin.RouterGroup {
	if len(roles) == 0 {
		return group
	}
	sub := group.Group("")
	sub.Use(middleware.RoleMiddleware(roles...))
	return sub
}

func configureSwagger(cfg *config.Config) {
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Title = "F5 API Gateway"
	docs.SwaggerInfo.Description = "HTTP gateway for employee, project, login, and analytics services"
	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	scheme := "http"
	if cfg.Environment == "production" {
		scheme = "https"
	}
	docs.SwaggerInfo.Schemes = []string{scheme}
}
