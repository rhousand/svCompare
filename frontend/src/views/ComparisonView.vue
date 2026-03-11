<template>
  <div class="comparison-view">
    <div v-if="store.loading && !store.current" class="empty-state">Loading…</div>
    <div v-else-if="store.error" class="error-msg">{{ store.error }}</div>

    <template v-else-if="store.current">
      <!-- Header -->
      <div class="page-header">
        <div>
          <RouterLink to="/dashboard" class="back-link">← Dashboard</RouterLink>
          <h2>{{ store.current.name }}</h2>
        </div>
        <ShareLink
          :shareToken="store.current.share_token"
          :expiresAt="store.current.expires_at"
        />
      </div>

      <!-- Boat management -->
      <div class="card boat-manager">
        <div class="boat-tabs">
          <span v-for="r in store.current.results" :key="r.boat.id" class="boat-tab">
            {{ r.boat.name }}
            <button class="remove-boat" @click="handleRemoveBoat(r.boat.id)" title="Remove boat">×</button>
          </span>

          <form
            v-if="store.current.results.length < 5"
            class="add-boat-form"
            @submit.prevent="handleAddBoat"
          >
            <input v-model="newBoatName" type="text" placeholder="+ Add boat name" />
            <button type="submit" class="btn btn-primary btn-sm" :disabled="!newBoatName.trim()">
              Add
            </button>
          </form>
          <span v-else class="max-boats-note">Maximum 5 boats</span>
        </div>
        <div v-if="boatError" class="error-msg" style="margin-top:.5rem">{{ boatError }}</div>
      </div>

      <!-- Scoring grid -->
      <template v-if="store.current.results.length > 0">
        <div class="scoring-legend">
          Score each question on a scale of
          <span class="legend-bad">1 = Bad</span>
          to
          <span class="legend-good">10 = Excellent</span>.
          Leave blank if unknown. Hover <span class="legend-icon">?</span> for scoring guidance.
        </div>
        <div class="scoring-grid-wrapper">
          <table class="table scoring-grid">
            <thead>
              <tr>
                <th class="question-col">Question</th>
                <th
                  v-for="r in store.current.results"
                  :key="r.boat.id"
                  class="boat-score-col"
                >
                  {{ r.boat.name }}
                  <div v-if="savingBoats[r.boat.id]" class="saving-indicator">saving…</div>
                </th>
              </tr>
            </thead>
            <tbody>
              <template v-for="section in SECTIONS" :key="section.name">
                <tr class="section-header-row">
                  <td :colspan="store.current.results.length + 1">
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
                    v-for="r in store.current.results"
                    :key="r.boat.id"
                    class="score-input-cell"
                    :class="getScoreClass(localScores[r.boat.id]?.[question.id])"
                  >
                    <input
                      type="number"
                      min="1"
                      max="10"
                      class="score-input"
                      :class="{ 'score-invalid': isInvalid(r.boat.id, question.id) }"
                      :value="localScores[r.boat.id]?.[question.id] ?? ''"
                      placeholder="—"
                      @input="onScoreInput(r.boat.id, question.id, $event.target.value)"
                      @blur="validateOnBlur(r.boat.id, question.id)"
                    />
                    <span
                      v-if="isInvalid(r.boat.id, question.id)"
                      class="score-error-tip"
                    >1–10</span>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>

        <!-- Weighted summary -->
        <div class="card results-card">
          <h3>Weighted Score Summary</h3>
          <WeightedTable :results="store.current.results" />
          <div class="pdf-actions">
            <button class="btn btn-secondary" @click="downloadPDF">⬇ Download PDF</button>
          </div>
        </div>
      </template>

      <div v-else class="empty-state">Add at least one boat to start scoring.</div>
    </template>
  </div>
</template>

<script setup>
import { ref, reactive, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useComparisonsStore } from '../stores/comparisons.js'
import ShareLink from '../components/ShareLink.vue'
import WeightedTable from '../components/WeightedTable.vue'
import TooltipIcon from '../components/TooltipIcon.vue'
import { SECTIONS, buildScoreMap } from '../scoring.js'

const route = useRoute()
const store = useComparisonsStore()

const newBoatName = ref('')
const boatError = ref('')

// localScores[boatId][questionId] = value (number | null)
const localScores = reactive({})
// Per-boat debounce timers
const debounceTimers = {}
// Per-boat saving indicator
const savingBoats = reactive({})
// invalidScores[boatId][questionId] = true when value is out of 1-10 range
const invalidScores = reactive({})

onMounted(async () => {
  await store.fetchOne(route.params.id)
  syncLocalScores()
})

watch(
  () => store.current?.results,
  () => { syncLocalScores() },
  { deep: false },
)

function syncLocalScores() {
  if (!store.current?.results) return
  for (const r of store.current.results) {
    if (!localScores[r.boat.id]) {
      localScores[r.boat.id] = buildScoreMap(r.boat.scores)
    }
  }
}

function isInvalid(boatId, questionId) {
  return !!(invalidScores[boatId]?.[questionId])
}

function validateOnBlur(boatId, questionId) {
  const v = localScores[boatId]?.[questionId]
  if (v !== null && v !== undefined && (v < 1 || v > 10)) {
    if (!invalidScores[boatId]) invalidScores[boatId] = {}
    invalidScores[boatId][questionId] = true
  }
}

function onScoreInput(boatId, questionId, rawValue) {
  const v = rawValue === '' ? null : Number(rawValue)
  if (!localScores[boatId]) localScores[boatId] = {}
  localScores[boatId][questionId] = v

  // Clear any existing invalid flag for this cell
  if (invalidScores[boatId]) {
    delete invalidScores[boatId][questionId]
  }

  // Only save if in-range or empty
  if (v !== null && (v < 1 || v > 10)) {
    if (!invalidScores[boatId]) invalidScores[boatId] = {}
    invalidScores[boatId][questionId] = true
    return
  }

  clearTimeout(debounceTimers[boatId])
  debounceTimers[boatId] = setTimeout(() => saveBoatScores(boatId), 800)
}

async function saveBoatScores(boatId) {
  const map = localScores[boatId] || {}
  const scores = Object.entries(map)
    .filter(([, v]) => v !== null && !isNaN(v) && v >= 1 && v <= 10)
    .map(([qid, v]) => ({ question_id: Number(qid), value: v, notes: '' }))

  savingBoats[boatId] = true
  try {
    await store.saveScores(route.params.id, boatId, scores)
  } catch (e) {
    console.error('Score save failed:', e.message)
  } finally {
    savingBoats[boatId] = false
  }
}

function getScoreClass(value) {
  if (value === null || value === undefined || value === '') return ''
  if (value <= 3) return 'score-red'
  if (value <= 5) return 'score-yellow'
  return ''
}

async function handleAddBoat() {
  boatError.value = ''
  const name = newBoatName.value.trim()
  if (!name) return
  try {
    await store.addBoat(route.params.id, name)
    newBoatName.value = ''
  } catch (e) {
    boatError.value = e.message
  }
}

function downloadPDF() {
  window.print()
}

async function handleRemoveBoat(boatId) {
  if (!confirm('Remove this boat and all its scores?')) return
  try {
    await store.removeBoat(route.params.id, boatId)
    delete localScores[boatId]
    delete invalidScores[boatId]
  } catch (e) {
    boatError.value = e.message
  }
}
</script>
