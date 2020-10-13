package main

import (
	"net/http"

	"github.com/labstack/echo"
)

type socialLink struct {
	Provider string
	Link     string
}

type teamMember struct {
	Name        string
	Title       string
	ImageURL    string
	SocialLinks []*socialLink
}

func newSocialLink(provider string, link string) *socialLink {
	return &socialLink{
		Provider: provider,
		Link:     link,
	}
}

func newTeamMember(name string, title string, imageURL string, socialLinks []*socialLink) *teamMember {
	return &teamMember{
		Name:        name,
		Title:       title,
		ImageURL:    imageURL,
		SocialLinks: socialLinks,
	}
}

func getTeamMembers() []*teamMember {
	return []*teamMember{
		newTeamMember("Ana Paula Gomes", "Pitaqueira", "/img/team/ana.jpeg", []*socialLink{
			newSocialLink("linkedin", "#"),
		}),
		newTeamMember("Aurélio Buarque", "Desenvolvedor", "/img/team/aurelio.jpeg", []*socialLink{
			newSocialLink("twitter", "#"),
			newSocialLink("github", "#"),
		}),
		newTeamMember("Bruno Morassutti", "Palpiteiro jurídico", "/img/team/bruno.jpeg", []*socialLink{
			newSocialLink("twitter", "#"),
			newSocialLink("linkedin", "#"),
		}),
		newTeamMember("Daniel Fireman", "Coordenador", "/img/team/daniel.jpeg", []*socialLink{
			newSocialLink("twitter", "#"),
			newSocialLink("github", "#"),
			newSocialLink("linkedin", "#"),
			newSocialLink("instagram", "#"),
		}),
		newTeamMember("Eduardo Cuducos", "Palpiteiro", "/img/team/eduardo.jpeg", []*socialLink{
			newSocialLink("twitter", "#"),
			newSocialLink("github", "#"),
		}),
		newTeamMember("Evelyn Gomes", "Articuladora ", "/img/team/evelyn.jpeg", []*socialLink{
			newSocialLink("twitter", "#"),
			newSocialLink("linkedin", "#"),
			newSocialLink("github", "#"),
		}),
		newTeamMember("Laura Cavalcante", "Advogada ", "/img/team/laura.jpeg", []*socialLink{
			newSocialLink("twitter", "#"),
			newSocialLink("linkedin", "#"),
		}),
		newTeamMember("Mariana Souto", "Designer ", "/img/team/mariana.jpeg", []*socialLink{
			newSocialLink("twitter", "#"),
			newSocialLink("instagram", "#"),
			newSocialLink("linkedin", "#"),
		}),
	}
}

func sobreHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "sobre.html", map[string]interface{}{
		"Team": getTeamMembers(),
	})
}
