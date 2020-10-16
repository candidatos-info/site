package main

import (
	"errors"
	"html/template"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/candidatos-info/site/db"
	"github.com/candidatos-info/site/email"
	"github.com/candidatos-info/site/token"
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
	contactEmail := os.Getenv("FALE_CONOSCO_MAIL")
	if contactEmail == "" {
		log.Fatal("missing FALE_CONOSCO_MAIL environment variable")
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
	templates["fale-conosco.html"] = template.Must(template.ParseFiles("web/templates/fale-conosco.html", "web/templates/layout.html"))
	templates["fale-conosco-success.html"] = template.Must(template.ParseFiles("web/templates/fale-conosco-success.html", "web/templates/layout.html"))
	e := echo.New()
	e.Renderer = &templateRegistry{
		templates: templates,
	}
	e.Static("/", "web/public")
	e.GET("/", newHomeHandler(dbClient))
	e.GET("/c/:year/:id", newCandidateHandler(dbClient))
	e.GET("/sobre", sobreHandler)
	e.GET("/sou-candidato", souCandidatoGET)
	e.POST("/sou-candidato", newSouCandidatoFormHandler(dbClient, tokenService, emailClient, currentYear))
	e.GET("/atualizar-candidatura", newAtualizarCandidaturaHandler(dbClient, tags, currentYear))
	e.POST("/atualizar-candidatura", newAtualizarCandidaturaFormHandler(dbClient, currentYear))
	e.POST("/aceitar-termo", newAceitarTermoFormHandler(dbClient, currentYear))
	e.GET("/fale-conosco", newFaleConoscoHandler(currentYear))
	e.POST("/fale-conosco", newFaleConoscoFormHandler(dbClient, tokenService, emailClient, contactEmail, currentYear))

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("missing PORT environment variable")
	}
	log.Fatal(e.Start(":" + port))
}
