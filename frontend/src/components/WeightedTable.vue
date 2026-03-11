<template>
  <div class="weighted-table-wrapper">
    <table class="table weighted-score-table">
      <thead>
        <tr>
          <th style="text-align:left">Section</th>
          <th>Weight</th>
          <th v-for="r in results" :key="r.boat.id">{{ r.boat.name }}</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(section, si) in SECTIONS"
          :key="section.name"
          :class="{ 'section-informational': section.weight === 0 }"
        >
          <td><strong>{{ section.name }}</strong></td>
          <td class="weight-cell">
            {{ section.weight > 0 ? `${(section.weight * 100).toFixed(0)}%` : '—' }}
          </td>
          <td
            v-for="r in results"
            :key="r.boat.id"
            class="score-cell"
            :class="getSectionClass(r.sections[si])"
          >
            <template v-if="r.sections[si] && r.sections[si].scored_count > 0">
              <span class="weighted-score">{{ r.sections[si].weighted_score.toFixed(2) }}</span>
              <small class="raw-avg">avg {{ r.sections[si].raw_average.toFixed(1) }}</small>
              <small class="scored-count">
                {{ r.sections[si].scored_count }}/{{ r.sections[si].total_count }} scored
              </small>
            </template>
            <span v-else class="unscored">—</span>
          </td>
        </tr>
      </tbody>
      <tfoot>
        <tr class="total-row">
          <td colspan="2"><strong>Total Weighted Score</strong></td>
          <td v-for="r in results" :key="r.boat.id" class="total-cell">
            <strong>{{ r.total_weighted.toFixed(2) }}</strong>
            <small> / 10.00</small>
          </td>
        </tr>
      </tfoot>
    </table>
  </div>
</template>

<script setup>
import { SECTIONS } from '../scoring.js'

defineProps({
  results: { type: Array, required: true },
})

function getSectionClass(sr) {
  if (!sr || sr.scored_count === 0 || sr.weight === 0) return ''
  if (sr.raw_average <= 3) return 'score-red'
  if (sr.raw_average <= 5) return 'score-yellow'
  return 'score-green'
}
</script>
