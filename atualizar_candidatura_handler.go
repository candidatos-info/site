package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/candidatos-info/descritor"
	"github.com/candidatos-info/site/db"
	"github.com/candidatos-info/site/token"
	"github.com/labstack/echo"
)

func newAtualizarCandidaturaFormHandler(dbClient *db.Client, year int) echo.HandlerFunc {
	return func(c echo.Context) error {
		encodedAccessToken := c.FormValue("token")
		if encodedAccessToken == "" {
			log.Printf("empty token")
			return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": "Código de acesso inválido",
				"Success":  false,
			})
		}
		accessTokenBytes, err := base64.StdEncoding.DecodeString(encodedAccessToken)
		if err != nil {
			log.Printf("error decoding access token %s", string(encodedAccessToken))
			return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": "Erro inesperado. Por favor, tente novamente mais tarde.",
				"Success":  false,
			})
		}
		if !tokenService.IsValid(string(accessTokenBytes)) {
			log.Printf("invalid access token")
			return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": "Código de acesso inválido",
				"Success":  false,
			})
		}
		claims, err := token.GetClaims(string(accessTokenBytes))
		if err != nil {
			log.Printf("failed to extract email from token claims, error %v\n", err)
			return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": "Erro inesperado. Por favor, tente novamente mais tarde.",
				"Success":  false,
			})
		}
		// Processing and validating form values.
		// TODO: get and process contact.
		numTags, err := strconv.Atoi(c.FormValue("numTags"))
		if err != nil {
			log.Printf("invalid numTags :%s, error %v\n", c.FormValue("numTags"), err)
			return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": "Erro inesperado. Por favor, tente novamente mais tarde.",
				"Success":  false,
			})
		}
		var props []*descritor.Proposal
		for i := 0; i < numTags; i++ {
			p := descritor.Proposal{
				Topic:       c.FormValue(fmt.Sprintf("descriptions[%d][tag]", i)),
				Description: c.FormValue(fmt.Sprintf("descriptions[%d][description]", i)),
			}
			if len(p.Description) > maxDescriptionTextSize {
				return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
					"ErrorMsg": fmt.Sprintf("Tamanho máximo de descrição é de %d caracteres. Tamanho das descrição do tópico %s é de %d caracteres", maxDescriptionTextSize, p.Topic, len(p.Description)),
					"Success":  false,
				})
			}
			props = append(props, &p)
		}
		bio := c.FormValue("biography")
		if len(bio) > maxBiographyTextSize {
			return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": fmt.Sprintf("Tamanho máximo de descrição é de %d caracteres.", maxBiographyTextSize),
				"Success":  false,
			})
		}
		// Fetching candidate and updating counters.
		email := claims["email"]
		candidate, err := dbClient.GetCandidateByEmail(email, year)
		if err != nil {
			log.Printf("failed to find candidate using email from token claims, erro %v\n", err)
			return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": "Código de acesso inválido",
				"Success":  false,
			})
		}
		candidate.Biography = bio
		candidate.Proposals = props
		counter := 0.0
		if candidate.Biography != "" {
			counter++
		}
		if candidate.Proposals != nil && len(candidate.Proposals) > 0 {
			counter++
		}
		if candidate.Contacts != nil && len(candidate.Contacts) > 0 {
			counter++
		}
		candidate.Transparency = counter / 3.0

		// Updating candidates.
		if _, err := dbClient.UpdateCandidateProfile(candidate); err != nil {
			log.Printf("failed to update candidates profile, erro %v\n", err)
			return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": "Erro inesperado. Por favor, tente novamente mais tarde.",
				"Success":  false,
			})
		}
		return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
			"ErrorMsg":     "<strong>Seus dados foram atualizados com sucesso!</strong>",
			"Success":      true,
			"Year":         year,
			"SequentialID": candidate.SequencialCandidate,
		})
	}
}
func newAtualizarCandidaturaHandler(dbClient *db.Client, tags []string, year int) echo.HandlerFunc {
	return func(c echo.Context) error {
		encodedAccessToken := c.QueryParam("access_token")
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
		if err != nil {
			log.Printf("failed to find candidate using email from token claims (email:%s, currentYear:%d), erro %q\n", email, currentYear, err)
			return c.JSON(http.StatusInternalServerError, defaultResponse{Message: "Falha ao buscar informaçōes de candidatos.", Code: http.StatusInternalServerError})
		}
		// @TODO: só mostrar a tela de aceitar-termo caso o candidato ainda não tenha aceitado
		if false {
			return c.Render(http.StatusOK, "aceitar-termo.html", map[string]interface{}{
				"Token":      encodedAccessToken,
				"TextoTermo": "Lorem ipsum dolor sit amet, consectetur adipisicing elit. Aliquam aliquid aspernatur at atque distinctio dolores in, iusto labore mollitia optio quia quibusdam quod tempora! Iste neque optio placeat provident quaerat. Lorem ipsum dolor sit amet, consectetur adipisicing elit. Aliquam aliquid aspernatur at atque distinctio dolores in, iusto labore mollitia optio quia quibusdam quod tempora! Iste neque optio placeat provident quaerat. Lorem ipsum dolor sit amet, consectetur adipisicing elit. Aliquam aliquid aspernatur at atque distinctio dolores in, iusto labore mollitia optio quia quibusdam quod tempora! Iste neque optio placeat provident quaerat.",
			})
		}
		r := c.Render(http.StatusOK, "atualizar-candidato.html", map[string]interface{}{
			"Token":     encodedAccessToken,
			"AllTags":   tags,
			"Candidato": foundCandidate,
		})
		fmt.Println(r)
		return r
	}
}
