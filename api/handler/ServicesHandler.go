package handler

import (
	"Positiv/api/models"
	"Positiv/internal/db"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ServicesHandler(c *gin.Context) {
	// Для выборки того что будем брать из базы данных
	name := c.Param("name") // CatalogDesign

	tx, err := db.DB.Begin()
	if err != nil {
		log.Println(err, "begin transaction")
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()

	var servicesPage models.Page
	row := tx.QueryRow("SELECT id, title, COALESCE(short_text, ''), COALESCE(text, ''), COALESCE(preview, ''), name, parent, active FROM pages WHERE name = $1", name)
	err = row.Scan(&servicesPage.ID, &servicesPage.Title, &servicesPage.ShortText, &servicesPage.Text, &servicesPage.Preview, &servicesPage.Name, &servicesPage.Parent, &servicesPage.Active)
	if err != nil {
		if err == sql.ErrNoRows {
			servicesPage = models.Page{}
		} else {
			log.Println(err, "page")
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	fmt.Println("Page Data:", servicesPage)

	var servicesSeo models.Seo
	row = tx.QueryRow("SELECT COALESCE(title, ''), COALESCE(description, ''), COALESCE(keywords, '') FROM seo WHERE page_id = $1", servicesPage.ID)
	err = row.Scan(&servicesSeo.Title, &servicesSeo.Description, &servicesSeo.Keywords)
	if err != nil {
		if err == sql.ErrNoRows {
			servicesSeo = models.Seo{}
		} else {
			log.Println(err, "seo")
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	fmt.Println("SEO Data:", servicesSeo)

	var servicesImages []models.Image
	rows, err := tx.Query("SELECT COALESCE(title, ''), COALESCE(image, '') FROM image WHERE page_id = $1", servicesPage.ID)
	if err != nil {
		log.Println(err, "images")
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var image models.Image
		err := rows.Scan(&image.Title, &image.Image)
		if err != nil {
			log.Println(err, "image scan")
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		servicesImages = append(servicesImages, image)
	}

	if len(servicesImages) == 0 {
		fmt.Println("No images found")
	}

	var servicesPrices []models.Price
	rows, err = tx.Query("SELECT id, COALESCE(title, ''), COALESCE(price, 0), COALESCE(deadline, '') FROM price WHERE page_id = $1", servicesPage.ID)
	if err != nil {
		log.Println(err, "prices")
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var price models.Price
		err := rows.Scan(&price.ID, &price.Title, &price.Price, &price.Deadline)
		if err != nil {
			log.Println(err, "price scan")
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		servicesPrices = append(servicesPrices, price)
	}

	cards, err := getRandomCards(tx, 8)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Передаем данные в шаблон HTML
	data := map[string]interface{}{
		"page":   servicesPage,
		"seo":    servicesSeo,
		"images": servicesImages,
		"prices": servicesPrices,
		"cards":  cards,
	}

	fmt.Println("Images Data:", servicesImages)

	c.HTML(http.StatusOK, "page.html", data)
}
