<template>
  <nav class="navbar">
    <div class="container navbar-inner">
      <RouterLink to="/dashboard" class="navbar-brand">⚓ svCompare</RouterLink>
      <div class="navbar-user">
        <span>{{ auth.user?.username }}</span>
        <button class="btn btn-secondary btn-sm" @click="handleLogout">Logout</button>
      </div>
    </div>
  </nav>
</template>

<script setup>
import { useAuthStore } from '../stores/auth.js'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()

async function handleLogout() {
  await auth.logout()
  router.push('/login')
}
</script>
