package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo"
)

type tmplt struct {
	templates *template.Template
}

func (t *tmplt) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func homePageHandler(c echo.Context) error {
	// TODO get states and candidate types from DB
	templateData := struct {
		States         []string
		CandidateTypes []string
	}{
		[]string{"ALAGOAS", "ACRE"},
		[]string{"Prefeito", "Verador", "Vice-Prefeito"},
	}
	return c.Render(http.StatusOK, "main.html", templateData)
}

func profilesPageHandler(c echo.Context) error {
	// TODO show chosen state
	// TODO show chosen city
	// TODO show chosen candidate role
	// TODO render a list of profiles
	return c.Render(http.StatusOK, "profiles.html", "")
}

func citiesOfState(c echo.Context) error {
	// TODO get state from query params using -> state := c.QueryParam("state")
	// TODO query cities of state 'state'
	return c.JSON(http.StatusOK, []string{"Maceio", "Capela", "Atalia", "Penedo"}) // TODO change for the query result
}

func main() {
	e := echo.New()
	e.Renderer = &tmplt{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
	e.Static("/static", "templates/")
	e.GET("/", homePageHandler)
	e.GET("/profiles", profilesPageHandler)
	e.GET("/api/v1/cities", citiesOfState) // return the cities of a given state passed as a query param
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("missing PORT environment variable")
	}
	log.Fatal(e.Start(":" + port))
}
