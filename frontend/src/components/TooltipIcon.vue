<template>
  <span
    class="tooltip-wrap"
    ref="triggerEl"
    @mouseenter="show"
    @mouseleave="hide"
    @focusin="show"
    @focusout="hide"
    tabindex="0"
    role="button"
    aria-label="Scoring guidance"
  >
    <span class="tooltip-icon">?</span>

    <Teleport to="body">
      <div
        v-if="visible"
        class="tooltip-box"
        :style="style"
        role="tooltip"
      >
        <p v-for="(line, i) in lines" :key="i" :class="lineClass(line)">{{ line }}</p>
      </div>
    </Teleport>
  </span>
</template>

<script setup>
import { ref, computed, nextTick } from 'vue'

const props = defineProps({
  text: { type: String, required: true },
})

const triggerEl = ref(null)
const visible = ref(false)
const style = ref({})

// Split the tooltip text on newlines for formatted display
const lines = computed(() => props.text.split('\n').filter(l => l.trim()))

function lineClass(line) {
  if (line.startsWith('Good:') || line.startsWith('Good sign:')) return 'tip-good'
  if (line.startsWith('Avoid:') || line.startsWith('Red flag:')) return 'tip-avoid'
  if (line.startsWith('Caution:')) return 'tip-caution'
  if (line.startsWith('Strategy:') || line.startsWith('Important:')) return 'tip-info'
  return 'tip-neutral'
}

async function show() {
  visible.value = true
  await nextTick()
  if (!triggerEl.value) return

  const rect = triggerEl.value.getBoundingClientRect()
  const vw = window.innerWidth
  const left = Math.min(rect.left, vw - 320)

  style.value = {
    position: 'fixed',
    top: `${rect.bottom + 8}px`,
    left: `${Math.max(8, left)}px`,
    zIndex: 9999,
  }
}

function hide() {
  visible.value = false
}
</script>
