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
	"github.com/spf13/cobra"
)

func init() {

	rootCmd := &cobra.Command{
		Use:   "app",
		Short: "Application CLI",
	}

	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Running migrations...")
			// Add migration logic here
			initContainer(false)
		},
	}

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the application server",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Starting server...")
			appContainer := initContainer(true)

			r := gin.Default()

			r.GET("/", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "Hello, World!",
				})
			})

			v1 := r.Group("/api/v1")
			router.SetUpAuthRoutes(v1, appContainer)

			r.Run(":" + config.App.Server.Port)
		},
	}

	rootCmd.AddCommand(migrateCmd, serveCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command, %s", err)
	}
}

func initContainer(skipMigrate bool) *app.AppContainer {
	ctx := context.Background()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := config.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Successfully connected to database", cfg.Database.Name)

	return app.NewAppContainer(
		db,
		nil,
		&ctx,
		skipMigrate,
		cfg.Server.BaseURL,
		app.WithAdminAuth(),
		app.WithHRIS(),
	)
}

```

See all code in [example folder](example)

run server with
```bash
go run main.go serve
```

run migrate with
```bash
go run main.go migrate
```


https://github.com/user-attachments/assets/bde4f52e-efe8-45bf-a343-8e8637145ea6


### Use CLI Generator
```bash
go install github.com/AMETORY/ametory-erp-modules/cmd/erpgen@latest
```
