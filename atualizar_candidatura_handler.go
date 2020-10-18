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
	"github.com/candidatos-info/site/token"
	"github.com/labstack/echo"
)

const (
	maxProposals     = 5
	maxContactsChars = 100
)

var (
	socialNetworksUI = map[string]string{
		"facebook":  "Facebook",
		"instagram": "Instagram",
		"twitter":   "Twitter",
		"email":     "E-mail",
		"whatsapp":  "Whatsapp",
		"telefone":  "Telefone",
		"paginaWeb": "Página Web",
	}
)

func newAtualizarCandidaturaFormHandler(dbClient *db.Client) echo.HandlerFunc {
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
		contact := c.FormValue("contact")
		if len(contact) == 0 {
			return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": "Contato é um campo obrigatório. Por favor, preencher",
				"Success":  false,
			})
		}
		if len(contact) > maxContactsChars {
			return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": fmt.Sprintf("Tamanho máximo do contato é de %d caracteres.", maxBiographyTextSize),
				"Success":  false,
			})
		}
		// Fetching candidate and updating counters.
		email := claims["email"]
		candidate, err := dbClient.GetCandidateByEmail(email, globals.Year)
		if err != nil {
			log.Printf("failed to find candidate using email from token claims, erro %v\n", err)
			return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": "Código de acesso inválido",
				"Success":  false,
			})
		}
		candidate.Biography = bio
		candidate.Proposals = props
		candidate.Contacts = []*descritor.Contact{getContact(contact, c.FormValue("provider"))}
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
		candidate.Transparency = (counter / 3.0) * 100

		// Updating candidates.
		if _, err := dbClient.UpdateCandidateProfile(candidate); err != nil {
			log.Printf("failed to update candidates profile, erro %v\n", err)
			return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": "Erro inesperado. Por favor, tente novamente mais tarde.",
				"Success":  false,
			})
		}
		return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
			"ErrorMsg":     "Seus dados foram atualizados com sucesso!",
			"Success":      true,
			"SequentialID": candidate.SequencialCandidate,
		})
	}
}

func mapMonthsToPortuguese(month time.Month) string {
	switch int(month) {
	case 1:
		return "Janeiro"
	case 2:
		return "Fevereiro"
	case 3:
		return "Março"
	case 4:
		return "Abril"
	case 5:
		return "Maio"
	case 6:
		return "Junho"
	case 7:
		return "Julho"
	case 8:
		return "Agosto"
	case 9:
		return "Setembro"
	case 10:
		return "Outubro"
	case 11:
		return "Novembro"
	case 12:
		return "Dezembro"
	}
	return ""
}

func newAtualizarCandidaturaHandler(dbClient *db.Client, tags []string) echo.HandlerFunc {
	return func(c echo.Context) error {
		encodedAccessToken := c.QueryParam("access_token")
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
		email := claims["email"]
		foundCandidate, err := dbClient.GetCandidateByEmail(email, globals.Year)
		if err != nil {
			log.Printf("failed find candidate on DB (email:%s), error %v\n", email, err)
			switch {
			case err != nil && err.(*exception.Exception).Code == exception.NotFound:
				return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
					"ErrorMsg": fmt.Sprintf("Não encontramos um cadastro de candidatura através do email %s. Por favor verifique se o email está correto.", email),
					"Success":  false,
				})
			case err != nil:
				return c.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
					"ErrorMsg": "Erro inesperado. Por favor, tente novamente mais tarde.",
					"Success":  false,
				})
			}
		}
		_, month, day := time.Now().Date()
		if foundCandidate.AcceptedTerms.IsZero() {
			return c.Render(http.StatusOK, "aceitar-termo.html", map[string]interface{}{
				"Token":                encodedAccessToken,
				"Candidate":            foundCandidate,
				"termsAcceptanceDay":   day,
				"termsAcceptanceMonth": mapMonthsToPortuguese(month),
			})
		}
		r := c.Render(http.StatusOK, "atualizar-candidato.html", map[string]interface{}{
			"Token":          encodedAccessToken,
			"AllTags":        tags,
			"Candidato":      foundCandidate,
			"MaxProposals":   maxProposals,
			"SocialNetworks": socialNetworksUI,
		})
		fmt.Println(r)
		return r
	}
}

func getContact(addr, provider string) *descritor.Contact {
	if strings.HasPrefix(addr, "http") {
		return &descritor.Contact{
			SocialNetwork: provider,
			Value:         addr,
		}
	}
	addrPrefix := ""
	switch provider {
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
	return &descritor.Contact{
		SocialNetwork: provider,
		Value:         addrPrefix + addr,
	}
}
