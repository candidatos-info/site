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
	pagination "github.com/gobeam/mongo-go-pagination"
	"github.com/labstack/echo"
)

//  in the format they are going to be presented in UI
var (
	uiRoles  = map[string]string{"vereador": "Vereador(a)", "prefeito": "Prefeito(a)", "vice-prefeito": "Vice Prefeito(a)"}
	uiStates = map[string]string{"AL": "Alagoas"}
)

type homeFilter struct {
	State    string
	Year     string
	City     string
	Role     string
	Tag      string
	NextPage int
}

func buildLoadMoreURL(filter *homeFilter, baseURL string) string {
	query := map[string]string{
		"estado": filter.State,
		"ano":    filter.Year,
		"cidade": filter.City,
		"cargo":  filter.Role,
	}

	var url string
	for key, val := range query {
		url = url + "&" + fmt.Sprintf("%s=%s", key, val)
	}

	return baseURL + "?" + strings.Trim(url, "&") + "&page=" + strconv.Itoa(filter.NextPage)
}

func newHomeHandler(db *db.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		candidates := []*candidateCard{}
		cities := []string{}
		page := 0

		year := c.QueryParam("ano")
		if year == "" {
			year = strconv.Itoa(time.Now().Year())
		}
		state := c.QueryParam("estado")
		city := c.QueryParam("cidade")

		// Check cookies and override query parameters when needed.

		if year == "" || state == "" || city == "" {
			cookie, _ := c.Cookie(searchCacheCookie)
			if cookie != nil {
				cookieValues := strings.Split(cookie.Value, ",")
				if year == "" {
					year = cookieValues[0]
				}
				if state == "" {
					state = cookieValues[1]
				}
				if city == "" && len(cookieValues) > 2 {
					aux, err := base64.StdEncoding.DecodeString(cookieValues[2])
					if err != nil {
						log.Printf("Error decoding city from cookie (%s):%q", cookieValues[2], err)
					} else {
						city = string(aux)
					}
				}
			}
		}
		if state != "" {
			var err error
			cities, err = db.GetCities(state)
			if err != nil {
				log.Printf("error fetching cities from a state (%s):%q\n", state, err)
				return c.String(http.StatusInternalServerError, "erro buscando cidades.")
			}
			candidates, page, err = filterCandidates(c, db)
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
			Tag:      c.QueryParam("tag"),
		}
		r := c.Render(http.StatusOK, "index.html", map[string]interface{}{
			"AllStates":     uiStates,
			"AllRoles":      uiRoles,
			"CitiesOfState": cities,
			"Filters":       filter,
			// TODO: aqui precisamos de dois slices: um pra candidaturas transparentes, outro para as demais candidaturas.
			"TransparentCandidates":     candidates,
			"NonTransparentCandidates":  candidates,
			"TransparentLoadMoreUrl":    buildLoadMoreURL(filter, "/transparent-partial"),
			"NonTransparentLoadMoreUrl": buildLoadMoreURL(filter, "/nontransparent-partial"),
			"Tags":                      tags,
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

func newHomeLoadMoreTransparentCandidates(db *db.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		year := c.QueryParam("ano")
		if year == "" {
			year = strconv.Itoa(time.Now().Year())
		}
		// TODO: aqui a gente só carrega candidaturas transparentes.
		candidates, page, err := filterCandidates(c, db)
		// TODO: substituir por página de erro.
		if err != nil {
			log.Printf("error filtering candidates:%q", err)
			return c.String(http.StatusInternalServerError, "erro filtrando candidatos.")
		}

		filter := &homeFilter{
			State:    c.QueryParam("estado"),
			City:     c.QueryParam("cidade"),
			Year:     year,
			Role:     c.QueryParam("cargo"),
			NextPage: page + 1,
			Tag:      c.QueryParam("tag"),
		}

		return c.Render(http.StatusOK, "index-transparent-load-more.html", map[string]interface{}{
			"TransparentCandidates": candidates,
			"Filters":               filter,
			"LoadMoreUrl":           buildLoadMoreURL(filter, "/transparent-partial"),
		})
	}
}

func newHomeLoadMoreNonTransparentCandidates(db *db.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		year := c.QueryParam("ano")
		if year == "" {
			year = strconv.Itoa(time.Now().Year())
		}
		// TODO: aqui a gente só carrega candidaturas não transparentes.
		candidates, page, err := filterCandidates(c, db)
		// TODO: substituir por página de erro.
		if err != nil {
			log.Printf("error filtering candidates:%q", err)
			return c.String(http.StatusInternalServerError, "erro filtrando candidatos.")
		}

		filter := &homeFilter{
			State:    c.QueryParam("estado"),
			City:     c.QueryParam("cidade"),
			Year:     year,
			Role:     c.QueryParam("cargo"),
			NextPage: page + 1,
			Tag:      c.QueryParam("tag"),
		}

		return c.Render(http.StatusOK, "index-nontransparent-load-more.html", map[string]interface{}{
			"NonTransparentCandidates": candidates,
			"LoadMoreUrl":              buildLoadMoreURL(filter, "/nontransparent-partial"),
		})
	}
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
