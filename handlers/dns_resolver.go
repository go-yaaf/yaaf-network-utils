package handlers

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-yaaf/yaaf-network-utils/model"
)

func DnsResolveHandler(c *gin.Context) {
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
		replies[i] = lookupPTR(ip)
	}
	c.JSON(http.StatusOK, model.NewBigQueryResponse(replies))
}

func lookupPTR(ip string) string {
	if ip == "" {
		return ""
	}
	r := net.Resolver{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	names, err := r.LookupAddr(ctx, ip)
	if err != nil || len(names) == 0 {
		return ""
	}
	return strings.Join(names, ", ")
}
