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
                    {provider: 'twitter', link: 'https://twitter.com/anapaulagomes'},
                    {provider: 'github', link: 'https://twitter.com/anapaulagomes'},
                ]
            },
            {
                name: 'Aurélio Buarque',
                profile_image: '/img/team/aurelio.jpeg',
                position: 'Desenvolvedor',
                socials: [
                    {provider: 'twitter', link: 'https://twitter.com/anapaulagomes'},
                ]
            },
            {
                name: 'Bruno Morassuti',
                profile_image: '/img/team/bruno.jpeg',
                position: 'Palpiteiro jurídico',
                socials: [
                    {provider: 'twitter', link: 'https://twitter.com/anapaulagomes'},
                ]
            },
            {
                name: 'Daniel Fireman',
                profile_image: '/img/team/daniel.jpeg',
                position: 'Coordenador',
                socials: [
                    {provider: 'twitter', link: 'https://twitter.com/anapaulagomes'},
                ]
            },
            {
                name: 'Eduardo Cuducos',
                profile_image: '/img/team/eduardo.jpeg',
                position: 'Palpiteiro',
                socials: [
                    {provider: 'twitter', link: 'https://twitter.com/anapaulagomes'},
                ]
            },
            {
                name: 'Evelyn Gomes',
                profile_image: '/img/team/evelyn.jpeg',
                position: 'Articuladora',
                socials: [
                    {provider: 'twitter', link: 'https://twitter.com/anapaulagomes'},
                ]
            },
            {
                name: 'Laura Cavalcante',
                profile_image: '/img/team/laura.jpeg',
                position: 'Advogada',
                socials: [
                    {provider: 'twitter', link: 'https://twitter.com/anapaulagomes'},
                ]
            },
            {
                name: 'Mariana Souto',
                profile_image: '/img/team/mariana.jpeg',
                position: 'Designer',
                socials: [
                    {provider: 'twitter', link: 'https://twitter.com/anapaulagomes'},
                ]
            },
        ],
    },
    mutations: {},
    actions: {},
    modules: {}
})
