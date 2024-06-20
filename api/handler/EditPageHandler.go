package handler

import (
	"Positiv/api/models"
	"Positiv/internal/db"
	"database/sql"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func EditPageHandler(c *gin.Context) {
	tx, err := db.DB.Begin()
	if err != nil {
		log.Println(err, "begin transaction")
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()

	id, _ := strconv.Atoi(c.Param("id"))

	page, err := getPageByID(tx, id)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	images, err := getImagesByPageID(tx, id)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	prices, err := getPricesByPageID(tx, id)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	seo, err := getSeoByPageID(tx, id)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Получение родительских страниц из базы данных
	var parentPages []models.Page
	rows, err := tx.Query("SELECT id, title FROM pages")
	if err != nil {
		log.Println(err, "query parent pages")
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var p models.Page
		err := rows.Scan(&p.ID, &p.Title)
		if err != nil {
			log.Println(err, "scan parent page")
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		parentPages = append(parentPages, p)
	}

	data := map[string]interface{}{
		"page":        page,
		"images":      images,
		"prices":      prices,
		"seo":         seo,
		"parentPages": parentPages,
	}

	c.HTML(http.StatusOK, "edit_page.html", data)

	err = tx.Commit()
	if err != nil {
		log.Println(err, "commit transaction")
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
}

func UpdatePageHandler(c *gin.Context) {
	tx, err := db.DB.Begin()
	if err != nil {
		log.Println(err, "begin transaction")
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()

	id, _ := strconv.Atoi(c.Param("id"))

	title := c.PostForm("title")
	shortText := c.PostForm("short_text")
	text := c.PostForm("text")
	name := c.PostForm("name")

	parentInt, err := strconv.Atoi(c.PostForm("parent"))
	if err != nil {
		log.Println(err)
	}

	parent := sql.NullInt64{
		Int64: int64(parentInt),
		Valid: true,
	}

	page := models.Page{
		ID:        int64(id),
		Title:     title,
		ShortText: shortText,
		Text:      text,
		Name:      name,
		Parent:    parent,
	}

	_, err = tx.Exec("UPDATE pages SET title = $1, short_text = $2, text = $3, name = $4, parent = $5 WHERE id = $6", title, shortText, text, name, parent, int64(id))

	if err != nil {
		log.Println(err, "update page")
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	const _24K = (1 << 10) * 24
	err = c.Request.ParseMultipartForm(_24K)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var imageTitles []string

	for _, hdr := range c.Request.MultipartForm.File["preview[]"] {

		var infile multipart.File
		c.AbortWithError(http.StatusInternalServerError, err)
		if infile, err = hdr.Open(); nil != err {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		tempFile, err := os.Create("./image/" + hdr.Filename)
		if err != nil {
			return
		}
		defer tempFile.Close()

		_, err = io.Copy(tempFile, infile)

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		imageTitles = append(imageTitles, hdr.Filename)
	}

	page.Preview = strings.Join(imageTitles, ",")

	for i, imageTitle := range imageTitles {
		image := models.Image{
			Title:  imageTitle,
			Image:  imageTitle,
			PageID: int(page.ID),
		}

		_, err = tx.Exec("INSERT INTO image (title, image, page_id) VALUES ($1, $2, $3)", image.Title, image.Image, image.PageID)
		if err != nil {
			log.Println(err, "insert image", i)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	var imageIDStrings []string
	imageIDStrings = append(imageIDStrings, imageTitles...)

	preview := strings.Join(imageIDStrings, ",")

	_, err = tx.Exec("UPDATE pages SET preview = $1 WHERE id = $2", preview, page.ID)
	if err != nil {
		log.Println(err, "update page preview")
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	priceTitles := c.PostFormArray("price_title[]")
	pricePrices := c.PostFormArray("price_price[]")
	priceDeadlines := c.PostFormArray("price_deadline[]")

	// Delete all existing prices for the page
	_, err = tx.Exec("DELETE FROM price WHERE page_id = $1", int64(id))
	if err != nil {
		log.Println(err, "delete prices")
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// Create and save new prices
	for i := 0; i < len(priceTitles); i++ {
		if len(priceTitles[i]) > 0 && len(pricePrices[i]) > 0 && len(priceDeadlines[i]) > 0 {
			pricePriceFloat, err := strconv.ParseFloat(pricePrices[i], 64)
			if err != nil {
				fmt.Println(err, "Это на строке 73 ошибка связанная с ценами)")
				c.AbortWithError(http.StatusBadRequest, err)
				return
			}

			price := models.Price{
				Title:    priceTitles[i],
				Price:    pricePriceFloat,
				Deadline: priceDeadlines[i],
				PageID:   int(page.ID),
			}

			_, err = tx.Exec("INSERT INTO price (title, price, deadline, page_id) VALUES ($1, $2, $3, $4)", price.Title, price.Price, price.Deadline, price.PageID)
			if err != nil {
				log.Println(err, "insert price", i)
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
		} else {
			fmt.Println("Это на строке 86 ошибка связанная с ценами)")
		}
	}

	seoTitle := c.PostForm("seo_title")
	seoDescription := c.PostForm("seo_description")
	seoKeywords := c.PostForm("seo_keywords")

	seoID, _ := strconv.Atoi(c.PostForm("seo_id"))

	seo := models.Seo{
		ID:          int64(seoID),
		Title:       seoTitle,
		Description: seoDescription,
		Keywords:    seoKeywords,
		PageID:      int64(page.ID),
	}

	_, err = tx.Exec("UPDATE seo SET title = $1, description = $2, keywords = $3 WHERE id = $4", seo.Title, seo.Description, seo.Keywords, seo.ID)
	if err != nil {
		log.Println(err, "update seo")
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Println(err, "commit transaction")
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusFound, "/admin/services/"+strconv.FormatInt(int64(page.ID), 10))
}

func getPageByID(tx *sql.Tx, id int) (*models.Page, error) {
	var page models.Page
	err := tx.QueryRow("SELECT id, title, short_text, text, name, parent, preview FROM pages WHERE id = $1", id).Scan(&page.ID, &page.Title, &page.ShortText, &page.Text, &page.Name, &page.Parent, &page.Preview)
	if err != nil {
		return nil, err

	}

	return &page, nil
}

func getImagesByPageID(tx *sql.Tx, id int) ([]models.Image, error) {
	rows, err := tx.Query("SELECT id, title, image FROM image WHERE page_id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []models.Image

	for rows.Next() {
		var image models.Image
		err := rows.Scan(&image.ID, &image.Title, &image.Image)
		if err != nil {
			return nil, err
		}
		images = append(images, image)
	}

	return images, nil
}

func getPricesByPageID(tx *sql.Tx, id int) ([]models.Price, error) {
	rows, err := tx.Query("SELECT id, title, price, deadline FROM price WHERE page_id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []models.Price

	for rows.Next() {
		var price models.Price
		err := rows.Scan(&price.ID, &price.Title, &price.Price, &price.Deadline)
		if err != nil {
			return nil, err
		}
		prices = append(prices, price)
	}

	return prices, nil
}

func getSeoByPageID(tx *sql.Tx, id int) (*models.Seo, error) {
	var seo models.Seo
	err := tx.QueryRow("SELECT id, title, description, keywords FROM seo WHERE page_id = $1", id).Scan(&seo.ID, &seo.Title, &seo.Description, &seo.Keywords)
	if err != nil {
		return nil, err
	}

	return &seo, nil
}
