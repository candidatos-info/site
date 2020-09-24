package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	b64 "encoding/base64"

	"github.com/candidatos-info/descritor"
	"github.com/candidatos-info/site/db"
	"github.com/candidatos-info/site/email"
	"github.com/candidatos-info/site/exception"
	"github.com/candidatos-info/site/token"
	"github.com/labstack/echo"
)

const (
	maxBiographyTextSize   = 500
	maxDescriptionTextSize = 500
	maxTagsSize            = 4
	instagramLogoURL       = "https://logodownload.org/wp-content/uploads/2017/04/instagram-logo-9.png"
	facebookLogoURL        = "https://logodownload.org/wp-content/uploads/2014/09/facebook-logo-11.png"
	twitterLogoURL         = "https://help.twitter.com/content/dam/help-twitter/brand/logo.png"
	websiteLogoURL         = "https://i.pinimg.com/originals/4e/d3/5b/4ed35b1c1bb4a3ddef205a3bbbe7fc17.jpg"
	whatsAppLogoURL        = "https://i0.wp.com/cantinhodabrantes.com.br/wp-content/uploads/2017/08/whatsapp-logo-PNG-Transparent.png?fit=1000%2C1000&ssl=1"
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
	tags                   = []string{"Urbanismo", "LBTQ+", "Meio ambiente", "Esporte", "Educação", "Ecossocialismo", "Transformação digital", "Cultura", "Economia"}
)

type defaultResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
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
		return c.JSON(http.StatusBadRequest, defaultResponse{Message: "Código de acesso é inválido.", Code: http.StatusBadRequest})
	}
	accessTokenBytes, err := b64.StdEncoding.DecodeString(encodedAccessToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, defaultResponse{Message: "Falha ao processar token de acesso.", Code: http.StatusInternalServerError})
	}
	// if !tokenService.IsValid(string(accessTokenBytes)) {
	// 	return c.JSON(http.StatusUnauthorized, defaultResponse{Message: "Código de acesso inválido.", Code: http.StatusUnauthorized})
	// }
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
		Biography     string   `json:"biography"`
		Description   string   `json:"description"`
		Tags          []string `json:"tags"`
		AvailableTags []string `json:"available_tags"`
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
		tags,
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
		Transparency float64 `json:"transparency"`
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
		foundCandidate.Role,
		foundCandidate.PhotoURL,
		foundCandidate.City,
		foundCandidate.State,
	}
	return c.JSON(http.StatusOK, response)
}

func updateProfileHandler(c echo.Context) error {
	encodedAccessToken := c.QueryParam("access_token")
	if encodedAccessToken == "" {
		return c.JSON(http.StatusBadRequest, defaultResponse{Message: "Código de acesso é inválido.", Code: http.StatusBadRequest})
	}
	accessTokenBytes, err := b64.StdEncoding.DecodeString(encodedAccessToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, defaultResponse{Message: "Falha ao processar token de acesso.", Code: http.StatusInternalServerError})
	}
	// if !tokenService.IsValid(string(accessTokenBytes)) {
	// 	return c.JSON(http.StatusUnauthorized, defaultResponse{Message: "Código de acesso inválido.", Code: http.StatusUnauthorized})
	// }
	claims, err := token.GetClaims(string(accessTokenBytes))
	if err != nil {
		log.Printf("failed to extract email from token claims, erro %v\n", err)
		return c.JSON(http.StatusInternalServerError, defaultResponse{Message: "Falha ao processar token de acesso.", Code: http.StatusInternalServerError})
	}
	email := claims["email"]
	request := struct {
		Conctact struct {
			Link          string `json:"link"`
			SocialNetWork string `json:"social_network"`
		} `json:"contact"`
		Biography   string   `json:"biography"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}{}
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, defaultResponse{Message: "Corpo de requisição inválido", Code: http.StatusBadRequest})
	}
	if len(request.Biography) > maxBiographyTextSize {
		return c.JSON(http.StatusBadRequest, defaultResponse{Message: fmt.Sprintf("Tamanho máximo de descrição é de %d caracteres.", maxBiographyTextSize), Code: http.StatusBadRequest})
	}
	if len(request.Description) > maxDescriptionTextSize {
		return c.JSON(http.StatusBadRequest, defaultResponse{Message: fmt.Sprintf("Tamanho máximo de descrição é de %d caracteres.", maxDescriptionTextSize), Code: http.StatusBadRequest})
	}
	if len(request.Tags) > maxTagsSize {
		return c.JSON(http.StatusBadRequest, defaultResponse{Message: fmt.Sprintf("Número máximo de tags é %d", maxTagsSize), Code: http.StatusBadRequest})
	}
	candidate, err := dbClient.GetCandidateByEmail(email, currentYear)
	if err != nil {
		log.Printf("failed to find candidate using email from token claims, erro %v\n", err)
		return c.JSON(http.StatusInternalServerError, defaultResponse{Message: "Falha ao buscar informaçōes de candidatos.", Code: http.StatusInternalServerError})
	}
	candidate.Biography = request.Biography
	candidate.Description = request.Description
	candidate.Tags = request.Tags
	candidate.Contact = resolveContact(request.Conctact.Link, request.Conctact.SocialNetWork)
	counter := 0.0
	if candidate.Biography != "" {
		counter++
	}
	if candidate.Description != "" {
		counter++
	}
	if len(candidate.Tags) > 0 {
		counter++
	}
	if candidate.Contact != nil {
		counter++
	}
	candidate.Transparence = counter / 4.0
	if _, err := dbClient.UpdateCandidateProfile(candidate); err != nil {
		log.Printf("failed to update candidates profile, erro %v\n", err)
		return c.JSON(http.StatusInternalServerError, defaultResponse{Message: "Falha ao atualizar dados de candidato. Tente novamente mais tarde.", Code: http.StatusInternalServerError})
	}
	return c.JSON(http.StatusOK, defaultResponse{Message: "Seus dados foram atualizados com sucesso!", Code: http.StatusOK})
}

func resolveContact(link, socialNetWork string) *descritor.Contact {
	c := descritor.Contact{
		Link: link,
	}
	switch socialNetWork {
	case "instagram":
		c.IconURL = instagramLogoURL
	case "twitter":
		c.IconURL = twitterLogoURL
	case "facebook":
		c.IconURL = facebookLogoURL
	case "website":
		c.IconURL = websiteLogoURL
	case "phone":
		c.IconURL = whatsAppLogoURL
	}
	return &c
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
	SequencialID string   `json:"sequencial_id"`
	Gender       string   `json:"gender"`
}

func candidatesHandler(c echo.Context) error {
	cacheToken := c.Request().Header.Get("search-cache-token")
	response := struct {
		Candidates []*candidateCard `json:"candidates"`
	}{}
	var candidatesFromDB []*descritor.CandidateForDB
	var err error
	if cacheToken != "" {
		candidatesFromDB, err = getLastCandidatesOfPreviousQuery(cacheToken)
	} else {
		candidatesFromDB, err = getCandidatesByParams(c)
	}
	if err != nil {
		var e *exception.Exception
		if errors.As(err, &e) {
			return c.JSON(e.Code, defaultResponse{Message: e.Message, Code: e.Code})
		}
		return c.JSON(http.StatusInternalServerError, defaultResponse{Message: "Erro interno de processamento!", Code: http.StatusInternalServerError})
	}
	for _, c := range candidatesFromDB {
		response.Candidates = append(response.Candidates, &candidateCard{
			c.Transparence,
			c.PhotoURL,
			c.BallotName,
			c.City,
			c.State,
			c.Role,
			c.Party,
			c.BallotNumber,
			c.Tags,
			c.SequencialCandidate,
			c.Gender,
		})
	}
	// TODO send cache token on headers
	return c.JSON(http.StatusOK, response)
}

func getCandidatesByParams(c echo.Context) ([]*descritor.CandidateForDB, error) {
	state := c.QueryParam("state")
	if state == "" {
		return nil, &exception.Exception{Message: "O estado deve ser fornecido.", Code: exception.InvalidParameters}
	}
	year := c.QueryParam("year")
	if year == "" {
		return nil, &exception.Exception{Message: "O ano deve ser fornecido.", Code: exception.InvalidParameters}
	}
	y, err := strconv.Atoi(year)
	if err != nil {
		log.Printf("failed to parse year [%s] to int, got error %v\n", year, err)
		return nil, &exception.Exception{Message: "Ano fornecido é inválido", Code: exception.ProcessmentError}
	}
	city := c.QueryParam("city")
	gender := c.QueryParam("gender")
	tags := c.QueryParam("tags")
	name := c.QueryParam("name")
	role := c.QueryParam("role")
	t := strings.Split(tags, ",")
	return dbClient.FindCandidatesWithParams(y, state, city, role, gender, t, name)
}

func getLastCandidatesOfPreviousQuery(cacheToken string) ([]*descritor.CandidateForDB, error) {
	return nil, nil
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
	log.Printf("failed to get cities of state [%s], error %v\n", state, err)
	if err != nil {
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

func main() {
	urlConnection := os.Getenv("DB_URL")
	if urlConnection == "" {
		log.Fatal("missing DB_URL environment variable")
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		log.Fatal("missing DN_NAME environment variable")
	}
	c, err := db.NewMongoClient(urlConnection, dbName)
	if err != nil {
		log.Fatalf("failed to connect to database at URL [%s], error %v\n", urlConnection, err)
	}
	log.Println("connected to database")
	dbClient = c
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
	e.GET("/api/v2/states", statesHandler)
	e.GET("/api/v2/cities", citiesHandler)
	e.GET("/api/v2/configs", configsHandler)
	e.POST("/api/v2/contact_us", contactHandler)
	e.POST("/api/v2/candidates/login", loginHandler)
	e.GET("/api/v2/candidates/login", requestAccessHandler)
	e.GET("/api/v2/candidates/:year/:sequencialID", profileHandler)
	e.PUT("/api/v2/candidates", updateProfileHandler)
	e.GET("/api/v2/candidates", candidatesHandler)
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("missing PORT environment variable")
	}
	log.Fatal(e.Start(":" + port))
}
