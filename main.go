package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/candidatos-info/descritor"
	"github.com/candidatos-info/site/db"
	"github.com/candidatos-info/site/email"
	"github.com/candidatos-info/site/exception"
	"github.com/candidatos-info/site/token"
	pagination "github.com/gobeam/mongo-go-pagination"
	"github.com/labstack/echo"
)

const (
	maxBiographyTextSize     = 500
	maxDescriptionTextSize   = 100
	maxProposalsPerCandidate = 5
	maxTagsSize              = 4
	instagramLogoURL         = "https://logodownload.org/wp-content/uploads/2017/04/instagram-logo-9.png"
	facebookLogoURL          = "https://logodownload.org/wp-content/uploads/2014/09/facebook-logo-11.png"
	twitterLogoURL           = "https://help.twitter.com/content/dam/help-twitter/brand/logo.png"
	websiteLogoURL           = "https://i.pinimg.com/originals/4e/d3/5b/4ed35b1c1bb4a3ddef205a3bbbe7fc17.jpg"
	whatsAppLogoURL          = "https://i0.wp.com/cantinhodabrantes.com.br/wp-content/uploads/2017/08/whatsapp-logo-PNG-Transparent.png?fit=1000%2C1000&ssl=1"
	searchCookieExpiration   = 360 //in hours
	searchCacheCookie        = "searchCookie"
	defaultPageSize          = 20
)

var (
	dbClient       *db.Client
	emailClient    *email.Client
	tokenService   *token.Token
	candidateRoles = []string{"vereador", "prefeito"} // available candidate roles
	siteURL        string
	suportEmails   = []string{"abuarquemf@gmail.com"}
	currentYear    int
	emailRegex     = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	rolesMap       = map[string]string{
		"EM":  "prefeito",
		"LM":  "vereador",
		"VEM": "vice-prefeito",
	}
	allowedToUpdateProfile bool
	tags                   = mustLoadTags()
)

type defaultResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// this struct is used olny as DTO on requests
// and responses about contact.
type contact struct {
	SocialNetwork string `json:"social_network,omitempty"`
	Value         string `json:"value,omitempty"`
}

// this struct is used olny as DTO on requests
// and responses about proposal.
type proposal struct {
	Topic       string `json:"topic,omitempty"`
	Description string `json:"description,omitempty"`
}

func contactHandler(c echo.Context) error {
	request := struct {
		Type    string `json:"type"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}{}
	if err := c.Bind(&request); err != nil {
		log.Printf("failed to read request body, error %v", err)
		return c.JSON(http.StatusBadRequest, defaultResponse{Message: "corpo de requisição inválido", Code: http.StatusBadRequest})
	}
	emailMessage := buildContactMessage(request.Type, request.Body)
	if err := emailClient.Send(emailClient.Email, suportEmails, request.Subject, emailMessage); err != nil {
		log.Printf("failed to send contact email to suport list, error %v\n", err)
		return c.JSON(http.StatusInternalServerError, defaultResponse{Message: "Falha ao enviar email para nosso suporte. Tente novamente", Code: http.StatusInternalServerError})
	}
	return c.JSON(http.StatusOK, defaultResponse{Message: "Obrigado pelo contato. Sua mensagem foi enviada com sucesso!", Code: http.StatusOK})
}

func configsHandler(c echo.Context) error {
	response := struct {
		AllowChangeProfile bool     `json:"allow_change_profile"`
		Currentyear        int      `json:"current_year"`
		Tags               []string `json:"tags"`
	}{
		allowedToUpdateProfile,
		currentYear,
		tags,
	}
	return c.JSON(http.StatusOK, response)
}

func profileHandler(c echo.Context) error {
	year := c.Param("year")
	if year == "" {
		return c.JSON(http.StatusBadRequest, defaultResponse{Message: "Ano inválido.", Code: http.StatusBadRequest})
	}
	y, err := strconv.Atoi(year)
	if err != nil {
		return c.JSON(http.StatusBadRequest, defaultResponse{Message: "Ano inválido.", Code: http.StatusBadRequest})
	}
	sequencialID := c.Param("sequencialID")
	if sequencialID == "" {
		return c.JSON(http.StatusBadRequest, defaultResponse{Message: "Sequencial ID inválido.", Code: http.StatusBadRequest})
	}
	foundCandidate, err := dbClient.FindCandidateBySequencialIDAndYear(y, sequencialID)
	if err != nil {
		var e *exception.Exception
		if errors.As(err, &e) {
			return c.JSON(e.Code, defaultResponse{Message: e.Message, Code: e.Code})
		}
		return c.JSON(http.StatusInternalServerError, defaultResponse{Message: "Erro interno de processamento!", Code: http.StatusInternalServerError})
	}
	response := struct {
		Transparency float64     `json:"transparency"`
		Email        string      `json:"email"`
		Name         string      `json:"name"`
		BallotNumber int         `json:"ballot_number"`
		Party        string      `json:"party"`
		Contacts     []*contact  `json:"contacts"`
		Biography    string      `json:"biography"`
		Proposals    []*proposal `json:"proposals"`
		Sex          string      `json:"sex"`
		Role         string      `json:"role"`
		Picture      string      `json:"picture_url"`
		City         string      `json:"city"`
		State        string      `json:"state"`
	}{
		foundCandidate.Transparency,
		strings.ToLower(foundCandidate.Email),
		foundCandidate.BallotName,
		foundCandidate.BallotNumber,
		foundCandidate.Party,
		parseDescritorContactsToDTO(foundCandidate.Contacts),
		foundCandidate.Biography,
		paseDescritorProposalsToDTO(foundCandidate.Proposals),
		foundCandidate.Gender,
		foundCandidate.Role,
		foundCandidate.PhotoURL,
		foundCandidate.City,
		foundCandidate.State,
	}
	return c.JSON(http.StatusOK, response)
}

func parseDescritorContactsToDTO(contacts []*descritor.Contact) []*contact {
	var c []*contact
	for _, dc := range contacts {
		c = append(c, &contact{
			SocialNetwork: dc.SocialNetwork,
			Value:         dc.Value,
		})
	}
	return c
}

func paseDescritorProposalsToDTO(proposals []*descritor.Proposal) []*proposal {
	var p []*proposal
	for _, dp := range proposals {
		p = append(p, &proposal{
			Topic:       dp.Topic,
			Description: dp.Description,
		})
	}
	return p
}

type candidateCard struct {
	Transparency float64  `json:"transparency"`
	Picture      string   `json:"picture_url"`
	Name         string   `json:"name"`
	City         string   `json:"city"`
	State        string   `json:"state"`
	Role         string   `json:"role"`
	Party        string   `json:"party"`
	Number       int      `json:"number"`
	Tags         []string `json:"tags"`
	SequentialID string   `json:"sequential_id"`
	Gender       string   `json:"gender"`
}

func filterCandidates(c echo.Context, dbClient *db.Client) ([]*candidateCard, int, error) {
	candidatesFromDB, pagination, err := getCandidatesByParams(c, dbClient)
	if err != nil {
		return nil, 0, err
	}
	var ret []*candidateCard
	for _, c := range candidatesFromDB {
		var candidateTags []string
		for _, proposal := range c.Proposals {
			candidateTags = append(candidateTags, proposal.Topic)
		}
		ret = append(ret, &candidateCard{
			c.Transparency,
			c.PhotoURL,
			c.BallotName,
			strings.Title(strings.ToLower(c.City)),
			c.State,
			uiRoles[c.Role],
			c.Party,
			c.BallotNumber,
			candidateTags,
			c.SequencialCandidate,
			c.Gender,
		})
	}
	return ret, int(pagination.Page), nil
}

func getCandidatesByParams(c echo.Context, dbClient *db.Client) ([]*descritor.CandidateForDB, *pagination.PaginationData, error) {
	queryMap, err := getQueryFilters(c)
	if err != nil {
		log.Printf("failed to get filters, error %v\n", err)
		return nil, nil, err
	}
	fmt.Println(queryMap)
	pageSize, err := strconv.Atoi(c.QueryParam("page_size"))
	if err != nil {
		pageSize = defaultPageSize
	}
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil {
		page = 1
	}
	candidatures, pagination, err := dbClient.FindCandidatesWithParams(queryMap, pageSize, page)
	return candidatures, pagination, err
}

func getQueryFilters(c echo.Context) (map[string]interface{}, error) {
	// TODO: change query parameters to English.
	year := c.QueryParam("ano")
	state := c.QueryParam("estado")
	city := c.QueryParam("cidade")
	gender := c.QueryParam("genero")
	name := c.QueryParam("nome")
	role := c.QueryParam("cargo")
	tags := c.QueryParam("tags")

	queryMap := make(map[string]interface{})
	if state != "" {
		queryMap["state"] = state
	}
	if city != "" {
		queryMap["city"] = city
	}
	if year != "" {
		y, err := strconv.Atoi(year)
		if err != nil {
			log.Printf("failed to parse year from string [%s] to int, error %v\n", year, err)
			return nil, exception.New(exception.ProcessmentError, "Ano fornecido é inválido.", nil)
		}
		queryMap["year"] = y
	}

	if gender != "" {
		queryMap["gender"] = gender
	}
	if role != "" {
		queryMap["role"] = role
	}
	if tags != "" {
		queryMap["tags"] = strings.Split(tags, ",")
	}
	if name != "" {
		queryMap["name"] = name
	}
	return queryMap, nil
}

func statesHandler(c echo.Context) error {
	states, err := dbClient.GetStates()
	if err != nil {
		log.Printf("failed to get states from db, error %v\n", err)
		return c.JSON(http.StatusInternalServerError, defaultResponse{Message: "Falha ao buscar estados disponíveis", Code: http.StatusInternalServerError})
	}
	response := struct {
		States []string `json:"states"`
	}{
		states,
	}
	return c.JSON(http.StatusOK, response)
}

func citiesHandler(c echo.Context) error {
	state := c.QueryParam("state")
	if state == "" {
		return c.JSON(http.StatusBadRequest, defaultResponse{Message: "Informe do estado", Code: http.StatusBadRequest})
	}
	cities, err := dbClient.GetCities(state)
	if err != nil {
		log.Printf("failed to get cities of state [%s], error %v\n", state, err)
		var e *exception.Exception
		if errors.As(err, &e) {
			return c.JSON(e.Code, defaultResponse{Message: e.Message, Code: e.Code})
		}
		return c.JSON(http.StatusInternalServerError, defaultResponse{Message: "Erro interno de processamento!", Code: http.StatusInternalServerError})
	}
	response := struct {
		Cities []string `json:"cities"`
	}{
		cities,
	}
	return c.JSON(http.StatusOK, response)
}

type templateRegistry struct {
	templates map[string]*template.Template
}

func (t *templateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.templates[name]
	if !ok {
		err := errors.New("template not found -> " + name)
		return err
	}
	return tmpl.ExecuteTemplate(w, "layout.html", data)
}

func main() {
	urlConnection := os.Getenv("DB_URL")
	if urlConnection == "" {
		log.Fatal("missing DB_URL environment variable")
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		log.Fatal("missing DN_NAME environment variable")
	}
	dbClient, err := db.NewMongoClient(urlConnection, dbName)
	if err != nil {
		log.Fatalf("failed to connect to database at URL [%s], error %v\n", urlConnection, err)
	}
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
	ey := os.Getenv("ELECTION_YEAR")
	if ey == "" {
		log.Fatal("missing ELECTION_YEAR environment variable")
	}
	electionYearAsInt, err := strconv.Atoi(ey)
	if err != nil {
		log.Fatalf("failed to parse environment variable ELECTION_YEAR with value [%s] to  int, error %v", ey, err)
	}
	currentYear = electionYearAsInt
	updateProfile := os.Getenv("UPDATE_PROFILE")
	if updateProfile == "" {
		log.Fatal("missing UPDATE_PROFILE environment variable")
	}
	r, err := strconv.Atoi(updateProfile)
	if err != nil {
		log.Fatalf("failed to parte environment variable UPDATE_PROFILE with value [%s] to int, error %v", updateProfile, err)
	}
	allowedToUpdateProfile = r == 1

	templates := make(map[string]*template.Template)
	templates["index.html"] = template.Must(template.ParseFiles("web/templates/index.html", "web/templates/layout.html"))
	templates["sobre.html"] = template.Must(template.ParseFiles("web/templates/sobre.html", "web/templates/layout.html"))
	templates["candidato.html"] = template.Must(template.ParseFiles("web/templates/candidato.html", "web/templates/layout.html"))
	templates["sou-candidato.html"] = template.Must(template.ParseFiles("web/templates/sou-candidato.html", "web/templates/layout.html"))
	templates["sou-candidato-success.html"] = template.Must(template.ParseFiles("web/templates/sou-candidato-success.html", "web/templates/layout.html"))
	templates["aceitar-termo.html"] = template.Must(template.ParseFiles("web/templates/aceitar-termo.html", "web/templates/layout.html"))
	templates["atualizar-candidato.html"] = template.Must(template.ParseFiles("web/templates/atualizar-candidato.html", "web/templates/layout.html"))
	templates["atualizar-candidato-success.html"] = template.Must(template.ParseFiles("web/templates/atualizar-candidato-success.html", "web/templates/layout.html"))

	e := echo.New()
	e.Renderer = &templateRegistry{
		templates: templates,
	}
	e.Static("/", "web/public")

	// Frontend
	e.GET("/", newHomeHandler(dbClient))
	e.GET("/c/:year/:id", newCandidateHandler(dbClient))
	e.GET("/sobre", sobreHandler)
	e.GET("/sou-candidato", souCandidatoGET)
	e.POST("/sou-candidato", newSouCandidatoFormHandler(dbClient, tokenService, emailClient, currentYear))
	e.GET("/atualizar-candidatura", newAtualizarCandidaturaHandler(dbClient, tags, currentYear))
	e.POST("/atualizar-candidatura", newAtualizarCandidaturaFormHandler(dbClient, currentYear))
	e.POST("/aceitar-termo", newAceitarTermoFormHandler(dbClient, currentYear))

	// API endpoints
	// e.GET("/api/v2/states", statesHandler)
	// e.GET("/api/v2/cities", citiesHandler)
	// e.GET("/api/v2/configs", configsHandler)
	// e.POST("/api/v2/contact_us", contactHandler)
	// e.POST("/api/v2/candidates/login", loginHandler)
	// e.GET("/api/v2/candidates/login", requestAccessHandler)
	// e.GET("/api/v2/candidates/:year/:sequencialID", profileHandler)
	// e.PUT("/api/v2/candidates", updateProfileHandler)
	// e.GET("/api/v2/candidates", candidatesHandler)
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("missing PORT environment variable")
	}
	log.Fatal(e.Start(":" + port))
}
