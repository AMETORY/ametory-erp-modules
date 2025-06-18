# ametory-erp-modules

## EXAMPLE INIT CODE

```go
package main

import (
	"ametory-erp/api/router"
	"ametory-erp/config"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/AMETORY/ametory-erp-modules/app"
	"github.com/gin-gonic/gin"
)

func main() { // Initialize the application container with options

	ctx := context.Background()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Gagal memuat konfigurasi: %v", err)
	}

	// Initialize database
	db, err := config.InitDB(cfg)
	if err != nil {
		log.Fatalf("Gagal menghubungkan ke database: %v", err)
	}

	fmt.Println("Berhasil terhubung ke database", cfg.Database.Name)
	appContainer := app.NewAppContainer(
		db,
		nil,
		&ctx,
		false,
		cfg.Server.BaseURL,
		app.WithAdminAuth(),
		app.WithHRIS(),
	)

	fmt.Println("Berhasil init AppContainer")

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
		})
	})

	v1 := r.Group("/api/v1")
	router.SetUpAuthRoutes(v1, appContainer)

	r.Run(":" + cfg.Server.Port)

}

```