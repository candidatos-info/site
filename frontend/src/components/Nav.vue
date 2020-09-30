<template>
  <nav class="bg-flux-candidate">
    <!-- Primary Navigation Menu -->
    <div class="px-4 sm:px-6 lg:px-8 flex justify-center">
      <div class="flex justify-between h-16">
        <div class="flex">
          <!-- Logo -->
          <div class="flex-shrink-0 flex items-center">
            <router-link to="/">
              <app-logo class="block h-10 w-auto"/>
            </router-link>
          </div>
        </div>
      </div>
    </div>
  </nav>
</template>

<script>
import AppLogo from '@/components/Logo.vue';

export default {
  name: 'Nav',
  components: {
    AppLogo,
  },
}
</script>

<style scoped>

</style>