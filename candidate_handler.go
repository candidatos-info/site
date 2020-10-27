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

const (
	relatedCandidaturesMaxCards = 15
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
		for _, sn := range candidate.Contacts {
			addrPrefix := ""
			switch sn.SocialNetwork {
			case "email":
				addrPrefix = "mailto:"
			case "telefone":
				addrPrefix = "tel:"
			case "whatsapp":
				addrPrefix = "https://wa.me/"
			case "facebook":
				addrPrefix = "http://facebook.com/"
			case "instagram":
				addrPrefix = "http://instagram.com/"
			case "twitter":
				addrPrefix = "http://twitter.com/"
			case "paginaWeb":
				addrPrefix = "http://"
			}
			sn.Value = addrPrefix + sn.Value
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
		relatedCandidatures, err := db.FindTransparentCandidatures(queryMap, relatedCandidaturesMaxCards)
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
		r := c.Render(http.StatusOK, "candidato.html", map[string]interface{}{
			"Candidato":         candidate,
			"RelatedCandidates": relatedCandidatesCards,
		})
		fmt.Println(r)
		return r
	}
}
