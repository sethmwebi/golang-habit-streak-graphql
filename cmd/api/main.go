package main

import (
	"fmt"
	"log"
	"os"

	"habit-streak-graphql/internal/database"
	"habit-streak-graphql/internal/graph/generated"
	"habit-streak-graphql/internal/graph/resolvers"
	"habit-streak-graphql/internal/middleware"
	"habit-streak-graphql/internal/repository"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	db, err := database.InitDB("postgres://postgres:sethmwebi@localhost:5437/habits?sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	defer db.Close()

	fmt.Println("Database connected successfully!")

	userRepo := repository.NewUserRepository(db)
	habitRepo := repository.NewHabitRepository(db)
	habitLogRepo := repository.NewHabitLogRepository(db)

	graphqlHandler := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: &resolvers.Resolver{
			UserRepo:     userRepo,
			HabitRepo:    habitRepo,
			HabitLogRepo: habitLogRepo,
		},
	}))

	var router *gin.Engine = gin.Default()
	router.SetTrustedProxies(nil)

	router.GET("/playground", func(c *gin.Context) {
		playground.Handler("GraphQL Playground", "/graphql").ServeHTTP(c.Writer, c.Request)
	})

	router.POST("/graphql", middleware.AuthMiddleware(), func(ctx *gin.Context) {
		graphqlHandler.ServeHTTP(ctx.Writer, ctx.Request)
	})

	fmt.Println("GraphQL Playground: http://localhost:3000/playground")
	fmt.Println("GraphQL Endpoint: http://localhost:3000/graphql")

	port := os.Getenv("PORT")

	router.Run(":" + port)
}
