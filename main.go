package main

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	_ "github.com/all4miller/gingin/docs"
	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/gin-swagger/swaggerFiles"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Sample struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	TimeStamp time.Time `json:"timestamp"`
	V0        *float64  `json:"v0,omitempty"`
	V1        *float64  `json:"v1,omitempty"`
}

type CreateSampleInput struct {
	Name      string    `json:"name" binding:"required"`
	TimeStamp time.Time `json:"timestamp" binding:"required" example:"2021-09-19T10:41:33.333Z"`
	V0        *float64  `json:"v0" binding:"-"`
	V1        *float64  `json:"v1" binding:"-"`
}

func APIRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": "Hello, World!"})
}

type HTTPError struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"status bad request"`
}

func NewError(ctx *gin.Context, status int, err error) {
	er := HTTPError{
		Code:    status,
		Message: err.Error(),
	}
	ctx.JSON(status, er)
}

func CreateSample(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	// Validate input
	var input CreateSampleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		// instead of c.JSON(400, gin.H{"error": err.Error()}) we use func + struct from swaggo
		NewError(c, http.StatusBadRequest, err)
		return
	}

	sample := Sample{Name: input.Name, TimeStamp: input.TimeStamp, V0: input.V0, V1: input.V1}
	db.Create(&sample)
	c.JSON(http.StatusOK, sample)
}

// GetSample retrieves single sample from database by ID
// @Summary Retrieve sample by ID
// @Param id path int true "Sample ID"
// @Success 200 {object} Sample
// @Failure 404 {object} HTTPError
// @Router /samples/{id} [get]
func GetSample(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var sample Sample
	if err := db.Raw("select * from samples where id = ? order by id limit 1", c.Param("id")).Scan(&sample).Error; err != nil {
		NewError(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "hey", "status": http.StatusOK})
}

func ListSamples(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var samples []Sample
	db.Find(&samples)
	c.JSON(http.StatusOK, samples)
}

func main() {
	db, err := gorm.Open(postgres.Open("postgres:///dbwriter_go?host=/var/pgsql_socket"), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		panic(err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(8)
	sqlDB.SetMaxOpenConns(8)
	sqlDB.SetConnMaxLifetime(time.Hour)

	db.AutoMigrate(&Sample{})

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	r.GET("/", APIRoot)
	r.POST("/samples", CreateSample)
	r.GET("/samples/:id", GetSample)
	r.GET("/samples", ListSamples)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	fmt.Printf("NumCPU: %d and GOMAXPROCS: %d\n", runtime.NumCPU(), runtime.GOMAXPROCS(-1))
	fmt.Println("Started up!")

	r.Run("0.0.0.0:8080")
}
