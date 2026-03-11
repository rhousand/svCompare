<template>
  <div class="dashboard">
    <div class="page-header">
      <h2>My Comparisons</h2>
      <button class="btn btn-primary" @click="showCreate = true">+ New Comparison</button>
    </div>

    <!-- Create form -->
    <div v-if="showCreate" class="card create-form">
      <h3>New Comparison</h3>
      <form @submit.prevent="handleCreate">
        <div class="form-group">
          <label>Name</label>
          <input
            v-model="newName"
            type="text"
            placeholder="e.g. 2025 Pacific Northwest Search"
            required
            autofocus
          />
        </div>
        <div v-if="createError" class="error-msg">{{ createError }}</div>
        <div class="form-actions">
          <button type="submit" class="btn btn-primary" :disabled="creating">
            {{ creating ? 'Creating…' : 'Create' }}
          </button>
          <button type="button" class="btn btn-secondary" @click="cancelCreate">Cancel</button>
        </div>
      </form>
    </div>

    <!-- List -->
    <div v-if="store.loading" class="empty-state">Loading…</div>
    <div v-else-if="store.error" class="error-msg">{{ store.error }}</div>
    <div v-else-if="store.list.length === 0" class="empty-state">
      No comparisons yet. Create one to get started.
    </div>
    <div v-else class="comparison-list">
      <div v-for="c in store.list" :key="c.id" class="card comparison-item">
        <div class="comparison-info">
          <RouterLink :to="`/comparison/${c.id}`" class="comparison-name">{{ c.name }}</RouterLink>
          <span class="comparison-meta">Expires {{ formatDate(c.expires_at) }}</span>
        </div>
        <div class="comparison-actions">
          <RouterLink :to="`/comparison/${c.id}`" class="btn btn-secondary btn-sm">Open</RouterLink>
          <button class="btn btn-danger btn-sm" @click="handleDelete(c.id)">Delete</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useComparisonsStore } from '../stores/comparisons.js'

const store = useComparisonsStore()
const router = useRouter()

const showCreate = ref(false)
const newName = ref('')
const creating = ref(false)
const createError = ref('')

onMounted(() => store.fetchList())

function cancelCreate() {
  showCreate.value = false
  newName.value = ''
  createError.value = ''
}

async function handleCreate() {
  createError.value = ''
  creating.value = true
  try {
    const c = await store.create(newName.value.trim())
    cancelCreate()
    router.push(`/comparison/${c.id}`)
  } catch (e) {
    createError.value = e.message
  } finally {
    creating.value = false
  }
}

async function handleDelete(id) {
  if (!confirm('Delete this comparison? This cannot be undone.')) return
  try {
    await store.remove(id)
  } catch (e) {
    alert(e.message)
  }
}

function formatDate(str) {
  return new Date(str).toLocaleDateString()
}
</script>
