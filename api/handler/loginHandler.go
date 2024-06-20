package handler

import (
	"Positiv/api/models"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func LoginPostHandler(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	var config models.Config
	err := readConfigFile(&config)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Internal server error",
		})
		return
	}

	configHashPass := md5.Sum([]byte(config.AdminPasswordHash))

	// Хеш
	hashedPassword := md5.Sum([]byte(password))

	if configHashPass != hashedPassword {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid password",
		})
		return
	}

	fmt.Println(hashedPassword, configHashPass)

	if hashedPassword != configHashPass || username != config.AdminUsername {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	c.SetCookie("admin_password_hash", config.AdminPasswordHash, 3600, "/", "", false, true)
	c.Redirect(http.StatusMovedPermanently, "/admin/create_page")
	c.Next()
	c.Abort()
}

func LoginHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func readConfigFile(config *models.Config) error {
	file, err := os.Open("config.json")
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(config)
}
