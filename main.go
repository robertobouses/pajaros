package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Pajaro struct {
	ID      int    `json:"id"`
	Nombre  string `json:"nombre"`
	Familia string `json:"familia"`
	Hembra  bool   `json:"hembra"`
}

type PajarosDB struct {
	*sql.DB
}

type DB interface {
	InsertPajaro(*Pajaro) error
	GetAllPajaros() ([]Pajaro, error)
	UpdatePajaro(int, *Pajaro) error
	GetHembras() ([]Pajaro, error)
	Close() error
}

func NewDB(dataSourceName string) (DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PajarosDB{db}, nil
}

func (db *PajarosDB) InsertPajaro(p *Pajaro) error {
	sqlStatement := `
		INSERT INTO pajaros (nombre, familia, hembra)
		VALUES ($1, $2, $3)
		RETURNING id`
	err := db.QueryRow(sqlStatement, p.Nombre, p.Familia, p.Hembra).Scan(&p.ID)
	if err != nil {
		return err
	}
	return nil
}

func (db *PajarosDB) GetAllPajaros() ([]Pajaro, error) {
	sqlStatement := `SELECT * FROM pajaros`
	rows, err := db.Query(sqlStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pajaros := []Pajaro{}
	for rows.Next() {
		var p Pajaro
		if err := rows.Scan(&p.ID, &p.Nombre, &p.Familia, &p.Hembra); err != nil {
			return nil, err
		}
		pajaros = append(pajaros, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return pajaros, nil
}

func (db *PajarosDB) UpdatePajaro(id int, p *Pajaro) error {
	sqlStatement := `
		UPDATE pajaros
		SET nombre = $2, familia = $3, hembra = $4
		WHERE id = $1`
	_, err := db.Exec(sqlStatement, id, p.Nombre, p.Familia, p.Hembra)
	if err != nil {
		return err
	}
	return nil
}

func (db *PajarosDB) GetHembras() ([]Pajaro, error) {
	sqlStatement := `SELECT * FROM pajaros WHERE hembra = true`
	rows, err := db.Query(sqlStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pajaros := []Pajaro{}
	for rows.Next() {
		var p Pajaro
		if err := rows.Scan(&p.ID, &p.Nombre, &p.Familia, &p.Hembra); err != nil {
			return nil, err
		}
		pajaros = append(pajaros, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return pajaros, nil
}

func (db *PajarosDB) Close() error {
	return db.DB.Close()
}

func main() {
	db, err := NewDB("postgres://postgres:password@localhost/pajarosdb?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	r := gin.Default()

	r.POST("/pajaros", func(c *gin.Context) {
		var pajaro Pajaro
		if err := c.ShouldBindJSON(&pajaro); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := db.InsertPajaro(&pajaro); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al insertar el pájaro"})
			return
		}
		c.JSON(http.StatusCreated, pajaro)
	})

	r.GET("/pajaros", func(c *gin.Context) {
		pajaros, err := db.GetAllPajaros()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener los pájaros"})
			return
		}
		c.JSON(http.StatusOK, pajaros)
	})

	r.PUT("/pajaros/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "El ID debe ser un número entero"})
			return
		}
		var pajaro Pajaro
		if err := c.ShouldBindJSON(&pajaro); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := db.UpdatePajaro(id, &pajaro); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar el pájaro"})
			return
		}
		c.JSON(http.StatusOK, pajaro)
	})

	r.GET("/pajaros/hembras", func(c *gin.Context) {
		pajaros, err := db.GetHembras()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener los pájaros hembra"})
			return
		}
		c.JSON(http.StatusOK, pajaros)
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
