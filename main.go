// go.mod: module resolver ; require github.com/gin-gonic/gin v1.10.0
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-yaaf/yaaf-network-utils/handlers"
)

func main() {
	r := gin.Default()
	r.POST("/dns", handlers.DnsResolveHandler)
	r.POST("/geo", handlers.GeoIPResolveHandler)
	r.POST("/addr", handlers.AddrResolveHandler)
	_ = r.Run(":8080") // PORT from env
}
