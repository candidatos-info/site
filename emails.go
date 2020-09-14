package main

import (
	"fmt"
	"strings"

	"github.com/candidatos-info/descritor"
)

func buildProfileAccessEmail(candidate *descritor.CandidateForDB, accessToken string) string {
	var emailBodyBuilder strings.Builder
	emailBodyBuilder.WriteString(fmt.Sprintf("Olá, %s!\n", candidate.Name))
	emailBodyBuilder.WriteString(fmt.Sprintf("para acessar seu perfil click no link: %s/profile?access_token=%s\n", siteURL, accessToken))
	return emailBodyBuilder.String()
}

func buildReportEmail(candidate *descritor.CandidateForDB, report string) string {
	var emailBodyBuilder strings.Builder
	emailBodyBuilder.WriteString(fmt.Sprintf("Nova denúcia do candidato %s: ", candidate.Name))
	emailBodyBuilder.WriteString(fmt.Sprintf("\n\n%s\n", report))
	return emailBodyBuilder.String()
}
