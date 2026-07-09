// middleware/auth.go
package middleware

import (
"net/http"
"os"

"github.com/gin-gonic/gin"
)

// AdminAuth protects routes that create, modify, or trigger scraping.
// It checks for a matching API key in the "X-API-Key" header.
// Set ADMIN_API_KEY in your .env / environment to enable this.
func AdminAuth() gin.HandlerFunc {
return func(c *gin.Context) {
expected := os.Getenv("ADMIN_API_KEY")

if expected == "" {
c.JSON(http.StatusServiceUnavailable, gin.H{
"error": "Admin routes are disabled: ADMIN_API_KEY is not configured on the server",
})
c.Abort()
return
}

provided := c.GetHeader("X-API-Key")
if provided == "" || provided != expected {
c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing API key"})
c.Abort()
return
}

c.Next()
}
}