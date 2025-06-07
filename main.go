package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type Contact struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

var db *sql.DB

func initDB() {
	dbURL := os.Getenv("TURSO_DATABASE_URL")
	if dbURL == "" {
		log.Fatal("La variable de entorno TURSO_DATABASE_URL no está configurada")
	}
	dbToken := os.Getenv("TURSO_AUTH_TOKEN")
	if dbToken == "" {
		log.Fatal("La variable de entorno TURSO_AUTH_TOKEN no está configurada")
	}

	dbUrl := fmt.Sprintf("%s?authToken=%s", dbURL, dbToken)

	var err error
	db, err = sql.Open("libsql", dbUrl)
	if err != nil {
		log.Fatal("Error conectando con Turso:", err)
	}

	// Verificar la conexión
	if err := db.Ping(); err != nil {
		log.Fatal("Error verificando conexión con Turso:", err)
	}

	// Crear tabla si no existe
	ctx := context.Background()
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS contacts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			message TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Fatal("Error creando tabla:", err)
	}
}

func main() {
	initDB()
	defer db.Close()

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.POST("/contact", createContact)
	r.GET("/contacts", listContacts)
	r.GET("/health", healthCheck)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Servidor iniciando en puerto %s", port)
	r.Run(":" + port)
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "timestamp": time.Now()})
}

func createContact(c *gin.Context) {
	var contact Contact
	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido: " + err.Error()})
		return
	}

	if contact.Name == "" || contact.Email == "" || contact.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Todos los campos son requeridos"})
		return
	}

	ctx := context.Background()
	result, err := db.ExecContext(ctx, `
		INSERT INTO contacts (name, email, message) VALUES (?, ?, ?)
	`, contact.Name, contact.Email, contact.Message)
	if err != nil {
		log.Printf("Error guardando contacto: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo guardar el contacto"})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error obteniendo ID: %v", err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Contacto guardado exitosamente",
		"id":      id,
	})
}

func listContacts(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	ctx := context.Background()

	var total int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM contacts").Scan(&total)
	if err != nil {
		log.Printf("Error contando contactos: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al contar contactos"})
		return
	}

	rows, err := db.QueryContext(ctx, `
		SELECT id, name, email, message, created_at
		FROM contacts
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		log.Printf("Error consultando contactos: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener los contactos"})
		return
	}
	defer rows.Close()

	var contacts []Contact
	for rows.Next() {
		var ct Contact
		var createdAtStr string
		if err := rows.Scan(&ct.ID, &ct.Name, &ct.Email, &ct.Message, &createdAtStr); err != nil {
			log.Printf("Error escaneando fila: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al leer datos"})
			return
		}

		// Parsear la fecha - intentar múltiples formatos
		ct.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
		if err != nil {
			// Intentar formato ISO 8601
			ct.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
			if err != nil {
				log.Printf("Error parseando fecha %s: %v", createdAtStr, err)
				ct.CreatedAt = time.Now() // Valor por defecto
			}
		}

		contacts = append(contacts, ct)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterando filas: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error procesando datos"})
		return
	}

	// Calcular información de paginación
	totalPages := (total + limit - 1) / limit

	response := gin.H{
		"contacts": contacts,
		"pagination": gin.H{
			"current_page":   page,
			"total_pages":    totalPages,
			"total_items":    total,
			"items_per_page": limit,
		},
	}

	c.JSON(http.StatusOK, response)
}
