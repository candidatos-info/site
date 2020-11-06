package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/candidatos-info/descritor"
	"github.com/candidatos-info/site/db"
	"github.com/candidatos-info/site/exception"
	"github.com/labstack/echo"
)

const (
	nonTransparentMaxCards = 15
	transparentMaxCards    = 20
)

//  in the format they are going to be presented in UI
var (
	uiRoles  = map[string]string{"vereador": "Vereador(a)", "prefeito": "Prefeito(a)", "vice-prefeito": "Vice Prefeito(a)"}
	uiStates = map[string]string{"AL": "Alagoas", "BA": "Bahia", "CE": "Ceará", "MA": "Maranhão", "PB": "Paraíba", "PE": "Pernambuco", "PI": "Piauí", "RN": "Rio Grande do Norte", "SE": "Sergipe"}
)

// struct with the result set from db
type rawHomeResultSet struct {
	transparentCandidatures    []*descritor.CandidateForDB
	nonTransparentCandidatures []*descritor.CandidateForDB
}

// struct which holds candidatures to be show on UI
type homeResultSet struct {
	transparentCandidatures    []*candidateCard
	nonTransparentCandidatures []*candidateCard
}

type homeFilter struct {
	State    string
	Year     string
	City     string
	Role     string
	Tag      []string
	NextPage int
	Name     string
}

func newHomeHandler(db *db.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		cities := []string{}
		page := 0

		year := c.QueryParam("ano")
		if year == "" {
			year = strconv.Itoa(time.Now().Year())
		}
		state := strings.ToUpper(c.QueryParam("estado"))

		// Hack to quickly make subdomains work using cloudflare page redirects.
		if state == "WWW" {
			return c.Redirect(http.StatusPermanentRedirect, "/")
		}

		city := c.QueryParam("cidade")

		// Check cookies and override query parameters when needed.

		// if year == "" || state == "" || city == "" {
		// 	cookie, _ := c.Cookie(searchCacheCookie)
		// 	if cookie != nil {
		// 		cookieValues := strings.Split(cookie.Value, ",")
		// 		if year == "" {
		// 			year = cookieValues[0]
		// 		}
		// 		if state == "" {
		// 			state = cookieValues[1]
		// 		}
		// 		if city == "" && len(cookieValues) > 2 {
		// 			aux, err := base64.StdEncoding.DecodeString(cookieValues[2])
		// 			if err != nil {
		// 				log.Printf("Error decoding city from cookie (%s):%q", cookieValues[2], err)
		// 			} else {
		// 				city = string(aux)
		// 			}
		// 		}
		// 	}
		// }
		homeResultSet := &homeResultSet{}
		if state != "" {
			var err error
			cities, err = db.GetCities(state)
			if err != nil {
				log.Printf("error fetching cities from a state (%s):%q\n", state, err)
				return c.String(http.StatusInternalServerError, "erro buscando cidades.")
			}
			homeResultSet, err = filterCandidates(c, db)
			// TODO: substituir por página de erro.
			if err != nil {
				log.Printf("error filtering candidates:%q", err)
				return c.String(http.StatusInternalServerError, "erro filtrando candidatos.")
			}
		}
		filter := &homeFilter{
			State:    state,
			City:     c.QueryParam("cidade"),
			Year:     year,
			Role:     c.QueryParam("cargo"),
			NextPage: page + 1,
			Tag:      c.Request().URL.Query()["tags"],
			Name:     c.QueryParam("nome"),
		}
		r := c.Render(http.StatusOK, "index.html", map[string]interface{}{
			"AllStates":                uiStates,
			"AllRoles":                 uiRoles,
			"CitiesOfState":            cities,
			"Filters":                  filter,
			"TransparentCandidates":    homeResultSet.transparentCandidatures,
			"Tags":                     tags,
			"TransparentMaxCards":      transparentMaxCards,
			"NonTransparentMaxCards":   nonTransparentMaxCards,
			"NonTransparentCandidates": homeResultSet.nonTransparentCandidatures,
		})
		fmt.Println(r)
		c.SetCookie(&http.Cookie{
			Name:    searchCacheCookie,
			Value:   fmt.Sprintf("%s,%s,%s", year, state, base64.StdEncoding.EncodeToString([]byte(city))),
			Expires: time.Now().Add(time.Hour * searchCookieExpiration),
		})
		return r
	}
}

func filterCandidates(c echo.Context, dbClient *db.Client) (*homeResultSet, error) {
	rawHomeResultSet, err := getCandidatesByParams(c, dbClient)
	if err != nil {
		return nil, err
	}
	var transparentCandidatures []*candidateCard
	for _, c := range rawHomeResultSet.transparentCandidatures {
		var candidateTags []string
		for _, proposal := range c.Proposals {
			candidateTags = append(candidateTags, proposal.Topic)
		}
		transparentCandidatures = append(transparentCandidatures, &candidateCard{
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
	var nonTransparentCandidatures []*candidateCard
	for _, c := range rawHomeResultSet.nonTransparentCandidatures {
		var candidateTags []string
		for _, proposal := range c.Proposals {
			candidateTags = append(candidateTags, proposal.Topic)
		}
		nonTransparentCandidatures = append(nonTransparentCandidatures, &candidateCard{
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
	return &homeResultSet{
		transparentCandidatures:    transparentCandidatures,
		nonTransparentCandidatures: nonTransparentCandidatures,
	}, nil
}

func getCandidatesByParams(c echo.Context, dbClient *db.Client) (*rawHomeResultSet, error) {
	queryMap, err := getQueryFilters(c)
	if err != nil {
		log.Printf("failed to get filters, error %v\n", err)
		return nil, err
	}
	fmt.Println("QUERY MAP TO FILTER ", queryMap)
	transparentCandidatures, err := dbClient.FindTransparentCandidatures(queryMap, transparentMaxCards)
	nonTransparentCandidatures, err := dbClient.FindNonTransparentCandidatures(queryMap, nonTransparentMaxCards)
	return &rawHomeResultSet{
		transparentCandidatures:    transparentCandidatures,
		nonTransparentCandidatures: nonTransparentCandidatures,
	}, err
}

func getQueryFilters(c echo.Context) (map[string]interface{}, error) {
	// TODO: change query parameters to English.
	year := c.QueryParam("ano")
	state := strings.ToUpper(c.QueryParam("estado"))
	city := c.QueryParam("cidade")
	gender := c.QueryParam("genero")
	name := c.QueryParam("nome")
	role := c.QueryParam("cargo")
	tags := c.Request().URL.Query()["tags"]

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
	if len(tags) > 0 {
		queryMap["tags"] = tags
	}
	if name != "" {
		queryMap["name"] = name
	}
	fmt.Println("[getQueryFilters] queryMap gerado ", queryMap)
	return queryMap, nil
}
