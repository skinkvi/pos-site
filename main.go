package main

import (
	router "Positiv/api/router"
)

func main() {
	r := router.SetupRouter()

	r.Run(":8080")
}
