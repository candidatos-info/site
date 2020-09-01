package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo"
)

// StatesWithCities is a struct to hold state and its cities
type StatesWithCities struct {
	State  string
	Cities []string
}

func homePageHandler(c echo.Context) error {
	tmpl := template.Must(template.ParseFiles("templates/main.html"))
	var html bytes.Buffer
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
	}{
		statesWithCities,
		states,
	}
	err := tmpl.Execute(&html, templateData)
	if err != nil {
		return c.HTML(http.StatusOK, "<h1>Error</h1>")
	}
	return c.HTML(http.StatusOK, string(html.Bytes()))
}

func profilesPageHandler(c echo.Context) error {
	request := c.Request()
	state := request.FormValue("stateForm")
	city := request.FormValue("cityForm")
	role := request.FormValue("rolesForm")
	fmt.Println(state)
	fmt.Println(city)
	fmt.Println(role)
	return c.String(http.StatusOK, "OI")
}

func main() {
	e := echo.New()
	e.GET("/", homePageHandler)
	e.POST("/profiles", profilesPageHandler)
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("missing PORT environment variable")
	}
	log.Fatal(e.Start(":" + port))
}
