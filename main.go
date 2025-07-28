package backend

import (
	"log"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/routes"

	"github.com/gin-gonic/gin"
)

func Main() {
	err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config file: %s", err)
	}
	database.ConnectDB()
	run()
}

func run() {
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	r := gin.Default()

	routes.Init(r)

	r.Run(config.Current.Port)
	log.Println("Starting server on", config.Current.Port)
}
