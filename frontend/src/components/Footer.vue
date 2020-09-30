<template>
  <footer class="px-2 py-8">
    <app-logo class="block h-8 w-auto mx-auto text-secondary-button"/>
  </footer>
</template>

<script>
import AppLogo from '@/components/Logo.vue';

export default {
  name: "Footer",
  components: {
    AppLogo,
  },
}
</script>
