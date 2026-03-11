<template>
  <div class="share-view">
    <div class="share-header">
      <h1>⚓ svCompare</h1>
      <p class="subtitle">Sailboat Comparison — Read Only</p>
    </div>

    <div v-if="loading" class="empty-state">Loading…</div>
    <div v-else-if="error" class="error-msg">{{ error }}</div>

    <template v-else-if="comparison">
      <h2>{{ comparison.name }}</h2>
      <p class="share-meta">Expires {{ formatDate(comparison.expires_at) }}</p>

      <div class="scoring-legend">
        Scores are on a scale of
        <span class="legend-bad">1 = Bad</span>
        to
        <span class="legend-good">10 = Excellent</span>.
        Blank means not scored. Hover <span class="legend-icon">?</span> for scoring guidance.
      </div>

      <div v-if="comparison.results.length === 0" class="empty-state">
        No boats have been added to this comparison yet.
      </div>

      <template v-else>
        <div class="scoring-grid-wrapper">
          <table class="table scoring-grid">
            <thead>
              <tr>
                <th class="question-col">Question</th>
                <th
                  v-for="r in comparison.results"
                  :key="r.boat.id"
                  class="boat-score-col"
                >
                  {{ r.boat.name }}
                </th>
              </tr>
            </thead>
            <tbody>
              <template v-for="section in SECTIONS" :key="section.name">
                <tr class="section-header-row">
                  <td :colspan="comparison.results.length + 1">
                    <strong>{{ section.name }}</strong>
                    <span class="section-weight">
                      {{ section.weight > 0
                        ? `${(section.weight * 100).toFixed(0)}% weighted`
                        : 'Informational only' }}
                    </span>
                  </td>
                </tr>
                <tr v-for="question in section.questions" :key="question.id">
                  <td class="question-text">
                    <span class="q-num">Q{{ question.id }}</span>
                    {{ question.text }}
                    <TooltipIcon :text="question.tooltip" />
                  </td>
                  <td
                    v-for="r in comparison.results"
                    :key="r.boat.id"
                    class="score-input-cell"
                    :class="getScoreClass(getScore(r.boat, question.id))"
                  >
                    <span class="score-display">
                      {{ getScore(r.boat, question.id) ?? '—' }}
                    </span>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>

        <div class="card results-card">
          <h3>Weighted Score Summary</h3>
          <WeightedTable :results="comparison.results" />
          <div class="pdf-actions">
            <button class="btn btn-secondary" @click="() => window.print()">⬇ Download PDF</button>
          </div>
        </div>
      </template>

      <div class="share-cta">
        <p>Want to compare your own boats?</p>
        <RouterLink to="/login" class="btn btn-primary">Get Started →</RouterLink>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import WeightedTable from '../components/WeightedTable.vue'
import TooltipIcon from '../components/TooltipIcon.vue'
import { SECTIONS } from '../scoring.js'

const route = useRoute()
const loading = ref(true)
const error = ref('')
const comparison = ref(null)

onMounted(async () => {
  try {
    const res = await fetch(`/api/share/${route.params.token}`)
    if (!res.ok) {
      error.value = 'Comparison not found or has expired.'
      return
    }
    comparison.value = await res.json()
  } catch {
    error.value = 'Failed to load comparison.'
  } finally {
    loading.value = false
  }
})

function getScore(boat, questionId) {
  const s = (boat.scores || []).find(s => s.question_id === questionId)
  return s?.value ?? null
}

function getScoreClass(value) {
  if (value === null || value === undefined) return ''
  if (value <= 3) return 'score-red'
  if (value <= 5) return 'score-yellow'
  return ''
}

function formatDate(str) {
  return new Date(str).toLocaleDateString()
}
</script>
