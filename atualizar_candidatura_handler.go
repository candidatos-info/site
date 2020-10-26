package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/candidatos-info/descritor"
	"github.com/candidatos-info/site/db"
	"github.com/candidatos-info/site/exception"
	"github.com/candidatos-info/site/token"
	"github.com/labstack/echo"
)

const (
	maxBiographyTextSize     = 500
	maxProposalsTextSize     = 100
	maxProposalsPerCandidate = 5
	maxTagsSize              = 4
	maxProposals             = 5
	maxContactsTextSize      = 100
	numTagsFieldName         = "numTags"
	bioFieldName             = "biography"
	contactFieldName         = "contact"
	providerFieldName        = "provider"
)

type atualizarCandidaturaParams struct {
	NumTags   int
	Bio       string
	Contacts  []*descritor.Contact
	Proposals []*descritor.Proposal
}

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
		// Processing and validating form values.
		params, err := parseFormValues(c)
		if err != nil {
			return err
		}
		candidate.Biography = params.Bio
		candidate.Proposals = params.Proposals
		candidate.Contacts = params.Contacts
		candidate.Transparency = 100 // Since we made all fields mandatory, if the candidate has registered, its transparency will be 100%

		// Updating candidates DB
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

func parseFormValues(ctx echo.Context) (atualizarCandidaturaParams, error) {
	numTags, err := strconv.Atoi(ctx.FormValue(numTagsFieldName))
	if err != nil {
		log.Printf("invalid num tags %s :%s, error %v\n", numTagsFieldName, ctx.FormValue(numTagsFieldName), err)
		return atualizarCandidaturaParams{}, ctx.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
			"ErrorMsg": "Erro inesperado. Por favor, tente novamente mais tarde.",
			"Success":  false,
		})
	}
	if numTags == 0 {
		return atualizarCandidaturaParams{}, ctx.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
			"ErrorMsg": "É necessário o preenchimento de, ao menos, uma pauta.",
			"Success":  false,
		})
	}
	var props []*descritor.Proposal
	for i := 0; i < numTags; i++ {
		p := descritor.Proposal{
			Topic:       ctx.FormValue(fmt.Sprintf("descriptions[%d][tag]", i)),
			Description: ctx.FormValue(fmt.Sprintf("descriptions[%d][description]", i)),
		}
		if len(p.Description) == 0 {
			return atualizarCandidaturaParams{}, ctx.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": fmt.Sprintf("O campo proposta da pauta %s é obrigatório", p.Topic),
				"Success":  false,
			})
		}
		if len(p.Description) > maxProposalsTextSize {
			return atualizarCandidaturaParams{}, ctx.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
				"ErrorMsg": fmt.Sprintf("Tamanho da proposta da pauta %s é de %d caracteres. O tamanho máximo permitido é de %d.", p.Topic, len(p.Description), maxProposalsTextSize),
				"Success":  false,
			})
		}
		props = append(props, &p)
	}
	bio := ctx.FormValue(bioFieldName)
	if len(bio) == 0 {
		return atualizarCandidaturaParams{}, ctx.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
			"ErrorMsg": "Biografia é um campo obrigatório. Por favor, preencher",
			"Success":  false,
		})
	}
	if len(bio) > maxBiographyTextSize {
		return atualizarCandidaturaParams{}, ctx.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
			"ErrorMsg": fmt.Sprintf("Tamanho máximo do campo mini-biografia é de %d caracteres.", maxBiographyTextSize),
			"Success":  false,
		})
	}
	contact := ctx.FormValue(contactFieldName)
	if len(contact) == 0 {
		return atualizarCandidaturaParams{}, ctx.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
			"ErrorMsg": "Contato é um campo obrigatório. Por favor, preencher",
			"Success":  false,
		})
	}
	if len(contact) > maxContactsTextSize {
		return atualizarCandidaturaParams{}, ctx.Render(http.StatusOK, "atualizar-candidato-success.html", map[string]interface{}{
			"ErrorMsg": fmt.Sprintf("Tamanho máximo do campo contato é de %d caracteres.", maxContactsTextSize),
			"Success":  false,
		})
	}
	return atualizarCandidaturaParams{
		NumTags: numTags,
		Bio:     bio,
		Contacts: []*descritor.Contact{&descritor.Contact{
			SocialNetwork: ctx.FormValue(providerFieldName),
			Value:         contact,
		}},
		Proposals: props,
	}, nil
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
