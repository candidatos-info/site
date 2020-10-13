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
		r := c.Render(http.StatusOK, "candidato.html", map[string]interface{}{
			"Candidato":         candidate,
			"RelatedCandidates": []*candidateCard{},
		})
		fmt.Println(r)
		return r
	}
}
