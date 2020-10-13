package main

import (
	b64 "encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/candidatos-info/site/db"
	"github.com/candidatos-info/site/email"
	"github.com/candidatos-info/site/exception"
	"github.com/candidatos-info/site/token"
	"github.com/labstack/echo"
)

func newSouCandidatoFormHandler(db *db.Client, tokenService *token.Token, emailClient *email.Client, year int) echo.HandlerFunc {
	return func(c echo.Context) error {
		email := c.FormValue("email")
		return c.Render(http.StatusOK, "sou-candidato-success.html", map[string]interface{}{
			"Text": login(db, tokenService, emailClient, email, year),
		})
	}
}

func login(db *db.Client, tokenService *token.Token, emailClient *email.Client, email string, year int) string {
	if !emailRegex.MatchString(email) {
		return fmt.Sprintf("email inválido %s", email)
	}
	foundCandidate, err := db.GetCandidateByEmail(strings.ToUpper(email), year)
	switch {
	case err != nil && err.(*exception.Exception).Code == exception.NotFound:
		return fmt.Sprintf("O email %s não foi encontrado no registro do TSE. Por favor verifique se houve algum erro na digitação.", email)
	case err != nil:
		log.Printf("erro searching for candidates by e-mail (%s):%q", email, err)
		return "Erro inesperado. Por favor tentar novamente mais tarde."
	}
	accessToken, err := tokenService.GetToken(email)
	if err != nil {
		log.Printf("failed to get acess token for e-mail (%s):%q", email, err)
		return "Erro inesperado. Por favor tentar novamente mais tarde."
	}
	encodedAccessToken := b64.StdEncoding.EncodeToString([]byte(accessToken))
	emailMessage := buildProfileAccessEmail(foundCandidate, encodedAccessToken)
	if err := emailClient.Send(emailClient.Email, []string{foundCandidate.Email}, "Código para acessar candidatos.info", emailMessage); err != nil {
		log.Printf("failed to sending email (%s):%q", email, err)
		return "Erro inesperado. Por favor tentar novamente mais tarde."
	}
	return fmt.Sprintf("Email com código de acesso enviado para <strong>%s</strong>. Verifique sua caixa de spam caso não encontre.", email)
}

func souCandidatoGET(c echo.Context) error {
	return c.Render(http.StatusOK, "sou-candidato.html", nil)
}
