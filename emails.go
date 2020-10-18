package main

import (
	"fmt"
	"strings"

	"github.com/candidatos-info/descritor"
)

func buildReportEmail(candidate *descritor.CandidateForDB, report string) string {
	var emailBodyBuilder strings.Builder
	emailBodyBuilder.WriteString(fmt.Sprintf("Nova denúcia do candidato %s: <br><br>", candidate.Name))
	emailBodyBuilder.WriteString(fmt.Sprintf("\n\n%s\n", report))
	return emailBodyBuilder.String()
}
