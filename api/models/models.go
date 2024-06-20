package models

import "database/sql"

type Page struct {
	ID        int64         `json:"id"`
	Title     string        `json:"title"`
	ShortText string        `json:"short_text"`
	Preview   string        `json:"preview"`
	Text      string        `json:"text"`
	Name      string        `json:"name"`
	Parent    sql.NullInt64 `json:"parent"`
	Active    bool          `json:"active"`
}

type Seo struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Keywords    string `json:"keywords"`
	PageID      int64  `json:"page_id"`
}

type Price struct {
	ID       int     `json:"id"`
	Title    string  `json:"title"`
	Price    float64 `json:"price"`
	Deadline string  `json:"deadline"`
	PageID   int     `json:"page_id"`
}

type Image struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Image  string `json:"image"`
	PageID int    `json:"page_id"`
}

type Contact struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Title string `json:"title"`
}

type Config struct {
	AdminUsername     string `json:"admin_username"`
	AdminPasswordHash string `json:"admin_password"`
}
