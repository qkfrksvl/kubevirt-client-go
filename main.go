package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	r := setupRouter()
	r.Run(":8080")

}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/ping", healthCheck)
	rgKV(r)
	return r
}

func healthCheck(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func rgKV(e *gin.Engine) {
	kv_rg := e.Group("/kv")
	kv_rg.GET("/vms", rgKV_vms)

}

func rgKV_vms(c *gin.Context) {

	vc := setClient()
	c.JSON(http.StatusOK, listVM(vc))

}
