package main

import (
	b64 "encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/candidatos-info/descritor"
	"github.com/candidatos-info/site/db"
	"github.com/candidatos-info/site/email"
	"github.com/candidatos-info/site/exception"
	"github.com/candidatos-info/site/token"
	"github.com/labstack/echo"
)

const (
	imageWidth  = 100
	imageHeight = 30
	logoURL     = "https://s3.amazonaws.com/candidatos.info-public/Logo-1px.png"
)

func newSouCandidatoFormHandler(db *db.Client, tokenService *token.Token, emailClient *email.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		email := c.FormValue("email")
		return c.Render(http.StatusOK, "sou-candidato-success.html", map[string]interface{}{
			"Text": login(db, tokenService, emailClient, email),
		})
	}
}

func login(db *db.Client, tokenService *token.Token, emailClient *email.Client, email string) string {
	if !emailRegex.MatchString(email) {
		return fmt.Sprintf("email inválido %s", email)
	}
	foundCandidate, err := db.GetCandidateByEmail(strings.ToUpper(email), globals.Year)
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
	subject := fmt.Sprintf("Link para acesso à candidatura %d de %s/%s", foundCandidate.BallotNumber, foundCandidate.City, foundCandidate.State)
	if err := emailClient.Send(emailClient.Email, []string{foundCandidate.Email}, subject, emailMessage); err != nil {
		log.Printf("failed on sending email (%s):%q\n", email, err)
		return "Erro inesperado. Por favor tentar novamente mais tarde."
	}
	return fmt.Sprintf("Email com código de acesso enviado para %s. Verifique sua caixa de spam caso não encontre.", email)
}

func buildProfileAccessEmail(candidate *descritor.CandidateForDB, accessToken string) string {
	link := fmt.Sprintf("%s/atualizar-candidatura?access_token=%s", siteURL, accessToken)
	var emailBodyBuilder strings.Builder
	emailBodyBuilder.WriteString(fmt.Sprintf("Olá, %s!<br><br>", candidate.Name))
	emailBodyBuilder.WriteString(fmt.Sprintf("Identificamos através dos dados públicos do TSE que você está cadastrado na eleição de %d na cidade de %s no estado de %s como %s.<br><br><br>", globals.Year, candidate.City, candidate.State, candidate.Role))
	emailBodyBuilder.WriteString(fmt.Sprintf("Recebemos sua solicitação para acessar a plataforma candidatos.info e editar seu perfil. Para acessar <a href=\"%s\">clique aqui</a>. <br><br>Caso o link não esteja funcionando copie e cole no navegador o seguinte link:<br> %s", link, link))
	emailBodyBuilder.WriteString("<br><br><br>Caso tenha recebido este email por engano, por favor desconsidere-o.<br>")
	emailBodyBuilder.WriteString(fmt.Sprintf("Atenciosamente, <br><img src=%s width=%d height=%d>", logoURL, imageWidth, imageHeight))
	return emailBodyBuilder.String()
}

func souCandidatoGET(c echo.Context) error {
	return c.Render(http.StatusOK, "sou-candidato.html", nil)
}
