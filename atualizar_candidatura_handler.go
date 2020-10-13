package main

import (
	b64 "encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/candidatos-info/site/db"
	"github.com/candidatos-info/site/token"
	"github.com/labstack/echo"
)

func newAtualizarCandidaturaHandler(dbClient *db.Client, tags []string, year int) echo.HandlerFunc {
	return func(c echo.Context) error {
		encodedAccessToken := c.QueryParam("access_token")
		if encodedAccessToken == "" {
			return c.JSON(http.StatusBadRequest, defaultResponse{Message: "Código de acesso é inválido.", Code: http.StatusBadRequest})
		}
		accessTokenBytes, err := b64.StdEncoding.DecodeString(encodedAccessToken)
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
		if err != nil {
			log.Printf("failed to find candidate using email from token claims (email:%s, currentYear:%d), erro %q\n", email, currentYear, err)
			return c.JSON(http.StatusInternalServerError, defaultResponse{Message: "Falha ao buscar informaçōes de candidatos.", Code: http.StatusInternalServerError})
		}
		// @TODO: só mostrar a tela de aceitar-termo caso o candidato ainda não tenha aceitado
		r := c.Render(http.StatusOK, "atualizar-candidato.html", map[string]interface{}{
			"Token":     string(accessTokenBytes),
			"AllTags":   tags,
			"Candidato": foundCandidate,
		})
		fmt.Println(r)
		return r
	}
}
