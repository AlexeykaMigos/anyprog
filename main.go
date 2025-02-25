package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()

	//r.GET("/api/products", getProducts)
	//r.POST("/api/products", addProduct)

	err := r.Run(":8080")
	if err != nil {
		return
	}
}
