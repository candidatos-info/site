package main

import (
	"errors"
	"fmt"
	"github.com/labstack/echo"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type HomeFilters struct {
	State    string
	Year     string
	City     string
	Position string
}

type CandidateTag struct {
	Tag         string
	Description string
}

type Candidate struct {
	Name                   string
	Email                  string
	Party                  string
	NumberOfTerms          int
	CandidatureNumber      string
	Position               string
	City                   string
	ImageURL               string
	TransparencyPercentage int
	Biography              string
	SocialLinks            []SocialLink
	Descriptions           []CandidateTag
}

type SocialLink struct {
	Provider string
	Link     string
}

type TeamMember struct {
	Name        string
	Title       string
	ImageURL    string
	SocialLinks []SocialLink
}

func newSocialLink(provider string, link string) SocialLink {
	return SocialLink{
		Provider: provider,
		Link:     link,
	}
}

func newTeamMember(name string, title string, imageUrl string, socialLinks []SocialLink) TeamMember {
	return TeamMember{
		Name:        name,
		Title:       title,
		ImageURL:    imageUrl,
		SocialLinks: socialLinks,
	}
}

func newHomeFilters(state string, year string, city string, position string) HomeFilters {
	return HomeFilters{
		State:    state,
		Year:     year,
		City:     city,
		Position: position,
	}
}

func newCandidate() *Candidate {
	return &Candidate{
		Name:                   "Fulado de Tal",
		Email:                  "fulano@example.com",
		CandidatureNumber:      "55555",
		Position:               "Vereador",
		Party:                  "PSOL",
		NumberOfTerms:          3,
		City:                   "Maceió - AL",
		ImageURL:               "/img/candidata.png",
		TransparencyPercentage: rand.Intn(100),
		Biography:              "Lorem ipsum dolor sit amet, consectetur adipisicing elit. Adipisci aliquam dignissimos in magnam nihil nostrum optio sint totam unde? Beatae ea illo iusto, laboriosam laudantium libero molestias necessitatibus quos vitae?",
		SocialLinks: []SocialLink{
			newSocialLink("twitter", "#"),
			newSocialLink("instagram", "#"),
			newSocialLink("linkedin", "#"),
		},
		Descriptions: []CandidateTag{
			CandidateTag{Tag: "Urbanismo", Description: "Lorem ipsum dolor sit amet, consectetur adipisicing elit."},
			CandidateTag{Tag: "Veganismo", Description: "Lorem ipsum dolor sit amet, consectetur adipisicing elit. Cum earum iusto nesciunt nobis quaerat quisquam reprehenderit repudiandae temporibus voluptate voluptates. Dolore doloribus expedita, iste laudantium magni nulla pariatur quia totam."},
		},
	}
}

// Define the template registry struct
type TemplateRegistry struct {
	templates map[string]*template.Template
}

// Implement e.Renderer interface
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.templates[name]
	if !ok {
		err := errors.New("template not found -> " + name)
		return err
	}
	return tmpl.ExecuteTemplate(w, "layout.html", data)
}

func getCandidatos(filters HomeFilters, _offset int) []*Candidate {
	if filters.State == "" || filters.City == "" {
		return []*Candidate{}
	}

	return []*Candidate{
		newCandidate(),
		newCandidate(),
		newCandidate(),
		newCandidate(),
		newCandidate(),
		newCandidate(),
		newCandidate(),
		newCandidate(),
		newCandidate(),
		newCandidate(),
	}
}

func getTeamMembers() []TeamMember {
	return []TeamMember{
		newTeamMember("Ana Paula Gomes", "Pitaqueira", "/img/team/ana.jpeg", []SocialLink{
			newSocialLink("linkedin", "#"),
		}),
		newTeamMember("Aurélio Buarque", "Desenvolvedor", "/img/team/aurelio.jpeg", []SocialLink{
			newSocialLink("twitter", "#"),
			newSocialLink("github", "#"),
		}),
		newTeamMember("Bruno Morassutti", "Palpiteiro jurídico", "/img/team/bruno.jpeg", []SocialLink{
			newSocialLink("twitter", "#"),
			newSocialLink("linkedin", "#"),
		}),
		newTeamMember("Daniel Fireman", "Coordenador", "/img/team/daniel.jpeg", []SocialLink{
			newSocialLink("twitter", "#"),
			newSocialLink("github", "#"),
			newSocialLink("linkedin", "#"),
			newSocialLink("instagram", "#"),
		}),
		newTeamMember("Eduardo Cuducos", "Palpiteiro", "/img/team/eduardo.jpeg", []SocialLink{
			newSocialLink("twitter", "#"),
			newSocialLink("github", "#"),
		}),
		newTeamMember("Evelyn Gomes", "Articuladora ", "/img/team/evelyn.jpeg", []SocialLink{
			newSocialLink("twitter", "#"),
			newSocialLink("linkedin", "#"),
			newSocialLink("github", "#"),
		}),
		newTeamMember("Laura Cavalcante", "Advogada ", "/img/team/laura.jpeg", []SocialLink{
			newSocialLink("twitter", "#"),
			newSocialLink("linkedin", "#"),
		}),
		newTeamMember("Mariana Souto", "Designer ", "/img/team/mariana.jpeg", []SocialLink{
			newSocialLink("twitter", "#"),
			newSocialLink("instagram", "#"),
			newSocialLink("linkedin", "#"),
		}),
	}
}

func findCandidate(_token string) *Candidate {
	return newCandidate()
}

func homeHandler(c echo.Context) error {
	year := c.QueryParam("ano")
	if year == "" {
		year = strconv.Itoa(time.Now().Year())
	}
	state := c.QueryParam("estado")
	citiesByState := map[string]interface{}{
		"Alagoas": [2]string{"Maceió", "Arapiraca"},
		"Bahia":   [1]string{"Salvador"},
	}
	cities, err := citiesByState[state]
	if !err {
		cities = []string{}
	}

	filters := newHomeFilters(state, year, c.QueryParam("cidade"), c.QueryParam("cargo"))
	offset, err := strconv.Atoi(c.QueryParam("offset"))
	if err != nil {
	     return fmt.Error("failed to parse offset string [%s] to int, error %v", c.QueryParam("offset"), err)
	}
	candidatos := getCandidatos(filters, offset)

	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"AllStates":     [2]string{"Alagoas", "Bahia"},
		"AllPositions":  [2]string{"Vereador", "Presidente"},
		"CitiesOfState": cities,
		"Filters":       filters,
		"Candidates":    candidatos,
		"LoadMoreUrl":   buildLoadMoreUrl(len(candidatos)+offset, filters),
	})
}

func buildLoadMoreUrl(offset int, filters HomeFilters) string {
	query := map[string]string{
		"estado": filters.State,
		"ano":    filters.Year,
		"cidade": filters.City,
		"cargo":  filters.Position,
	}

	var url string
	for key, val := range query {
		url = url + "&" + fmt.Sprintf("%s=%s", key, val)
	}

	return "?" + strings.Trim(url, "&") + "&offset=" + strconv.Itoa(offset)
}

func getAllTags() []string {
	return []string{
		"Veganismo",
		"Urbanismo",
		"Educação",
	}
}

func getRelatedCandidate(_candidate *Candidate) []*Candidate {
	return []*Candidate{
		newCandidate(),
		newCandidate(),
		newCandidate(),
		newCandidate(),
		newCandidate(),
		newCandidate(),
		newCandidate(),
		newCandidate(),
		newCandidate(),
		newCandidate(),
	}
}

func findCandidateById(_id string) *Candidate {
	return newCandidate()
}

func sobreHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "sobre.html", map[string]interface{}{
		"Team": getTeamMembers(),
	})
}

func souCandidatoHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "sou-candidato.html", map[string]interface{}{})
}

func souCandidatoFormHandler(c echo.Context) error {
	email := c.FormValue("email")
	// @TODO: enviar email com token para o candidato
	return c.Render(http.StatusOK, "sou-candidato-success.html", map[string]interface{}{
		"Email": email,
	})
}

func atualizarCandidatoHandler(c echo.Context) error {
	token := c.QueryParam("token")
	// @TODO: validar token
	// @TODO: só mostrar a tela de aceitar-termo caso o candidato ainda não tenha aceitado
	return c.Render(http.StatusOK, "atualizar-candidato.html", map[string]interface{}{
		"Token":     token,
		"AllTags":   getAllTags(),
		"Candidato": findCandidate(token),
	})
}

func atualizarCandidatoFormHandler(c echo.Context) error {
	// @TODO: processar form.
	return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{})
}

func aceitarTermoHandler(c echo.Context) error {
	token := c.QueryParam("token")
	// @TODO: validar token
	// @TODO: verificar se
	return c.Render(http.StatusOK, "aceitar-termo.html", map[string]interface{}{
		"Token":      token,
		"TextoTermo": "Lorem ipsum dolor sit amet, consectetur adipisicing elit. Aliquam aliquid aspernatur at atque distinctio dolores in, iusto labore mollitia optio quia quibusdam quod tempora! Iste neque optio placeat provident quaerat. Lorem ipsum dolor sit amet, consectetur adipisicing elit. Aliquam aliquid aspernatur at atque distinctio dolores in, iusto labore mollitia optio quia quibusdam quod tempora! Iste neque optio placeat provident quaerat. Lorem ipsum dolor sit amet, consectetur adipisicing elit. Aliquam aliquid aspernatur at atque distinctio dolores in, iusto labore mollitia optio quia quibusdam quod tempora! Iste neque optio placeat provident quaerat.",
	})
}

func aceitarTermoFormHandler(c echo.Context) error {
	token := c.FormValue("token")

	// @TODO: validar termo
	// @TODO: marcar termo como aceito
	return c.Redirect(http.StatusSeeOther, "/atualizar-candidato?token="+token)
}

func candidateHandler(c echo.Context) error {
	// @TODO: get from route.
	id := "123"
	candidate := findCandidate(id)

	return c.Render(http.StatusOK, "candidato.html", map[string]interface{}{
		"Candidato":         candidate,
		"RelatedCandidates": getRelatedCandidate(candidate),
	})
}

func main() {
	templates := make(map[string]*template.Template)
	templates["index.html"] = template.Must(template.ParseFiles("templates/index.html", "templates/layout.html"))
	templates["sobre.html"] = template.Must(template.ParseFiles("templates/sobre.html", "templates/layout.html"))
	templates["candidato.html"] = template.Must(template.ParseFiles("templates/candidato.html", "templates/layout.html"))
	templates["sou-candidato.html"] = template.Must(template.ParseFiles("templates/sou-candidato.html", "templates/layout.html"))
	templates["sou-candidato-success.html"] = template.Must(template.ParseFiles("templates/sou-candidato-success.html", "templates/layout.html"))
	templates["aceitar-termo.html"] = template.Must(template.ParseFiles("templates/aceitar-termo.html", "templates/layout.html"))
	templates["atualizar-candidato.html"] = template.Must(template.ParseFiles("templates/atualizar-candidato.html", "templates/layout.html"))
	templates["atualizar-candidato-success.html"] = template.Must(template.ParseFiles("templates/atualizar-candidato-success.html", "templates/layout.html"))

	e := echo.New()
	e.Renderer = &TemplateRegistry{
		templates: templates,
	}
	e.Static("/", "public")
	e.GET("/", homeHandler)
	e.GET("/candidatos/:id", candidateHandler)
	e.GET("/sobre", sobreHandler)
	e.GET("/sou-candidato", souCandidatoHandler)
	e.POST("/sou-candidato", souCandidatoFormHandler)
	e.GET("/atualizar-candidato", atualizarCandidatoHandler)
	e.POST("/atualizar-candidato", atualizarCandidatoFormHandler)
	e.GET("/aceitar-termo", aceitarTermoHandler)
	e.POST("/aceitar-termo", aceitarTermoFormHandler)

	e.Logger.Fatal(e.Start(":1323"))
}
