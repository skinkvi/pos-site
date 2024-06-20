package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"Positiv/api/models"

	"github.com/gin-gonic/gin"
)

func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("admin_password_hash")
		if err != nil || cookie == "" {
			c.Redirect(http.StatusMovedPermanently, "/admin/login")
			c.Abort()
			return
		}
		fmt.Println(cookie)

		var config models.Config
		if err = readConfigFile(&config); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Internal server error",
			})
			return
		}

		if cookie != config.AdminPasswordHash {
			c.Redirect(http.StatusMovedPermanently, "/admin/login")
			c.Abort()
			return
		}

	}
}

func readConfigFile(config *models.Config) error {
	file, err := os.Open("config.json")
	if err != nil {
		return err
	}
	defer file.Close()
	return json.NewDecoder(file).Decode(config)
}
