package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo"
)

type states struct {
	State  string
	Cities []string
}

func homePageHandler(c echo.Context) error {
	tmpl := template.Must(template.ParseFiles("templates/main.html"))
	var html bytes.Buffer
	shortPosts := []states{
		{
			State:  "AL",
			Cities: []string{"Selecione uma cidade", "Maceio", "Capela"},
		},
		{
			State:  "AC",
			Cities: []string{"Selecione uma cidade", "Rio branco", "Rio Negro"},
		},
	}
	states := struct {
		State []states
	}{
		shortPosts,
	}
	err := tmpl.Execute(&html, states)
	if err != nil {
		return c.HTML(http.StatusOK, "<h1>Error</h1>")
	}
	return c.HTML(http.StatusOK, string(html.Bytes()))
}

func main() {
	e := echo.New()
	e.GET("/", homePageHandler)
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("missing PORT environment variable")
	}
	log.Fatal(e.Start(":" + port))
}
