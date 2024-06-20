package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AdminHandlerMain(c *gin.Context) {
	data := map[string]interface{}{
		"pages": getParentPages(),
	}

	c.HTML(http.StatusOK, "admin.html", data)
}
