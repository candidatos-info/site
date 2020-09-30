<template>
  <div class="shadow rounded px-2 py-6 flex flex-col space-y-4">
    <img class="mx-auto w-20 h-20 rounded-full shadow object-cover" style="filter: grayscale(100%);" :src="member.profile_image" alt="Imagem de perfil" />
    <div>
      <p class="text-center text-sm text-flux-candidate">{{ member.name }}</p>
      <p class="text-center text-xs text-disabled-button">{{ member.position }}</p>
    </div>
    <ul class="flex justify-center space-x-2">
      <li v-for="(social, index) in member.socials" :key="index">
        <a :href="social.link">
          <app-social-icon class="text-secondary-button h-4 w-4" :provider="social.provider" />
        </a>
      </li>
    </ul>
  </div>
</template>

<script>
import AppSocialIcon from '@/components/SocialIcon.vue';

export default {
  name: "TeamMember",
  components: {
    AppSocialIcon,
  },
  props: {
    member: Object,
  },
}
</script>
