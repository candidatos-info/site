import Vue from 'vue'
import Vuex from 'vuex'

Vue.use(Vuex)

export default new Vuex.Store({
    state: {
        team: [
            {
                name: 'Ana Paula Gomes',
                profile_image: '/img/team/ana.jpeg',
                position: 'Pitaqueira',
                socials: [
                    { provider: 'github', link: 'https://twitter.com/anapaulagomes' },
                ]
            },
            {
                name: 'Aurélio Buarque',
                profile_image: '/img/team/aurelio.jpeg',
                position: 'Desenvolvedor',
                socials: [
                    { provider: 'github', link: 'https://github.com/ABuarque' },
                    { provider: 'twitter', link: 'https://twitter.com/abuarquemf' },
                    { provider: 'linkedin', link: 'https://www.linkedin.com/in/aurelio-buarque/' },
                ]
            },
            {
                name: 'Bruno Morassuti',
                profile_image: '/img/team/bruno.jpeg',
                position: 'Palpiteiro jurídico',
                socials: [
                    { provider: 'github', link: 'https://github.com/jedibruno' },
                ]
            },
            {
                name: 'Daniel Fireman',
                profile_image: '/img/team/daniel.jpeg',
                position: 'Coordenador',
                socials: [
                    { provider: 'twitter', link: 'https://twitter.com/daniellfireman' },
                    { provider: 'github', link: 'https://github.com/danielfireman' }
                ]
            },
            {
                name: 'Eduardo Cuducos',
                profile_image: '/img/team/eduardo.jpeg',
                position: 'Palpiteiro',
                socials: [
                    { provider: 'twitter', link: 'https://twitter.com/cuducos' },
                    { provider: 'github', link: 'https://github.com/cuducos' },
                ]
            },
            {
                name: 'Evelyn Gomes',
                profile_image: '/img/team/evelyn.jpeg',
                position: 'Articuladora',
                socials: [
                    { provider: 'instagram', link: 'https://www.instagram.com/evygomes' },
                ]
            },
            {
                name: 'Laura Cavalcante',
                profile_image: '/img/team/laura.jpeg',
                position: 'Advogada',
                socials: [
                    { provider: 'instagram', link: 'https://www.instagram.com/lauracavalcante/' },
                ]
            },
            {
                name: 'Mariana Souto',
                profile_image: '/img/team/mariana.jpeg',
                position: 'Designer',
                socials: [
                    { provider: 'github', link: 'https://github.com/soutoam' },
                    { provider: 'instagram', link: 'https://www.instagram.com/soutoam/' },
                    { provider: 'linkedin', link: 'https://www.linkedin.com/in/soutomariana/' },
                ]
            },
        ],
    },
    mutations: {},
    actions: {},
    modules: {}
})
