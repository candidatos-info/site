package main

import (
    "net/http"
    "strings"
    "fmt"
    "errors"
    "strconv"
    "math/rand"
    "io"
    "time"
    "html/template"
    "github.com/labstack/echo"
)

type HomeFilters struct {
    State string
    Year string
    City string
    Position string
}

type CandidateTag struct {
    Tag string
    Description string
}

type Candidate struct {
    Name string
    CandidatureNumber string
    Position string
    City string
    ImageURL string
    TransparencyPercentage int
    Biography string
    Descriptions *[]CandidateTag
}

type SocialLink struct {
    Provider string
    Link string
}

type TeamMember struct {
    Name string
    Title string
    ImageURL string
    SocialLinks []SocialLink
}

func newSocialLink(provider string, link string) SocialLink {
    return SocialLink{
        Provider: provider,
        Link: link,
    }
}

func newTeamMember(name string, title string, imageUrl string, socialLinks []SocialLink) TeamMember {
    return TeamMember{
        Name: name,
        Title: title,
        ImageURL: imageUrl,
        SocialLinks: socialLinks,
    }
}

func newHomeFilters(state string, year string, city string, position string) *HomeFilters {
    return &HomeFilters{
        State: state,
        Year: year,
        City: city,
        Position: position,
    }
}

func newCandidate() Candidate {
    return Candidate{
        Name: "Fulado de Tal",
        CandidatureNumber: "55555",
        Position: "Vereador",
        City: "Maceió - AL",
        ImageURL: "/img/candidata.png",
        TransparencyPercentage: rand.Intn(100),
        Biography: "Lorem ipsum dolor sit amet, consectetur adipisicing elit. Adipisci aliquam dignissimos in magnam nihil nostrum optio sint totam unde? Beatae ea illo iusto, laboriosam laudantium libero molestias necessitatibus quos vitae?",
        Descriptions: &[]CandidateTag{
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
    err := errors.New("Template not found -> " + name)
    return err
  }
  return tmpl.ExecuteTemplate(w, "layout.html", data)
}

func getCandidatos(filters *HomeFilters, _offset int) *[]Candidate {
    if filters.State == "" || filters.City == "" {
        return &[]Candidate{}
    }

    return &[]Candidate{
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

func homeHandler(c echo.Context) error {
    year := c.QueryParam("ano")
    if year == "" {
        year = strconv.Itoa(time.Now().Year())
    }
    state := c.QueryParam("estado")
    citiesByState := map[string]interface{}{
        "Alagoas": [2]string{"Maceió", "Arapiraca"},
        "Bahia": [1]string{"Salvador"},
    }
    cities, err := citiesByState[state]
    if ! err {
        cities = []string{}
    }

    filters := newHomeFilters(state, year, c.QueryParam("cidade"), c.QueryParam("cargo"))
    offset, _ := strconv.Atoi(c.QueryParam("offset"))
    candidatos := getCandidatos(filters, offset)

    return c.Render(http.StatusOK, "index.html", map[string]interface{}{
        "AllStates": [2]string{"Alagoas", "Bahia"},
        "AllPositions": [2]string{"Vereador", "Presidente"},
        "CitiesOfState": cities,
        "Filters": filters,
        "Candidates": candidatos,
        "LoadMoreUrl": buildLoadMoreUrl(len(*candidatos) + offset, filters),
    })
}

func buildLoadMoreUrl(offset int, filters *HomeFilters) string {
    query := map[string]string{
        "estado": filters.State,
        "ano": filters.Year,
        "cidade": filters.City,
        "cargo": filters.Position,
    }

    var url string
    for key, val := range query {
        url = url + "&" + fmt.Sprintf("%s=%s", key, val)
    }

    return "?" + strings.Trim(url, "&") + "&offset=" + strconv.Itoa(offset)
}

func sobreHandler(c echo.Context) error {
    return c.Render(http.StatusOK, "sobre.html", map[string]interface{}{
        "Team": getTeamMembers(),
    });
}

func main() {
    templates := make(map[string]*template.Template)
    templates["index.html"] = template.Must(template.ParseFiles("templates/index.html", "templates/layout.html"))
    templates["sobre.html"] = template.Must(template.ParseFiles("templates/sobre.html", "templates/layout.html"))

    e := echo.New()
    e.Renderer = &TemplateRegistry{
        templates: templates,
    }
    e.Static("/", "public")
	e.GET("/", homeHandler)
	e.GET("/sobre", sobreHandler)

	e.Logger.Fatal(e.Start(":1323"))
}