package main

import (
	"fmt"
	"strings"

	"github.com/candidatos-info/descritor"
)

func buildProfileAccessEmail(candidate *descritor.CandidateForDB, accessToken string) string {
	return fmt.Sprintf(`
	Olá, %s!<br><br>
	Identificamos através dos dados públicos do TSE que você está cadastrado na eleição de %d na cidade de %s no estado de %s como %s.<br>
	Recebemos sua solicitação para acessar a plataforma candidatos.info e editar seu perfil. Para isso clique no seguinte link (ou copie e cole no navegador): <a src="%s/atualizar-candidatura?access_token=%s">%s/atualizar-candidatura?access_token=%s</a>
	<br><br><br>Caso tenha recebido este email por engano apenas desconsidere-o.<br>`, candidate.Name, candidate.Year, candidate.City, candidate.State, candidate.Role, siteURL, accessToken)
}

func buildReportEmail(candidate *descritor.CandidateForDB, report string) string {
	var emailBodyBuilder strings.Builder
	emailBodyBuilder.WriteString(fmt.Sprintf("Nova denúcia do candidato %s: <br><br>", candidate.Name))
	emailBodyBuilder.WriteString(fmt.Sprintf("\n\n%s\n", report))
	return emailBodyBuilder.String()
}
