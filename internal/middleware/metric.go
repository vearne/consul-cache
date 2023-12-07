package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vearne/consul-cache/internal/resource"
	"strconv"
	"strings"
)

// prometheus labels name
const (
	PromLabelPath   = "path"   //http path
	PromLabelStatus = "status" //http response status
)

// Metric metric middleware
func Metric() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		path := parsePath(c.Request.URL.Path)
		resource.HTTPReqTotal.With(prometheus.Labels{
			PromLabelPath:   path,
			PromLabelStatus: strconv.Itoa(c.Writer.Status()),
		}).Inc()
	}
}

func parsePath(path string) string {
	itemList := strings.Split(path, "/")
	if len(itemList) >= 5 {
		return strings.Join(itemList[0:5], "/")
	}
	return path
}
