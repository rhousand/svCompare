<template>
  <div>
    <NavBar v-if="auth.user" />
    <main class="container">
      <RouterView />
    </main>
  </div>
</template>

<script setup>
import { useAuthStore } from './stores/auth.js'
import NavBar from './components/NavBar.vue'

const auth = useAuthStore()
</script>
