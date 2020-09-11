package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"text/template"

	"github.com/candidatos-info/descritor"
	"github.com/candidatos-info/site/db"
	"github.com/labstack/echo"
)

var (
	dbClient       *db.DataStoreClient
	candidateRoles = []string{"vereador", "prefeito"} // available candidate roles
)

type tmplt struct {
	templates *template.Template
}

func (t *tmplt) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func homePageHandler(c echo.Context) error {
	states, err := dbClient.GetStates()
	if err != nil {
		log.Printf("failed to retrieve states from db, erro %v", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}
	templateData := struct {
		States         []string
		CandidateTypes []string
	}{
		states,
		candidateRoles,
	}
	return c.Render(http.StatusOK, "main.html", templateData)
}

func profilesPageHandler(c echo.Context) error {
	city := c.QueryParam("city")
	if city == "" {
		return c.String(http.StatusBadRequest, "cidade inválida")
	}
	state := c.QueryParam("state")
	if state == "" {
		return c.String(http.StatusBadRequest, "estado inválido")
	}
	role := c.QueryParam("role")
	if role == "" {
		return c.String(http.StatusBadRequest, "cargo inválido")
	}
	year := c.Param("year")
	if year == "" {
		return c.String(http.StatusBadRequest, "ano inválido")
	}
	yearAsInt, err := strconv.Atoi(year)
	if err != nil {
		log.Printf("failed to parse given year [%s] to int, erro %v", year, err)
		return c.String(http.StatusInternalServerError, err.Error())
	}
	candidates, _ := dbClient.FindCandidatesWithParams(state, city, role, yearAsInt)
	templateData := struct {
		State        string
		City         string
		Role         string
		Candidatures []*descritor.CandidateForDB
		Year         int
	}{
		state,
		city,
		role,
		candidates,
		yearAsInt,
	}
	return c.Render(http.StatusOK, "profiles.html", templateData)
}

func candidatePageHandler(c echo.Context) error {
	year := c.Param("year")
	if year == "" {
		return c.String(http.StatusBadRequest, "ano inválido")
	}
	yearAsInt, err := strconv.Atoi(year)
	if err != nil {
		log.Printf("failed to parse given year [%s] to int, erro %v", year, err)
		return c.String(http.StatusInternalServerError, err.Error())
	}
	state := c.Param("state")
	if state == "" {
		return c.String(http.StatusBadRequest, "estado inválido")
	}
	city := c.Param("city")
	if city == "" {
		return c.String(http.StatusBadRequest, "cidade inválida")
	}
	role := c.Param("role")
	if role == "" {
		return c.String(http.StatusBadRequest, "cargo inválido")
	}
	sequencialCandidate := c.Param("sequencialCandidate")
	if sequencialCandidate == "" {
		return c.String(http.StatusBadRequest, "sequencial de candidato inválido")
	}
	candidate, err := dbClient.GetCandidateBySequencialID(yearAsInt, state, city, sequencialCandidate)
	if err != nil {
		log.Printf("failed to retrieve candidates using year [%d], state [%s], city [%s] and sequencial code [%s], erro %v\n", yearAsInt, state, city, sequencialCandidate, err)
		return c.String(http.StatusInternalServerError, err.Error())
	}
	templateData := struct {
		State        string
		City         string
		Role         string
		PhotoURL     string
		Name         string
		Party        string
		Twitter      string
		Description  string
		BallotNumber int
	}{
		state,
		city,
		role,
		candidate.PhotoURL,
		candidate.BallotName,
		candidate.Party,
		candidate.Twitter,
		candidate.Description,
		candidate.BallotNumber,
	}
	return c.Render(http.StatusOK, "candidate.html", templateData)
}

func citiesOfState(c echo.Context) error {
	state := c.QueryParam("state")
	if state == "" {
		return c.String(http.StatusBadRequest, "estado inválido")
	}
	citesOfState, err := dbClient.GetCities(state)
	if err != nil {
		log.Printf("failed to retrieve cities of state [%s], erro %v", state, err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("erro ao buscar cidades do estado [%s], erro %v", state, err))
	}
	return c.JSON(http.StatusOK, citesOfState)
}

func main() {
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		log.Fatal("missing PROJECT_ID environment variable")
	}
	dbClient = db.NewDataStoreClient(projectID)
	log.Println("connected to database")
	e := echo.New()
	e.Renderer = &tmplt{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
	e.Static("/static", "templates/")
	e.GET("/", homePageHandler)
	e.POST("/profiles/:year", profilesPageHandler)
	e.GET("/candidato/:year/:state/:city/:role/:sequencialCandidate", candidatePageHandler)
	e.GET("/api/v1/cities", citiesOfState) // return the cities of a given state passed as a query param
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("missing PORT environment variable")
	}
	log.Fatal(e.Start(":" + port))
}
