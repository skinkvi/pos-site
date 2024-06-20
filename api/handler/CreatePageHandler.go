package handler

import (
	"Positiv/api/models"
	"Positiv/internal/db"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func CreatePageHandler(c *gin.Context) {
	data := map[string]interface{}{
		"pages": getParentPages(),
	}
	c.HTML(http.StatusOK, "create_page.html", data)
}

func SavePageHandler(c *gin.Context) {
	title := c.PostForm("title")
	shortText := c.PostForm("short_text")
	text := c.PostForm("text")
	name := c.PostForm("name")

	var parent sql.NullInt64

	parentIDStr := c.PostForm("parent")
	if parentIDStr == "" {
	} else {
		parentID, err := strconv.Atoi(parentIDStr)
		if err != nil {
			return
		}

		if parentID == 0 {
		} else {
			parent.Int64 = int64(parentID)
			parent.Valid = true
		}
	}

	page := models.Page{
		Title:     title,
		ShortText: shortText,
		Text:      text,
		Name:      name,
		Parent:    parent,
	}

	tx, err := db.DB.Begin()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()

	err = tx.QueryRow("INSERT INTO pages (title, short_text, preview, text, name, parent) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id", title, shortText, page.Preview, text, name, parent).Scan(&page.ID)
	if err != nil {
		log.Println(err, "insert page")
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	seoTitle := c.PostForm("seo_title")
	seoDescription := c.PostForm("seo_description")
	seoKeywords := c.PostForm("seo_keywords")

	seo := models.Seo{
		Title:       seoTitle,
		Description: seoDescription,
		Keywords:    seoKeywords,
		PageID:      int64(page.ID),
	}

	form, err := c.MultipartForm()
	if err != nil {
		log.Println(err, "<- это там где MultipartForm()")
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	var previewImages []string

	files := form.File["preview[]"]

	if len(files) == 0 {
		// Если файлы не были выбраны, то просто пропускаем сохранение изображений
		page.Preview = ""
	} else {
		// Сохраняем изображения и устанавливаем preview-изображение для страницы
		var previewImages []string

		for _, file := range files {
			err = c.SaveUploadedFile(file, "./image/"+file.Filename)
			if err != nil {
				log.Println(err, "save uploaded file")
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}

			previewImages = append(previewImages, file.Filename)

			if file == files[0] {
				page.Preview = file.Filename
			} else {
				image := models.Image{
					Title:  file.Filename,
					Image:  file.Filename,
					PageID: int(page.ID),
				}

				fmt.Println(image, "<- это у нас массив с изображениями")

				_, err = tx.Exec("INSERT INTO image (title, image, page_id) VALUES ($1, $2, $3)", image.Title, image.Image, image.PageID)
				if err != nil {
					log.Println(err, "insert image")
					c.AbortWithError(http.StatusInternalServerError, err)
					return
				}
			}
		}

		page.Preview = strings.Join(previewImages, "")
	}

	page.Preview = strings.Join(previewImages, ", ")

	err = tx.QueryRow("INSERT INTO seo (title, description, keywords, page_id) VALUES ($1, $2, $3, $4) RETURNING id", seo.Title, seo.Description, seo.Keywords, seo.PageID).Scan(&seo.ID)
	if err != nil {
		log.Println(err, "insert seo")
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	priceTitles := c.PostFormArray("price_title[]")
	pricePrices := c.PostFormArray("price_price[]")
	priceDeadlines := c.PostFormArray("price_deadline[]")

	for i := 0; i < len(priceTitles); i++ {
		priceTitle := priceTitles[i]
		pricePriceStr := pricePrices[i]
		re := regexp.MustCompile(`^\d+(\.\d+)?$`)
		if !re.MatchString(pricePriceStr) {
			c.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid price price: %s", pricePriceStr))
			return
		}
		pricePrice, err := strconv.ParseFloat(pricePriceStr, 64)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to parse price price: %s", pricePriceStr))
			return
		}

		priceDeadline := priceDeadlines[i]

		price := models.Price{
			Title:    priceTitle,
			Price:    pricePrice,
			Deadline: priceDeadline,
			PageID:   int(page.ID),
		}

		_, err = tx.Exec("INSERT INTO price (title, price, deadline, page_id) VALUES ($1, $2, $3, $4)", price.Title, price.Price, price.Deadline, price.PageID)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

	}

	err = tx.Commit()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Redirect(http.StatusFound, "/admin/services/"+strconv.FormatInt(int64(page.ID), 10))

}

func getParentPages() []models.Page {
	rows, err := db.DB.Query("SELECT id, title, name FROM pages WHERE active = true")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var pages []models.Page

	for rows.Next() {
		var page models.Page
		err := rows.Scan(&page.ID, &page.Title, &page.Name)
		if err != nil {
			log.Fatal(err)
		}
		pages = append(pages, page)
	}

	return pages
}
