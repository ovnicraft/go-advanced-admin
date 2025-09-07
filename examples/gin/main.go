package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ovnicraft/go-advanced-admin"
	// TODO: Add back when gin integrator is compatible with v1.0.1
	// admingin "github.com/ovnicraft/go-advanced-admin-gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Persona represents a user persona with basic information
type Persona struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Name     string `json:"name" gorm:"not null"`
	Email    string `json:"email" gorm:"uniqueIndex;not null"`
	Age      int    `json:"age"`
	IsActive bool   `json:"is_active" gorm:"default:true"`
}

func main() {
	// Initialize Gin router
	router := gin.Default()

	// Initialize database
	db, err := gorm.Open(sqlite.Open("personas.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&Persona{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Seed some sample data
	seedData(db)

	// Create ORM integrator (using GORM - you'll need to implement or use an existing one)
	// ormIntegrator := gorm.NewGORMIntegrator(db) // This would need to be implemented

	// Create web integrator (TODO: Add back when gin integrator is compatible)
	// webIntegrator := admingin.NewGinIntegrator(router.Group("/admin"))

	// Permission function - for demo, allow all actions
	// permissionFunc := func(req admin.PermissionRequest, ctx interface{}) (bool, error) {
	//	return true, nil // Allow all actions for demo purposes
	// }

	// Note: admin.Config is the correct type, not PanelConfig

	// Create admin panel (commented out until compatible ORM integrator is available)

	/*
		panel, err := admin.NewPanel(ormIntegrator, webIntegrator, permissionFunc, nil)
		if err != nil {
			log.Fatal("Failed to create admin panel:", err)
		}

		// Register the Persona model
		app, err := panel.RegisterApp("Personas", "Persona Management", nil)
		if err != nil {
			log.Fatal("Failed to register app:", err)
		}

		_, err = app.RegisterModel(&Persona{}, nil)
		if err != nil {
			log.Fatal("Failed to register model:", err)
		}
	*/

	// Temporary routes for testing
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Gin Admin Panel Example",
			"version": admin.Version,
		})
	})

	router.GET("/personas", func(c *gin.Context) {
		var personas []Persona
		db.Find(&personas)
		c.JSON(http.StatusOK, personas)
	})

	log.Println("Server starting on :8080")
	log.Println("Admin panel will be available at: http://localhost:8080/admin")
	log.Fatal(router.Run(":8080"))
}

func seedData(db *gorm.DB) {
	// Check if we already have data
	var count int64
	db.Model(&Persona{}).Count(&count)
	if count > 0 {
		return
	}

	// Create sample personas
	personas := []Persona{
		{Name: "John Doe", Email: "john@example.com", Age: 30, IsActive: true},
		{Name: "Jane Smith", Email: "jane@example.com", Age: 25, IsActive: true},
		{Name: "Bob Johnson", Email: "bob@example.com", Age: 35, IsActive: false},
		{Name: "Alice Brown", Email: "alice@example.com", Age: 28, IsActive: true},
	}

	for _, persona := range personas {
		db.Create(&persona)
	}

	log.Println("Sample data seeded successfully")
}
