package handler

import (
	"Positiv/api/models"
	"Positiv/internal/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetContact(c *gin.Context) {
	tx, err := db.DB.Begin()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()

	var contactsPage []models.Contact
	rows, err := tx.Query("SELECT id, key, value FROM contacts")
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var contact models.Contact
		err := rows.Scan(&contact.ID, &contact.Key, &contact.Value)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		contactsPage = append(contactsPage, contact)
	}

	data := map[string]interface{}{
		"contacts": contactsPage,
	}

	c.HTML(http.StatusOK, "contacts.html", data)
}
