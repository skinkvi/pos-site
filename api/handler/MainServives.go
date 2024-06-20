package handler

import (
	"Positiv/api/models"
	"Positiv/internal/db"
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func DesignDevelopmentHandler(c *gin.Context) {
	name := c.Param("name")

	var page models.Page
	var err error
	if err = db.DB.QueryRow("SELECT title, short_text, active, id, name FROM pages WHERE name = $1", name).Scan(&page.Title, &page.ShortText, &page.Active, &page.ID, &page.Name); err != nil {
		log.Println(err, "не удалось получить информацию о странице")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	rows, err := db.DB.Query("SELECT id, title, preview, name FROM pages WHERE parent = $1 AND active = true", page.ID)
	if err != nil {
		log.Println(err, "это где у нас запрос PAGE")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var services []map[string]interface{}

	for rows.Next() {
		var service = make(map[string]interface{})
		var id int
		var title, image, name string
		err := rows.Scan(&id, &title, &image, &name)
		if err != nil {
			log.Println(err, "это там где идет скан некст вся эта штука")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		service["id"] = id
		service["title"] = title
		service["image"] = image
		service["name"] = name

		var minPrice sql.NullFloat64
		err = db.DB.QueryRow("SELECT MIN(price) FROM price WHERE page_id = $1", id).Scan(&minPrice)
		if err != nil {
			log.Println(err, "это там где минимальное число")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if minPrice.Valid {
			service["min_price"] = minPrice.Float64
		} else {
			service["min_price"] = nil
		}

		services = append(services, service)
	}

	onePicture, err := getChildPageImage(int(page.ID))
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	log.Printf("Имя файла изображения: %s", onePicture)

	c.HTML(http.StatusOK, "design_development.html", gin.H{
		"page":       page,
		"services":   services,
		"onePicture": onePicture,
	})
}

func getChildPageImage(pageID int) (string, error) {
	var onePicture string
	err := db.DB.QueryRow("SELECT image FROM image WHERE page_id = (SELECT id FROM pages WHERE parent = $1 LIMIT 1)", pageID).Scan(&onePicture)
	if err != nil {
		return "", err
	}
	return onePicture, nil
}
