<template>
  <div class="bg-flux-visitor">
    <app-nav/>
    <router-view/>
    <app-footer />
  </div>
</template>

<script>
import AppNav from '@/components/Nav.vue';
import AppFooter from '@/components/Footer.vue';

export default {
  name: 'AppLayout',
  components: {
    AppNav,
    AppFooter,
  },
}
</script>
