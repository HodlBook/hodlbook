package controller

import (
	"io"

	"github.com/gin-gonic/gin"
)

// SSEPrices godoc
// @Summary Stream live prices
// @Description Server-Sent Events endpoint for real-time price updates
// @Tags prices
// @Produce text/event-stream
// @Success 200 {string} string "SSE stream"
// @Router /api/prices/stream [get]
func SSEPrices(priceCh <-chan []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		c.Stream(func(w io.Writer) bool {
			select {
			case msg, ok := <-priceCh:
				if !ok {
					return false
				}
				c.SSEvent("prices", string(msg))
				c.Writer.Flush()
				return true
			case <-c.Request.Context().Done():
				return false
			}
		})
	}
}
