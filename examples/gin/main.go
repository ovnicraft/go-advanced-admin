package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	admingorm "github.com/go-advanced-admin/orm-gorm"
	"github.com/ovnicraft/go-advanced-admin"
	admingin "github.com/ovnicraft/go-advanced-admin-gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ormAdapter adapts the orm-gorm integrator to the latest ORMIntegrator interface.
type ormAdapter struct{ *admingorm.Integrator }

// DeleteByID maps to DeleteInstance for older integrator versions.
func (a ormAdapter) DeleteByID(model interface{}, id interface{}) error {
	return a.DeleteInstance(model, id)
}

// GetAll maps to FetchInstances for older integrator versions.
func (a ormAdapter) GetAll(model interface{}) (interface{}, error) {
	return a.FetchInstances(model)
}

// Persona represents a user persona with basic information
type Persona struct {
    ID       uint   `json:"id" gorm:"primaryKey;column:ID"`
    Name     string `json:"name" gorm:"not null;column:Name"`
    Email    string `json:"email" gorm:"uniqueIndex;not null;column:Email"`
    Age      int    `json:"age" gorm:"column:Age"`
    IsActive bool   `json:"is_active" gorm:"default:true;column:IsActive"`
}

func main() {
	// Initialize Gin router
	router := gin.Default()

    // Initialize database (use fresh DB for this example)
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
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

	// Create web integrator (use root group; admin uses its own prefix)
	webIntegrator := admingin.NewIntegrator(router.Group(""))

	// Permission function - for demo, allow all actions
	permissionFunc := func(req admin.PermissionRequest, ctx interface{}) (bool, error) {
		return true, nil // Allow all actions for demo purposes
	}

	// ORM integrator for GORM
	ormIntegrator := admingorm.NewIntegrator(db)

	// Create admin panel
	// Wrap ORM integrator to satisfy newer interface methods if needed.
	panel, err := admin.NewPanel(ormAdapter{ormIntegrator}, webIntegrator, permissionFunc, nil)
	if err != nil {
		log.Fatal("Failed to create admin panel:", err)
	}

	// Register the Persona model
	app, err := panel.RegisterApp("personas", "Persona Management", nil)
	if err != nil {
		log.Fatal("Failed to register app:", err)
	}

	if _, err := app.RegisterModel(&Persona{}, nil); err != nil {
		log.Fatal("Failed to register model:", err)
	}

	log.Println("Gin integrator initialized:", webIntegrator != nil)
	log.Println("Admin panel initialized:", panel != nil)

	// No example-level asset workaround needed; integrator v0.0.4 serves assets

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

	// Log routes for visibility
	for _, r := range router.Routes() {
		log.Println(r.Method, r.Path)
	}

	log.Println("Server starting on :8080")
	log.Println("Admin panel available at: http://localhost:8080/admin")
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
