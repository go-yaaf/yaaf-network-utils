package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/go-yaaf/yaaf-common-net/utils"

	"github.com/go-yaaf/yaaf-network-utils/model"
)

func GeoIPResolveHandler(c *gin.Context) {
	req := model.BigQueryRequest{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.NewBigQueryResponse([]string{""}))
		return
	}

	// Build the list
	replies := make([]string, len(req.Calls))
	for i, call := range req.Calls {
		var ip string
		if len(call) > 0 {
			ip = call[0]
		}
		replies[i] = lookupGeoIP(ip)
	}
	c.JSON(http.StatusOK, model.NewBigQueryResponse(replies))
}

func lookupGeoIP(ip string) string {
	if ip == "" {
		return ""
	}
	if geo, err := utils.IPUtils("").GeoLookupWKT(ip); err != nil {
		return ""
	} else {
		return geo
	}
}
