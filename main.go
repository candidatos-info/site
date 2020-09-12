package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/candidatos-info/descritor"
	"github.com/candidatos-info/site/db"
	"github.com/candidatos-info/site/email"
	"github.com/candidatos-info/site/exception"
	"github.com/candidatos-info/site/token"
	"github.com/labstack/echo"
)

var (
	dbClient       *db.DataStoreClient
	emailClient    *email.Client
	tokenService   *token.Token
	candidateRoles = []string{"vereador", "prefeito"} // available candidate roles
	siteURL        string
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

func requestProfileAccess(c echo.Context) error {
	request := struct {
		Email string
	}{}
	if err := c.Bind(&request); err != nil {
		log.Printf("failed to get request body, erro %v\n", err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("falha ao pegar corpo da requisição, erro %v", err))
	}
	givenEmail := request.Email
	if givenEmail == "" {
		return c.String(http.StatusBadRequest, "email inválido")
	}
	response := struct {
		Message string
	}{}
	foundCandidate, err := dbClient.GetCandidateByEmail(strings.ToUpper(givenEmail))
	if err != nil {
		log.Printf("failed to find candidate by email, error %v", err)
		var e *exception.Exception
		if errors.As(err, &e) {
			response.Message = e.Message
			return c.JSON(e.Code, response)
		}
		return c.String(http.StatusInternalServerError, "erro de processamento")
	}
	accessToken, err := tokenService.GetToken(givenEmail)
	if err != nil {
		log.Printf("failed to get acess token, error %v\n", err)
		return c.String(http.StatusInternalServerError, "falha ao gerar access token")
	}
	emailMessage := buildEmailMessage(foundCandidate, accessToken)
	err = func() error {
		at := descritor.AccessToken{
			Code:     accessToken,
			IssuedAt: time.Now(),
		}
		if _, err := dbClient.SaveAccessToken(&at); err != nil {
			log.Printf("failed to save access token on db, error %v\n", err)
			return fmt.Errorf("Falha ao salvar access token no banco")
		}
		if err := emailClient.Send(emailClient.Email, []string{"abuarquemf@gmail.com"}, "Código para acessar candidatos.info", emailMessage); err != nil {
			log.Printf("failed to send email to [%s], erro %v\n", givenEmail, err)
			return fmt.Errorf("Falha ao enviar email ")
		}
		return nil
	}()
	if err != nil {
		if err := dbClient.DeleteAccessToken(accessToken); err != nil {
			log.Printf("failed to delete access token, erro %v\n", err)
			return c.String(http.StatusInternalServerError, "erro interno")
		}
		return c.String(http.StatusInternalServerError, err.Error())
	}
	response.Message = "Verifique seu email"
	return c.JSON(http.StatusOK, response)
}

func buildEmailMessage(candidate *descritor.CandidateForDB, accessToken string) string {
	var emailBodyBuilder strings.Builder
	emailBodyBuilder.WriteString(fmt.Sprintf("Olá, %s!\n", candidate.Name))
	emailBodyBuilder.WriteString(fmt.Sprintf("para acessar seu perfil click no link: %s/profile?access_token=%s\n", siteURL, accessToken))
	return emailBodyBuilder.String()
}

func main() {
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		log.Fatal("missing PROJECT_ID environment variable")
	}
	dbClient = db.NewDataStoreClient(projectID)
	log.Println("connected to database")
	emailAccount := os.Getenv("EMAIL")
	if emailAccount == "" {
		log.Fatal("missing EMAIL environment variable")
	}
	password := os.Getenv("PASSWORD")
	if password == "" {
		log.Fatal("missing PASSWORD environment variable")
	}
	siteURL = os.Getenv("SITE_URL")
	if siteURL == "" {
		log.Fatal("missing SITE_URL environment variable")
	}
	emailClient = email.New(emailAccount, password)
	authSecret := os.Getenv("SECRET")
	if authSecret == "" {
		log.Fatal("missing SECRET environment variable")
	}
	tokenService = token.New(authSecret)
	e := echo.New()
	e.Renderer = &tmplt{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
	e.Static("/static", "templates/")
	e.GET("/", homePageHandler)
	e.POST("/profiles/:year", profilesPageHandler)
	e.GET("/candidato/:year/:state/:city/:role/:sequencialCandidate", candidatePageHandler)
	e.GET("/api/v1/cities", citiesOfState) // return the cities of a given state passed as a query param
	e.POST("/api/v1/profiles", requestProfileAccess)
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("missing PORT environment variable")
	}
	log.Fatal(e.Start(":" + port))
}
