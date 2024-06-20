package handler

import (
	"Positiv/api/models"
	"Positiv/internal/db"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func MainHandler(c *gin.Context) {
	// Начинаем транзакцию
	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer tx.Rollback() // Отменяем транзакцию в случае ошибки

	// Получаем главную страницу из базы данных
	mainPage, err := getMainPage(tx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Получаем категории, которые должны отображаться на главной странице
	categories, err := getCategories(tx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Получаем 8 рандомных карточек, которые должны отображаться на главной странице
	pages, err := getRandomCards(tx, 8)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Фиксируем транзакцию
	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Передаем данные в HTML-шаблон
	data := gin.H{
		"mainPage":   mainPage,
		"categories": categories,
		"pages":      pages,
	}
	c.HTML(http.StatusOK, "main.html", data)
}

func getMainPage(tx *sql.Tx) (*models.Page, error) {
	var page models.Page
	row := tx.QueryRow(`
		SELECT id, title, short_text, preview, text, name, parent, active
		FROM pages
		WHERE name = 'MainPage' AND active = true
	`)
	err := row.Scan(
		&page.ID, &page.Title, &page.ShortText, &page.Preview, &page.Text, &page.Name, &page.Parent, &page.Active,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("main page not found")
	} else if err != nil {
		return nil, err
	}
	return &page, nil
}

func getCategories(tx *sql.Tx) ([]models.Page, error) {
	rows, err := tx.Query(`
		SELECT id, title, short_text, preview, text, name, parent, active
		FROM pages
		WHERE parent IS NULL AND name != 'MainPage' AND active = true
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []models.Page
	for rows.Next() {
		var p models.Page
		err := rows.Scan(
			&p.ID, &p.Title, &p.ShortText, &p.Preview, &p.Text, &p.Name, &p.Parent, &p.Active,
		)
		if err != nil {
			return nil, err
		}
		pages = append(pages, p)
	}

	return pages, nil
}

// getRandomCards возвращает список из n рандомных карточек, у которых parent != null
func getRandomCards(tx *sql.Tx, n int) ([]models.Page, error) {
	var pages []models.Page
	rows, err := tx.Query("SELECT id, title, short_text, preview, text, name, parent FROM pages WHERE parent IS NOT NULL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p models.Page
		err := rows.Scan(&p.ID, &p.Title, &p.ShortText, &p.Preview, &p.Text, &p.Name, &p.Parent)
		if err != nil {
			return nil, err
		}
		pages = append(pages, p)
	}

	if len(pages) == 0 {
		return nil, fmt.Errorf("no cards found")
	}

	if len(pages) < n {
		n = len(pages)
	}

	rand.Seed(time.Now().UnixNano())
	var randomCards []models.Page
	for i := 0; i < n; i++ {
		index := rand.Intn(len(pages))
		randomCards = append(randomCards, pages[index])
		pages = append(pages[:index], pages[index+1:]...)
	}

	return randomCards, nil
}
