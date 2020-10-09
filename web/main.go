package main

import (
    "net/http"
    "errors"
    "strconv"
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
        ImageURL: "",
        TransparencyPercentage: 55,
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

func getCandidatos(filters *HomeFilters) *[]Candidate {
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
    }
}

func homeHandler(c echo.Context) error {
    year := c.QueryParam("year")
    if year == "" {
        year = strconv.Itoa(time.Now().Year())
    }
    state := c.QueryParam("state")
    citiesByState := map[string]interface{}{
        "Alagoas": [2]string{"Maceió", "Arapiraca"},
        "Bahia": [1]string{"Salvador"},
    }
    cities, err := citiesByState[state]
    if ! err {
        cities = []string{}
    }

    filters := newHomeFilters(state, year, c.QueryParam("city"), c.QueryParam("position"))
    candidatos := getCandidatos(filters)

    return c.Render(http.StatusOK, "index.html", map[string]interface{}{
        "AllStates": [2]string{"Alagoas", "Bahia"},
        "AllPositions": [2]string{"Vereador", "Presidente"},
        "CitiesOfState": cities,
        "Filters": filters,
        "Candidates": candidatos,
    })
}

func sobreHandler(c echo.Context) error {
    return c.Render(http.StatusOK, "sobre.html", map[string]interface{}{});
}

func main() {
    templates := make(map[string]*template.Template)
    templates["index.html"] = template.Must(template.ParseFiles("templates/index.html", "templates/layout.html"))
    templates["sobre.html"] = template.Must(template.ParseFiles("templates/sobre.html", "templates/layout.html"))

    e := echo.New()
    e.Renderer = &TemplateRegistry{
        templates: templates,
    }

	e.GET("/", homeHandler)
	e.GET("/sobre", sobreHandler)

	e.Logger.Fatal(e.Start(":1323"))
}