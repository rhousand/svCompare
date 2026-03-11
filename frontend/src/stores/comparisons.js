import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useComparisonsStore = defineStore('comparisons', () => {
  const list = ref([])
  const current = ref(null)
  const loading = ref(false)
  const error = ref(null)

  async function fetchList() {
    loading.value = true
    error.value = null
    try {
      const res = await fetch('/api/comparisons', { credentials: 'include' })
      if (!res.ok) throw new Error('Failed to load comparisons')
      const data = await res.json()
      list.value = data.data || []
    } catch (e) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  async function fetchOne(id) {
    loading.value = true
    error.value = null
    try {
      const res = await fetch(`/api/comparisons/${id}`, { credentials: 'include' })
      if (!res.ok) throw new Error('Failed to load comparison')
      current.value = await res.json()
    } catch (e) {
      error.value = e.message
      current.value = null
    } finally {
      loading.value = false
    }
  }

  async function create(name) {
    const res = await fetch('/api/comparisons', {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name }),
    })
    if (!res.ok) {
      const data = await res.json()
      throw new Error(data.error || 'Failed to create comparison')
    }
    const c = await res.json()
    list.value.unshift(c)
    return c
  }

  async function remove(id) {
    const res = await fetch(`/api/comparisons/${id}`, {
      method: 'DELETE',
      credentials: 'include',
    })
    if (!res.ok) throw new Error('Failed to delete comparison')
    list.value = list.value.filter(c => c.id !== id)
    if (current.value?.id === id) current.value = null
  }

  async function addBoat(comparisonId, name) {
    const res = await fetch(`/api/comparisons/${comparisonId}/boats`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name }),
    })
    if (!res.ok) {
      const data = await res.json()
      throw new Error(data.error || 'Failed to add boat')
    }
    await fetchOne(comparisonId)
  }

  async function removeBoat(comparisonId, boatId) {
    const res = await fetch(`/api/comparisons/${comparisonId}/boats/${boatId}`, {
      method: 'DELETE',
      credentials: 'include',
    })
    if (!res.ok) throw new Error('Failed to remove boat')
    await fetchOne(comparisonId)
  }

  async function saveScores(comparisonId, boatId, scores) {
    const res = await fetch(`/api/comparisons/${comparisonId}/boats/${boatId}/scores`, {
      method: 'PUT',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ scores }),
    })
    if (!res.ok) {
      const data = await res.json()
      throw new Error(data.error || 'Failed to save scores')
    }
    // Refresh to get updated weighted scores from the server
    await fetchOne(comparisonId)
  }

  return {
    list, current, loading, error,
    fetchList, fetchOne, create, remove,
    addBoat, removeBoat, saveScores,
  }
})
