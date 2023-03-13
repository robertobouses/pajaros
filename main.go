package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// Estructura para representar a los pájaros
type Pajaro struct {
	ID      int    `json:"id"`
	Nombre  string `json:"nombre"`
	Familia string `json:"familia"`
	Hembra  bool   `json:"hembra"`
}

// Estructura para manejar la conexión con la base de datos
type DB struct {
	*sql.DB
}

// Método para obtener todos los pájaros de la base de datos
func (db *DB) GetAllPajaros() ([]Pajaro, error) {
	rows, err := db.Query("SELECT * FROM pajaros")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pajaros := []Pajaro{}

	for rows.Next() {
		var pajaro Pajaro
		err := rows.Scan(&pajaro.ID, &pajaro.Nombre, &pajaro.Familia, &pajaro.Hembra)
		if err != nil {
			return nil, err
		}
		pajaros = append(pajaros, pajaro)
	}

	return pajaros, nil
}

// Método para obtener todos los pájaros que sean hembra
func (db *DB) GetHembras() ([]Pajaro, error) {
	rows, err := db.Query("SELECT * FROM pajaros WHERE hembra = true")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pajaros := []Pajaro{}

	for rows.Next() {
		var pajaro Pajaro
		err := rows.Scan(&pajaro.ID, &pajaro.Nombre, &pajaro.Familia, &pajaro.Hembra)
		if err != nil {
			return nil, err
		}
		pajaros = append(pajaros, pajaro)
	}

	return pajaros, nil
}

// Método para crear un nuevo pájaro en la base de datos
func (db *DB) CreatePajaro(pajaro *Pajaro) error {
	_, err := db.Exec("INSERT INTO pajaros (nombre, familia, hembra) VALUES ($1, $2, $3)", pajaro.Nombre, pajaro.Familia, pajaro.Hembra)
	if err != nil {
		return err
	}
	return nil
}

// Método para actualizar un pájaro existente en la base de datos
func (db *DB) UpdatePajaro(id int, pajaro *Pajaro) error {
	_, err := db.Exec("UPDATE pajaros SET nombre=$1, familia=$2, hembra=$3 WHERE id=$4", pajaro.Nombre, pajaro.Familia, pajaro.Hembra, id)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	// Conexión a la base de datos
	db, err := sql.Open("postgres", "user=your_user password=your_password dbname=your_db_name sslmode=disable")
	if err != nil {
		return
	}
	defer db.Close()

	// Creamos un objeto DB que encapsula la conexión a la base de datos
	dbObj := &DB{db}

	// Creamos un objeto Gin que manejará las rutas y los métodos HTTP
	r := gin.Default()

	// Ruta para obtener todos los pájaros
	r.GET("/pajaros", func(c *gin.Context) {
		pajaros, err := dbObj.GetAllPajaros()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener los pájaros"})
			return
		}
		c.JSON(http.StatusOK, pajaros)
	})

	// Ruta para obtener los pájaros que sean hembra
	r.GET("/pajaros/hembras", func(c *gin.Context) {
		pajaros, err := dbObj.GetHembras()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener los pájaros"})
			return
		}
		c.JSON(http.StatusOK, pajaros)
	})

	// Ruta para crear un nuevo pájaro
	r.POST("/pajaros", func(c *gin.Context) {
		var pajaro Pajaro
		if err := c.ShouldBindJSON(&pajaro); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := dbObj.CreatePajaro(&pajaro); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear el pájaro"})
			return
		}
		c.JSON(http.StatusCreated, pajaro)
	})

	// Ruta para actualizar un pájaro existente
	r.PUT("/pajaros/:id", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))

		var pajaro Pajaro
		if err := c.ShouldBindJSON(&pajaro); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := dbObj.UpdatePajaro(id, &pajaro); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar el pájaro"})
			return
		}
		c.JSON(http.StatusOK, pajaro)
	})

	// Iniciamos el servidor en el puerto 8080
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
