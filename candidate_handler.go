package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/candidatos-info/site/db"
	"github.com/candidatos-info/site/exception"
	"github.com/labstack/echo"
)

func newCandidateHandler(db *db.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: Create error page.
		id := c.Param("id")
		year, err := strconv.Atoi(c.Param("year"))
		if err != nil {
			log.Printf("Parâmetro year inválido (%s):%q\n", c.Param("year"), err)
			return echo.ErrBadRequest
		}
		candidate, err := db.FindCandidateBySequencialIDAndYear(year, id)
		switch {
		case err != nil && err.(*exception.Exception).Code == exception.NotFound:
			return echo.ErrNotFound
		case err != nil:
			log.Printf("%q", err)
			return echo.ErrInternalServerError
		}
		queryMap := make(map[string]interface{})
		queryMap["city"] = candidate.City
		queryMap["state"] = candidate.State
		var candidateTags []string
		for _, proposal := range candidate.Proposals {
			candidateTags = append(candidateTags, proposal.Topic)
		}
		queryMap["tags"] = candidateTags
		queryMap["role"] = candidate.Role
		var page int
		queryPage := c.QueryParam("p")
		if queryPage == "" {
			page = 1
		} else {
			p, err := strconv.Atoi(queryPage)
			if err != nil {
				log.Printf("failed to parse query page [%s] to int, error %v\n", queryPage, err)
				return echo.ErrInternalServerError
			}
			page = p
		}
		relatedCandidatures, paginationData, err := db.FindRelatedCandidatesWithParams(queryMap, defaultPageSize, page)
		if err != nil {
			log.Printf("failed to find related candidatures, error %v\n", err)
			return echo.ErrInternalServerError
		}
		var relatedCandidatesCards []*candidateCard
		for _, rc := range relatedCandidatures {
			if rc.SequencialCandidate != id {
				var tags []string
				for _, p := range rc.Proposals {
					tags = append(tags, p.Topic)
				}
				relatedCandidatesCards = append(relatedCandidatesCards, &candidateCard{
					Transparency: rc.Transparency,
					Picture:      rc.PhotoURL,
					Name:         rc.BallotName,
					City:         rc.City,
					State:        rc.State,
					Role:         rc.Role,
					Party:        rc.Party,
					Number:       rc.BallotNumber,
					Tags:         tags,
					SequentialID: rc.SequencialCandidate,
					Gender:       rc.Gender,
				})
			}
		}
		loadMoreURL := fmt.Sprintf("%s/c/%d/%s?p=%d", siteURL, year, candidate.SequencialCandidate, paginationData.Next)
		r := c.Render(http.StatusOK, "candidato.html", map[string]interface{}{
			"HasMore":           paginationData.Next != 0,
			"LoadMoreURL":       loadMoreURL,
			"Candidato":         candidate,
			"RelatedCandidates": relatedCandidatesCards,
		})
		fmt.Println(r)
		return r
	}
}
