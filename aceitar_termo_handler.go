package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/candidatos-info/descritor"
	"github.com/candidatos-info/site/db"
	"github.com/candidatos-info/site/exception"
	"github.com/candidatos-info/site/token"
	"github.com/labstack/echo"
)

func newAceitarTermoFormHandler(dbClient *db.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		encodedAccessToken := c.FormValue("token")
		accessTokenBytes, err := base64.StdEncoding.DecodeString(encodedAccessToken)
		if err != nil {
			log.Printf("error decoding access token %s", string(encodedAccessToken))
			return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": "Erro inesperado. Por favor, tente novamente mais tarde.",
				"Success":  false,
			})
		}
		if !tokenService.IsValid(string(accessTokenBytes)) {
			log.Printf("invalid access token:%s\n", string(accessTokenBytes))
			return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": "Código de acesso inválido",
				"Success":  false,
			})
		}
		claims, err := token.GetClaims(string(accessTokenBytes))
		if err != nil {
			log.Printf("failed to extract claims, error %v\n", err)
			return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": "Erro inesperado. Por favor, tente novamente mais tarde.",
				"Success":  false,
			})
		}
		var foundCandidate *descritor.CandidateForDB
		if s, ok := claims["seqid"]; ok {
			foundCandidate, err = dbClient.FindCandidateBySequencialIDAndYear(globals.Year, s)
			if err != nil {
				log.Printf("Failed find candidate on DB (seqID:%s, year:%d), error %q\n", s, globals.Year, err)
				if err != nil {
					return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
						"ErrorMsg": "Erro inesperado. Por favor, tente novamente mais tarde.",
						"Success":  false,
					})
				}
			}
		}
		if foundCandidate == nil { // fallback on the old behavior.
			email := claims["email"]
			foundCandidate, err = dbClient.GetCandidateByEmail(email, globals.Year)
			if err != nil {
				switch {
				case err != nil && err.(*exception.Exception).Code == exception.NotFound:
					return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
						"ErrorMsg": fmt.Sprintf("Não encontramos um cadastro de candidatura através do email %s. Por favor verifique se o email está correto.", email),
						"Success":  false,
					})
				case err != nil:
					log.Printf("failed find candidate on DB (email:%s), error %v\n", email, err)
					return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
						"ErrorMsg": "Erro inesperado. Por favor, tente novamente mais tarde.",
						"Success":  false,
					})
				}
			}
		}
		loc, err := time.LoadLocation("UTC")
		if err != nil {
			log.Printf("failed load location (UTC), error %v\n", err)
			return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": "Erro inesperado. Por favor, tente novamente mais tarde.",
				"Success":  false,
			})
		}
		foundCandidate.AcceptedTerms = time.Now().In(loc)
		if _, err := dbClient.UpdateCandidateProfile(foundCandidate); err != nil {
			log.Printf("failed to update candidate with time that terms were accepted, error %v", err)
			return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": "Erro inesperado. Por favor, tente novamente mais tarde.",
				"Success":  false,
			})
		}
		return c.Redirect(http.StatusSeeOther, "/atualizar-candidatura?access_token="+encodedAccessToken)
	}
}
