package main

import (
	"encoding/base64"
	"log"
	"net/http"

	"github.com/candidatos-info/site/email"
	"github.com/candidatos-info/site/token"
	"github.com/labstack/echo"
)

func newFaleConoscoHandler(year int) echo.HandlerFunc {
	// TODO remove this struct
	type defaultResponse struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}
	type SelectOption struct {
		Label string
		Value string
	}

	return func(c echo.Context) error {
		encodedAccessToken := c.QueryParam("access_token")
		if encodedAccessToken == "" {
			return c.JSON(http.StatusBadRequest, defaultResponse{Message: "Código de acesso é inválido.", Code: http.StatusBadRequest})
		}

		return c.Render(http.StatusOK, "fale-conosco.html", map[string]interface{}{
			"Token": encodedAccessToken,
			"TypeOptions": []SelectOption{
				SelectOption{Label: "Sugestão", Value: "sugestão"},
				SelectOption{Label: "Reclamação", Value: "reclamação"},
				SelectOption{Label: "Denúncia", Value: "denúncia"},
				SelectOption{Label: "Pergunta", Value: "Pergunta"},
				SelectOption{Label: "Requisitar nova Causa", Value: "nova-causa"},
			},
		})
	}
}

func newFaleConoscoFormHandler(tokenService *token.Token, emailClient *email.Client, contactEmail string, year int) echo.HandlerFunc {
	// TODO remove this struct
	type defaultResponse struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	}
	return func(c echo.Context) error {
		messageType := c.FormValue("type")
		if messageType == "" {
			return c.JSON(http.StatusBadRequest, defaultResponse{Message: "Tipo da mensagem é inválido.", Code: http.StatusBadRequest})
		}
		subject := c.FormValue("assunto")
		if subject == "" {
			return c.JSON(http.StatusBadRequest, defaultResponse{Message: "Assunto da mensagem é inválido.", Code: http.StatusBadRequest})
		}
		description := c.FormValue("descricao")
		if description == "" {
			return c.JSON(http.StatusBadRequest, defaultResponse{Message: "Descrição da mensagem é inválido.", Code: http.StatusBadRequest})
		}
		encodedAccessToken := c.FormValue("access_token")
		if encodedAccessToken == "" {
			return c.JSON(http.StatusBadRequest, defaultResponse{Message: "Código de acesso é inválido.", Code: http.StatusBadRequest})
		}
		accessTokenBytes, err := base64.StdEncoding.DecodeString(encodedAccessToken)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, defaultResponse{Message: "Falha ao processar token de acesso.", Code: http.StatusInternalServerError})
		}
		if !tokenService.IsValid(string(accessTokenBytes)) {
			return c.JSON(http.StatusUnauthorized, defaultResponse{Message: "Código de acesso inválido.", Code: http.StatusUnauthorized})
		}
		claims, err := token.GetClaims(string(accessTokenBytes))
		if err != nil {
			log.Printf("failed to extract email from token claims, erro %v\n", err)
			return c.JSON(http.StatusInternalServerError, defaultResponse{Message: "Falha ao processar token de acesso.", Code: http.StatusInternalServerError})
		}
		email := claims["email"]
		foundCandidate, err := dbClient.GetCandidateByEmail(email, year)

		emailMessage := buildContactMessage(messageType, description)
		if err := emailClient.Send(emailClient.Email, []string{contactEmail}, "Fale conosco: "+subject+" candidatos.info", emailMessage); err != nil {
			log.Printf("failed to sending email (%s):%q", contactEmail, err)
			return c.Render(http.StatusOK, "fale-conosco-success.html", map[string]interface{}{
				"ErrorMsg": "Erro inesperado. Por favor, tente novamente mais tarde.",
				"Success":  false,
			})
		}

		return c.Render(http.StatusOK, "fale-conosco-success.html", map[string]interface{}{
			"Candidate":    foundCandidate,
			"Success":      true,
			"Year":         year,
			"SequentialID": foundCandidate.SequencialCandidate,
		})
	}
}
