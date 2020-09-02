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

// StatesWithCities is a struct to hold state and its cities
type StatesWithCities struct {
	State  string
	Cities []string
}

func homePageHandler(c echo.Context) error {
	statesWithCities := []StatesWithCities{
		{
			State:  "AL",
			Cities: []string{"Selecione uma cidade", "Maceio", "Capela"},
		},
		{
			State:  "AC",
			Cities: []string{"Selecione uma cidade", "Rio branco", "Rio Negro"},
		},
	}
	states := []string{"ALAGOAS-AL", "ACRE-AC"}
	templateData := struct {
		StateWithCities []StatesWithCities
		States          []string
		CandidateTypes  []string
	}{
		statesWithCities,
		states,
		[]string{"Prefeito", "Verador", "Vice-Prefeito"},
	}
	return c.Render(http.StatusOK, "main.html", templateData)
}

func profilesPageHandler(c echo.Context) error {
	// state := c.FormValue("stateForm")
	// city := c.FormValue("cityForm")
	// role := c.FormValue("rolesForm")
	// fmt.Println(state)
	// fmt.Println(city)
	// fmt.Println(role)
	// return c.String(http.StatusOK, "")
	return c.Render(http.StatusOK, "profiles.html", "")
}

func main() {
	e := echo.New()
	e.Renderer = &tmplt{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
	e.Static("/static", "templates/")
	e.GET("/", homePageHandler)
	e.GET("/profiles", profilesPageHandler)
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("missing PORT environment variable")
	}
	log.Fatal(e.Start(":" + port))
}
