package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/candidatos-info/site/db"
	"github.com/candidatos-info/site/email"
	"github.com/candidatos-info/site/token"
	"github.com/labstack/echo"
)

func newFaleConoscoHandler(year int) echo.HandlerFunc {
	return func(c echo.Context) error {
		encodedAccessToken := c.QueryParam("access_token")
		accessTokenBytes, err := base64.StdEncoding.DecodeString(encodedAccessToken)
		if err != nil {
			log.Printf("error decoding access token %s", string(encodedAccessToken))
			return c.Render(http.StatusOK, "fale-conosco-success.html", map[string]interface{}{
				"ErrorMsg": "Código de acesso inválido",
				"Success":  false,
			})
		}
		if !tokenService.IsValid(string(accessTokenBytes)) {
			log.Printf("invalid access token")
			return c.Render(http.StatusOK, "fale-conosco-success.html", map[string]interface{}{
				"ErrorMsg": "Código de acesso inválido",
				"Success":  false,
			})
		}
		return c.Render(http.StatusOK, "fale-conosco.html", map[string]interface{}{
			"Token": encodedAccessToken,
			"TypeOptions": []struct {
				Label string
				Value string
			}{
				{Label: "Sugestão", Value: "sugestão"},
				{Label: "Reclamação", Value: "reclamação"},
				{Label: "Denúncia", Value: "denúncia"},
				{Label: "Pergunta", Value: "pergunta"},
				{Label: "Requisitar nova Causa/Pauta", Value: "nova-causa"},
			},
		})
	}
}

func newFaleConoscoFormHandler(db *db.Client, tokenService *token.Token, emailClient *email.Client, contactEmail string, year int) echo.HandlerFunc {
	return func(c echo.Context) error {
		encodedAccessToken := c.FormValue("access_token")
		accessTokenBytes, err := base64.StdEncoding.DecodeString(encodedAccessToken)
		if err != nil {
			log.Printf("error decoding access token %s", string(encodedAccessToken))
			return c.Render(http.StatusOK, "fale-conosco-success.html", map[string]interface{}{
				"ErrorMsg": "Código de acesso inválido",
				"Success":  false,
			})
		}
		if !tokenService.IsValid(string(accessTokenBytes)) {
			log.Printf("invalid access token")
			return c.Render(http.StatusOK, "fale-conosco-success.html", map[string]interface{}{
				"ErrorMsg": "Código de acesso inválido",
				"Success":  false,
			})
		}
		mType := c.FormValue("tipo")
		subject := c.FormValue("assunto")
		content := c.FormValue("descricao")
		if mType == "" || subject == "" || content == "" {
			return c.Render(http.StatusOK, "fale-conosco-success.html", map[string]interface{}{
				"ErrorMsg": "Tipo, assunto e descrição são campos obrigatórios.",
				"Success":  false,
			})
		}
		claims, err := token.GetClaims(string(accessTokenBytes))
		if err != nil {
			log.Printf("failed to extract email from token claims, erro %v\n", err)
			return c.Render(http.StatusOK, "fale-conosco-success.html", map[string]interface{}{
				"ErrorMsg": "Erro inesperado. Por favor, tente novamente mais tarde.",
				"Success":  false,
			})
		}
		email := claims["email"]
		cand, err := db.GetCandidateByEmail(email, year)
		mSub := fmt.Sprintf("[Fale conosco] %s", mType)
		mContent := fmt.Sprintf(`
Saudações Equipe Técnica do Candidatos.info,

%s

Cordialmente,
%s(%s),%d`, content, cand.BallotName, cand.Name, cand.BallotNumber)
		if err := emailClient.Send(emailClient.Email, []string{contactEmail}, mSub, mContent); err != nil {
			log.Printf("failed to sending email (%s):%q", contactEmail, err)
			return c.Render(http.StatusOK, "fale-conosco-success.html", map[string]interface{}{
				"ErrorMsg": "Erro inesperado. Por favor, tente novamente mais tarde.",
				"Success":  false,
			})
		}
		return c.Render(http.StatusOK, "fale-conosco-success.html", map[string]interface{}{
			"Candidate":    cand,
			"Success":      true,
			"Year":         year,
			"SequentialID": cand.SequencialCandidate,
		})
	}
}
