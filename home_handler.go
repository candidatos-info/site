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
	fmt.Println("HOME REQUEST")
	return func(c echo.Context) error {
		cities := []string{}
		page := 0

		year := c.QueryParam("ano")
		if year == "" {
			year = strconv.Itoa(time.Now().Year())
		}
		state := c.QueryParam("estado")
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
			homeResultSet, page, err = filterCandidates(c, db)
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
			"AllStates":                uiStates,
			"AllRoles":                 uiRoles,
			"CitiesOfState":            cities,
			"Filters":                  filter,
			"TransparentCandidates":    homeResultSet.transparentCandidatures,
			"TransparentLoadMoreUrl":   buildLoadMoreURL(filter, "/transparent-partial"),
			"Tags":                     tags,
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

func newHomeLoadMoreTransparentCandidates(db *db.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		year := c.QueryParam("ano")
		if year == "" {
			year = strconv.Itoa(time.Now().Year())
		}
		page := c.QueryParam("page")
		var nextPage int
		if page == "" {
			nextPage = 1
		} else {
			p, err := strconv.Atoi(page)
			if err != nil {
				log.Printf("failed to parse page [%s] to int, error %v\n", page, err)
			}
			nextPage = p
		}
		queryMap, err := getQueryFilters(c)
		transparentCandidatures, paginationData, err := db.FindTransparentCandidatures(queryMap, defaultPageSize, nextPage)
		// TODO: substituir por página de erro.
		if err != nil {
			log.Printf("error filtering candidates:%q", err)
			return c.String(http.StatusInternalServerError, "erro filtrando candidatos.")
		}
		var cc []*candidateCard
		for _, c := range transparentCandidatures {
			var candidateTags []string
			for _, proposal := range c.Proposals {
				candidateTags = append(candidateTags, proposal.Topic)
			}
			cc = append(cc, &candidateCard{
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
		filter := &homeFilter{
			State:    c.QueryParam("estado"),
			City:     c.QueryParam("cidade"),
			Year:     year,
			Role:     c.QueryParam("cargo"),
			NextPage: int(paginationData.Next),
			Tag:      c.QueryParam("tag"),
		}
		r := c.Render(http.StatusOK, "index-transparent-load-more.html", map[string]interface{}{
			"TransparentCandidates": []*candidateCard{},
			"Filters":               filter,
			"LoadMoreUrl":           buildLoadMoreURL(filter, "/transparent-partial"),
		})
		if r != nil {
			log.Printf("[newHomeLoadMoreTransparentCandidates] failed to render template, error %v\n", err)
		}
		return r
	}
}

func filterCandidates(c echo.Context, dbClient *db.Client) (*homeResultSet, int, error) {
	rawHomeResultSet, pagination, err := getCandidatesByParams(c, dbClient)
	if err != nil {
		return nil, 0, err
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
	}, int(pagination.Page), nil
}

func getCandidatesByParams(c echo.Context, dbClient *db.Client) (*rawHomeResultSet, *pagination.PaginationData, error) {
	queryMap, err := getQueryFilters(c)
	if err != nil {
		log.Printf("failed to get filters, error %v\n", err)
		return nil, nil, err
	}
	page := 1
	if c.QueryParam("page") != "" {
		var err error
		page, err = strconv.Atoi(c.QueryParam("page"))
		if err != nil {
			log.Printf("failed to get page, error %v\n", err)
		}
	}
	fmt.Println("QUERY MAP TO FILTER ", queryMap)
	transparentCandidatures, pagination, err := dbClient.FindTransparentCandidatures(queryMap, defaultPageSize, page)
	nonTransparentCandidatures, err := dbClient.FindNonTransparentCandidatures(queryMap, nonTransparentMaxCards)
	return &rawHomeResultSet{
		transparentCandidatures:    transparentCandidatures,
		nonTransparentCandidatures: nonTransparentCandidatures,
	}, pagination, err
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
	fmt.Println("[getQueryFilters] queryMap gerado ", queryMap)
	return queryMap, nil
}
