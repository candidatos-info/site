package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	b64 "encoding/base64"

	"github.com/candidatos-info/site/db"
	"github.com/candidatos-info/site/email"
	"github.com/candidatos-info/site/exception"
	"github.com/candidatos-info/site/token"
	"github.com/labstack/echo"
)

type tmplt struct {
	templates *template.Template
}

func (t *tmplt) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

var (
	dbClient       *db.DataStoreClient
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
)

type defaultResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// func homePageHandler(c echo.Context) error {
// 	states, err := dbClient.GetStates()
// 	if err != nil {
// 		log.Printf("failed to retrieve states from db, erro %v", err)
// 		return c.String(http.StatusInternalServerError, err.Error())
// 	}
// 	templateData := struct {
// 		States         []string
// 		CandidateTypes []string
// 	}{
// 		states,
// 		candidateRoles,
// 	}
// 	return c.Render(http.StatusOK, "main.html", templateData)
// }

// func profilesPageHandler(c echo.Context) error {
// 	city := c.QueryParam("city")
// 	if city == "" {
// 		return c.String(http.StatusBadRequest, "cidade inválida")
// 	}
// 	state := c.QueryParam("state")
// 	if state == "" {
// 		return c.String(http.StatusBadRequest, "estado inválido")
// 	}
// 	role := c.QueryParam("role")
// 	if role == "" {
// 		return c.String(http.StatusBadRequest, "cargo inválido")
// 	}
// 	year := c.Param("year")
// 	if year == "" {
// 		return c.String(http.StatusBadRequest, "ano inválido")
// 	}
// 	yearAsInt, err := strconv.Atoi(year)
// 	if err != nil {
// 		log.Printf("failed to parse given year [%s] to int, erro %v", year, err)
// 		return c.String(http.StatusInternalServerError, err.Error())
// 	}
// 	candidates, _ := dbClient.FindCandidatesWithParams(state, city, role, yearAsInt)
// 	templateData := struct {
// 		State        string
// 		City         string
// 		Role         string
// 		Candidatures []*descritor.CandidateForDB
// 		Year         int
// 	}{
// 		state,
// 		city,
// 		role,
// 		candidates,
// 		yearAsInt,
// 	}
// 	return c.Render(http.StatusOK, "profiles.html", templateData)
// }

// func candidatePageHandler(c echo.Context) error {
// 	year := c.Param("year")
// 	if year == "" {
// 		return c.String(http.StatusBadRequest, "ano inválido")
// 	}
// 	yearAsInt, err := strconv.Atoi(year)
// 	if err != nil {
// 		log.Printf("failed to parse given year [%s] to int, erro %v", year, err)
// 		return c.String(http.StatusInternalServerError, err.Error())
// 	}
// 	state := c.Param("state")
// 	if state == "" {
// 		return c.String(http.StatusBadRequest, "estado inválido")
// 	}
// 	city := c.Param("city")
// 	if city == "" {
// 		return c.String(http.StatusBadRequest, "cidade inválida")
// 	}
// 	role := c.Param("role")
// 	if role == "" {
// 		return c.String(http.StatusBadRequest, "cargo inválido")
// 	}
// 	sequencialCandidate := c.Param("sequencialCandidate")
// 	if sequencialCandidate == "" {
// 		return c.String(http.StatusBadRequest, "sequencial de candidato inválido")
// 	}
// 	candidate, err := dbClient.GetCandidateBySequencialID(yearAsInt, state, city, sequencialCandidate)
// 	if err != nil {
// 		log.Printf("failed to retrieve candidates using year [%d], state [%s], city [%s] and sequencial code [%s], erro %v\n", yearAsInt, state, city, sequencialCandidate, err)
// 		return c.String(http.StatusInternalServerError, err.Error())
// 	}
// 	templateData := struct {
// 		State        string
// 		City         string
// 		Role         string
// 		PhotoURL     string
// 		Name         string
// 		Party        string
// 		Twitter      string
// 		Description  string
// 		BallotNumber int
// 	}{
// 		state,
// 		city,
// 		role,
// 		candidate.PhotoURL,
// 		candidate.BallotName,
// 		candidate.Party,
// 		candidate.Twitter,
// 		candidate.Description,
// 		candidate.BallotNumber,
// 	}
// 	return c.Render(http.StatusOK, "candidate.html", templateData)
// }

// func citiesOfState(c echo.Context) error {
// 	state := c.QueryParam("state")
// 	if state == "" {
// 		return c.String(http.StatusBadRequest, "estado inválido")
// 	}
// 	citesOfState, err := dbClient.GetCities(state)
// 	if err != nil {
// 		log.Printf("failed to retrieve cities of state [%s], erro %v", state, err)
// 		return c.String(http.StatusInternalServerError, fmt.Sprintf("erro ao buscar cidades do estado [%s], erro %v", state, err))
// 	}
// 	return c.JSON(http.StatusOK, citesOfState)
// }

// func requestProfileAccess(c echo.Context) error {
// 	request := struct {
// 		Email string
// 	}{}
// 	if err := c.Bind(&request); err != nil {
// 		log.Printf("failed to get request body, erro %v\n", err)
// 		return c.String(http.StatusInternalServerError, fmt.Sprintf("falha ao pegar corpo da requisição, erro %v", err))
// 	}
// 	givenEmail := request.Email
// 	if givenEmail == "" {
// 		return c.String(http.StatusBadRequest, "email inválido")
// 	}
// 	response := struct {
// 		Message string
// 	}{}
// 	foundCandidate, err := dbClient.GetCandidateByEmail(strings.ToUpper(givenEmail), currentYear)
// 	if err != nil {
// 		log.Printf("failed to find candidate by email, error %v", err)
// 		var e *exception.Exception
// 		if errors.As(err, &e) {
// 			response.Message = e.Message
// 			return c.JSON(e.Code, response)
// 		}
// 		return c.String(http.StatusInternalServerError, "erro de processamento")
// 	}
// 	accessToken, err := tokenService.GetToken(givenEmail)
// 	if err != nil {
// 		log.Printf("failed to get acess token, error %v\n", err)
// 		return c.String(http.StatusInternalServerError, "falha ao gerar access token")
// 	}
// 	emailMessage := buildProfileAccessEmail(foundCandidate, accessToken)
// 	if err := emailClient.Send(emailClient.Email, []string{"abuarquemf@gmail.com"}, "Código para acessar candidatos.info", emailMessage); err != nil {
// 		log.Printf("failed to send email to [%s], erro %v\n", givenEmail, err)
// 		return fmt.Errorf("Falha ao enviar email ")
// 	}
// 	response.Message = "Verifique seu email"
// 	return c.JSON(http.StatusOK, response)
// }

// func profileHandle(c echo.Context) error {
// 	accessToken := c.QueryParam("access_token")
// 	if accessToken != "" {
// 		return resolveForAccessToken(accessToken, c)
// 	}
// 	return resolveForEmail(c)
// }

// func resolveForEmail(c echo.Context) error {
// 	year := c.QueryParam("year")
// 	if year == "" {
// 		return c.String(http.StatusBadRequest, "ano inválido")
// 	}
// 	email := c.QueryParam("email")
// 	if email == "" {
// 		return c.String(http.StatusBadRequest, "email inválido")
// 	}
// 	candidate, err := dbClient.GetCandidateByEmail(email, currentYear)
// 	if err != nil {
// 		log.Printf("failed to get candidate by email [%s], erro %v\n", email, err)
// 		return c.String(http.StatusInternalServerError, "falha interna de processamento")
// 	}
// 	templateData := struct {
// 		State        string
// 		City         string
// 		Role         string
// 		PhotoURL     string
// 		Name         string
// 		Party        string
// 		Twitter      string
// 		Description  string
// 		BallotNumber int
// 	}{
// 		candidate.State,
// 		candidate.City,
// 		candidate.Role,
// 		candidate.PhotoURL,
// 		candidate.BallotName,
// 		candidate.Party,
// 		candidate.Twitter,
// 		candidate.Description,
// 		candidate.BallotNumber,
// 	}
// 	return c.Render(http.StatusOK, "candidate.html", templateData)
// }

// func resolveForAccessToken(accessToken string, c echo.Context) error {
// 	if !tokenService.IsValid(accessToken) {
// 		return c.String(http.StatusUnauthorized, "código de acesso inváldio")
// 	}
// 	claims, err := token.GetClaims(accessToken)
// 	if err != nil {
// 		log.Printf("failed to extract email from token claims, erro %v\n", err)
// 		return c.String(http.StatusInternalServerError, "falha ao validar token de acesso")
// 	}
// 	email := claims["email"]
// 	foundCandidate, err := dbClient.GetCandidateByEmail(email, currentYear)
// 	if err != nil {
// 		log.Printf("failed to find candidate using email from token claims, erro %v\n", err)
// 		return c.String(http.StatusInternalServerError, "falha ao buscar informações de candidato")
// 	}
// 	templateData := struct {
// 		Name          string
// 		Authorization string
// 		Site          string
// 		Instagram     string
// 		Twitter       string
// 		Facebook      string
// 		Biography     string
// 		Description   string
// 	}{
// 		foundCandidate.Name,
// 		accessToken,
// 		foundCandidate.Site,
// 		foundCandidate.Instagram,
// 		foundCandidate.Twitter,
// 		foundCandidate.Facebook,
// 		foundCandidate.Biography,
// 		foundCandidate.Description,
// 	}
// 	return c.Render(http.StatusOK, "profile.html", templateData)
// }

// func handleProfileUpdate(c echo.Context) error {
// 	request := struct {
// 		Authorization string `json:"authorization"`
// 		Site          string `json:"site"`
// 		Instagram     string `json:"instagram"`
// 		Twitter       string `json:"twitter"`
// 		Facebook      string `json:"facebook"`
// 		Biography     string `json:"biography"`
// 		Description   string `json:"description"`
// 	}{}
// 	if err := c.Bind(&request); err != nil {
// 		log.Printf("failed to bind request body, erro %v\n", err)
// 		return c.String(http.StatusBadRequest, "corpo de requisição inválido")
// 	}
// 	if !tokenService.IsValid(request.Authorization) {
// 		return c.String(http.StatusUnauthorized, "credencial inválida")
// 	}
// 	tokenClaims, err := token.GetClaims(request.Authorization)
// 	if err != nil {
// 		return c.String(http.StatusInternalServerError, "falha ao processar requisição")
// 	}
// 	candidateEmail := tokenClaims["email"]
// 	votingCity, err := dbClient.GetVotingCityByCandidateEmail(candidateEmail)
// 	if err != nil {
// 		log.Printf("failed to get voting city with email [%s], erro %v\n", candidateEmail, err)
// 		return c.String(http.StatusInternalServerError, "falha ao buscar local de votação")
// 	}
// 	for _, candidate := range votingCity.Candidates { // TODO change candidatures from slice to map to make this query O(1)
// 		if candidate.Email == candidateEmail {
// 			if request.Site != "" || request.Site != candidate.Site {
// 				candidate.Site = request.Site
// 			}
// 			if request.Instagram != "" || request.Instagram != candidate.Instagram {
// 				candidate.Instagram = request.Instagram
// 			}
// 			if request.Twitter != "" || request.Twitter != candidate.Twitter {
// 				candidate.Twitter = request.Twitter
// 			}
// 			if request.Facebook != "" || request.Facebook != candidate.Facebook {
// 				candidate.Facebook = request.Facebook
// 			}
// 			if request.Biography != "" || request.Biography != candidate.Biography {
// 				candidate.Biography = request.Biography
// 			}
// 			if request.Description != "" || request.Description != candidate.Description {
// 				candidate.Description = request.Description
// 			}
// 		}
// 	}
// 	response := struct {
// 		Message string `json:"message"`
// 	}{}
// 	if _, err := dbClient.UpdateVotingCity(votingCity); err != nil {
// 		log.Printf("failed to update voting city, erro %v\n", err)
// 		response.Message = "Falha ao atualizar dados"
// 		return c.JSON(http.StatusOK, response)
// 	}
// 	response.Message = "Dados atualizados com sucesso!"
// 	return c.JSON(http.StatusOK, request)
// }

// func handleReports(c echo.Context) error {
// 	request := struct {
// 		Report        string `json:"report"`
// 		Authorization string `json:"authorization"`
// 	}{}
// 	if err := c.Bind(&request); err != nil {
// 		return c.String(http.StatusBadRequest, "corpo de requisição inválida")
// 	}
// 	if !tokenService.IsValid(request.Authorization) {
// 		return c.String(http.StatusUnauthorized, "credenciais inválida")
// 	}
// 	tokenClaims, err := token.GetClaims(request.Authorization)
// 	if err != nil {
// 		return c.String(http.StatusInternalServerError, "falha ao processar requisição")
// 	}
// 	candidateEmail := tokenClaims["email"]
// 	foundCandidate, err := dbClient.GetCandidateByEmail(candidateEmail, currentYear)
// 	if err != nil {
// 		return c.String(http.StatusInternalServerError, "falha ao buscar dados de candidato")
// 	}
// 	emailMessage := buildReportEmail(foundCandidate, request.Report)
// 	if err := emailClient.Send(emailClient.Email, suportEmails, "Nova denúncia do Candidatos.info", emailMessage); err != nil {
// 		log.Printf("failed to send report email to suport list, error %v\n", err)
// 	}
// 	return c.String(http.StatusOK, "Denúnicia enviada com sucesso!")
// }

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

func loginHandler(c echo.Context) error {
	request := struct {
		Email string `json:"email"`
	}{}
	if err := c.Bind(&request); err != nil {
		log.Printf("failed to read request body, error %v", err)
		return c.JSON(http.StatusBadRequest, defaultResponse{Message: "corpo de requisição inválido", Code: http.StatusBadRequest})
	}
	if !emailRegex.MatchString(request.Email) {
		return c.JSON(http.StatusBadRequest, defaultResponse{Message: "Email fornecido é inválido.", Code: http.StatusBadRequest})
	}
	foundCandidate, err := dbClient.GetCandidateByEmail(strings.ToUpper(request.Email), currentYear)
	if err != nil {
		log.Printf("failed to find candidate by email, error %v\n", err)
		var e *exception.Exception
		if errors.As(err, &e) {
			return c.JSON(e.Code, defaultResponse{Message: e.Message, Code: e.Code})
		}
		return c.JSON(http.StatusInternalServerError, defaultResponse{Message: "Erro interno de processamento!", Code: http.StatusInternalServerError})
	}
	accessToken, err := tokenService.GetToken(request.Email)
	if err != nil {
		log.Printf("failed to get acess token, error %v\n", err)
		return c.JSON(http.StatusInternalServerError, defaultResponse{Message: "Falha ao gerar código de acesso ao sisteme. Tente novamente mais tarde.", Code: http.StatusInternalServerError})
	}
	encodedAccessToken := b64.StdEncoding.EncodeToString([]byte(accessToken))
	emailMessage := buildProfileAccessEmail(foundCandidate, encodedAccessToken)
	if err := emailClient.Send(emailClient.Email, []string{"abuarquemf@gmail.com"}, "Código para acessar candidatos.info", emailMessage); err != nil {
		log.Printf("failed to send email to [%s], erro %v\n", request.Email, err)
		return c.JSON(http.StatusInternalServerError, defaultResponse{Message: "Falha ao enviar email com código de acesso. Por favor tente novamente mais tarde.", Code: http.StatusInternalServerError})
	}
	return c.JSON(http.StatusOK, defaultResponse{Message: "Email com código de acesso enviado. Verifique sua caixa de spam caso não encontre.", Code: http.StatusOK})
}

func requestAccessHandler(c echo.Context) error {
	encodedAccessToken := c.QueryParam("access_token")
	if encodedAccessToken == "" {
		return c.JSON(http.StatusBadRequest, defaultResponse{Message: "Código de acesso é inválido", Code: http.StatusBadRequest})
	}
	accessTokenBytes, err := b64.StdEncoding.DecodeString(encodedAccessToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, defaultResponse{Message: "Falha ao processar token de acesso.", Code: http.StatusInternalServerError})
	}
	claims, err := token.GetClaims(string(accessTokenBytes))
	if err != nil {
		log.Printf("failed to extract email from token claims, erro %v\n", err)
		return c.JSON(http.StatusInternalServerError, defaultResponse{Message: "Falha ao processar token de acesso.", Code: http.StatusInternalServerError})
	}
	email := claims["email"]
	foundCandidate, err := dbClient.GetCandidateByEmail(email, currentYear)
	if err != nil {
		log.Printf("failed to find candidate using email from token claims, erro %v\n", err)
		return c.JSON(http.StatusInternalServerError, defaultResponse{Message: "Falha ao buscar informaçōes de candidatos.", Code: http.StatusInternalServerError})
	}
	response := struct {
		Transparence float64 `json:"transparence"`
		Email        string  `json:"email"`
		Name         string  `json:"name"`
		BallotNumber int     `json:"ballot_number"`
		Party        string  `json:"party"`
		Contact      struct {
			Icon string `json:"icon"`
			Link string `json:"link"`
		} `json:"contact"`
		Biography   string   `json:"biography"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}{
		foundCandidate.Transparence,
		strings.ToLower(foundCandidate.Email),
		foundCandidate.BallotName,
		foundCandidate.BallotNumber,
		foundCandidate.Party,
		struct {
			Icon string "json:\"icon\""
			Link string "json:\"link\""
		}{
			func() string {
				if foundCandidate.Contact != nil {
					return foundCandidate.Contact.IconURL
				}
				return ""
			}(),
			func() string {
				if foundCandidate.Contact != nil {
					return foundCandidate.Contact.Link
				}
				return ""
			}(),
		},
		foundCandidate.BallotName,
		foundCandidate.Description,
		foundCandidate.Tags,
	}
	return c.JSON(http.StatusOK, response)
}

func configsHandler(c echo.Context) error {
	response := struct {
		AllowChangeProfile bool `json:"allow_change_profile"`
	}{
		allowedToUpdateProfile,
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
		Transparence float64 `json:"transparence"`
		Email        string  `json:"email"`
		Name         string  `json:"name"`
		BallotNumber int     `json:"ballot_number"`
		Party        string  `json:"party"`
		Contact      struct {
			Icon string `json:"icon"`
			Link string `json:"link"`
		} `json:"contact"`
		Biography   string   `json:"biography"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
		Sex         string   `json:"sex"`
		Role        string   `json:"role"`
		Picture     string   `json:"picture_url"`
		City        string   `json:"city"`
		State       string   `json:"state"`
	}{
		foundCandidate.Transparence,
		strings.ToLower(foundCandidate.Email),
		foundCandidate.BallotName,
		foundCandidate.BallotNumber,
		foundCandidate.Party,
		struct {
			Icon string "json:\"icon\""
			Link string "json:\"link\""
		}{
			func() string {
				if foundCandidate.Contact != nil {
					return foundCandidate.Contact.IconURL
				}
				return ""
			}(),
			func() string {
				if foundCandidate.Contact != nil {
					return foundCandidate.Contact.Link
				}
				return ""
			}(),
		},
		foundCandidate.BallotName,
		foundCandidate.Description,
		foundCandidate.Tags,
		foundCandidate.Gender,
		rolesMap[foundCandidate.Role],
		foundCandidate.PhotoURL,
		foundCandidate.City,
		foundCandidate.State,
	}
	return c.JSON(http.StatusOK, response)
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
	e := echo.New()
	e.Renderer = &tmplt{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
	e.Static("/static", "templates/")
	// e.GET("/", homePageHandler)
	// e.POST("/profiles/:year", profilesPageHandler)
	// e.GET("/candidato/:year/:state/:city/:role/:sequencialCandidate", candidatePageHandler)
	// e.GET("/profile", profileHandle)
	// e.GET("/api/v1/cities", citiesOfState)
	// e.POST("/api/v1/profiles", requestProfileAccess)
	// e.POST("/api/v1/profiles/update", handleProfileUpdate)
	// e.POST("/api/v1/reports", handleReports)
	e.GET("/api/v2/configs", configsHandler)
	e.POST("/api/v2/contact_us", contactHandler)
	e.POST("/api/v2/candidates/login", loginHandler)
	e.GET("/api/v2/candidates", requestAccessHandler)
	e.GET("/api/v2/candidates/:year/:sequencialID", profileHandler)
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("missing PORT environment variable")
	}
	log.Fatal(e.Start(":" + port))
}
