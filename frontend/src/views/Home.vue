<template>
  <div class="font-sans leading-loose">
    <!-- Sobre -->
    <section class="max-w-4xl mx-auto flex flex-col space-y-6 p-8 text-sm sm:text-lg text-text">
      <h3 class="text-xl text-center font-semibold">Sobre</h3>

      <p class="text-center">
        <strong>Página ainda em construção.</strong>
      </p>

      <p>
        O TSE já provê várias informações sobre candidaturas no repositório de dados eleitorais, por exemplo, nome na urna, número, etc. Ações de transparência pública como esta alimenta páginas como esta e esta, que permite que as informações sobre as candidaturas sejam encontradas e visualizadas em um navegador. Um dos problemas é que essas páginas não contêm informações relacionadas a biografia, nem as intenções caso eleita (propostas) e nem meios para o eleitorado se comunicar com as pessoas candidatas. Outro problema é que as páginas mostram diferentes aspectos das candidaturas, não existindo um lugar que publique os diversos tipos de informações e/ou proveja os dados consolidados.
      </p>
      <p>
        Uma vez que o impacto das pessoas candidatas no eleitorado potencial é muito desigual, acreditamos que unir levantamento colaborativo, agregação e publicação de informações sobre as candidaturas são um meio importante de ajudar a população a conhecer mais e melhor seus possíveis representantes. Por exemplo, uma candidata pode contribuir com conteúdo do site atualizando suas propostas e uma desenvolvedora pode contribuir com o código que obtém informações sobre patrimônio declarado dos políticos. Dessa forma, o objetivo deste projeto é a criação de uma plataforma colaborativa e aberta que permita o enriquecimento e disponibilização das informações sobre as candidaturas legislativas e executivas no âmbito municipal (vereadores e prefeitos).
      </p>
    </section>
    <!-- End of Sobre -->

    <!-- Equipe -->
    <section class="mt-10 p-8 bg-text text-flux-candidate">
      <div class="max-w-4xl mx-auto">
        <h3 class="text-xl text-center font-semibold">Equipe</h3>

        <div class="grid gap-4 grid-cols-2 sm:grid-cols-3 md:grid-cols-4">
          <app-team-member v-for="(member, index) in team" :key="index" :member="member" />
        </div>
      </div>
    </section>
    <!-- End of Equipe -->
  </div>
</template>

<script>
import AppTeamMember from '@/components/TeamMember.vue';
import { mapState } from 'vuex';

export default {
  name: 'Home',
  components: {
    AppTeamMember,
  },
  computed: {
    ...mapState({
      team: (state) => state.team,
    }),
  },
}
</script>
