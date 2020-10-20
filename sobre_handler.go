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
			newSocialLink("twitter", "https://twitter.com/anapaulagomess"),
			newSocialLink("github", "https://github.com/anapaulagomes"),
		}),
		newTeamMember("Aurélio Buarque", "Desenvolvedor", "/img/team/aurelio.jpeg", []*socialLink{
			newSocialLink("twitter", "https://twitter.com/abuarquemf"),
			newSocialLink("github", "https://github.com/ABuarque"),
			newSocialLink("linkedin", "https://www.linkedin.com/in/aurelio-buarque/"),
		}),
		newTeamMember("Bruno Morassutti", "Palpiteiro jurídico", "/img/team/bruno.jpeg", []*socialLink{
			newSocialLink("github", "https://github.com/jedibruno"),
		}),
		newTeamMember("Daniel Fireman", "Coordenador", "/img/team/daniel.jpeg", []*socialLink{
			newSocialLink("twitter", "https://twitter.com/daniellfireman"),
			newSocialLink("github", "https://github.com/danielfireman"),
		}),
		newTeamMember("Eduardo Cuducos", "Palpiteiro", "/img/team/eduardo.jpeg", []*socialLink{
			newSocialLink("twitter", "https://twitter.com/cuducos"),
			newSocialLink("github", "https://github.com/cuducos"),
		}),
		newTeamMember("Evelyn Gomes", "Articuladora ", "/img/team/evelyn.jpeg", []*socialLink{
			newSocialLink("instagram", "https://www.instagram.com/evygomes"),
		}),
		newTeamMember("Laura Cavalcante", "Advogada ", "/img/team/laura.jpeg", []*socialLink{
			newSocialLink("instagram", "https://www.instagram.com/lauracavalcante/"),
		}),
		newTeamMember("Mariana Souto", "Designer ", "/img/team/mariana.jpeg", []*socialLink{
			newSocialLink("github", "https://github.com/soutoam"),
			newSocialLink("instagram", "https://www.instagram.com/soutoam/"),
			newSocialLink("linkedin", "https://www.linkedin.com/in/soutomariana/"),
		}),
	}
}

func sobreHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "sobre.html", map[string]interface{}{
		"Team": getTeamMembers(),
	})
}
