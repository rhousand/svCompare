<template>
  <div class="share-link-box">
    <label>Share link (read-only · expires {{ formatDate(expiresAt) }})</label>
    <div class="share-link-row">
      <input type="text" :value="shareUrl" readonly />
      <button class="btn btn-secondary btn-sm" @click="copy">
        {{ copied ? 'Copied!' : 'Copy' }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  shareToken: { type: String, required: true },
  expiresAt:  { type: String, required: true },
})

const copied = ref(false)

const shareUrl = computed(() => `${window.location.origin}/share/${props.shareToken}`)

async function copy() {
  await navigator.clipboard.writeText(shareUrl.value)
  copied.value = true
  setTimeout(() => { copied.value = false }, 2000)
}

function formatDate(str) {
  return new Date(str).toLocaleDateString()
}
</script>
