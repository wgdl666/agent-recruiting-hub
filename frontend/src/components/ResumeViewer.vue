<script setup lang="ts">
defineProps<{ candidateId: number }>()
</script>

<template>
  <div class="viewer">
    <iframe :src="`/api/candidates/${candidateId}/resume`" title="resume" />
  </div>
</template>

<style scoped>
.viewer {
  height: calc(100vh - 220px);
  min-height: 480px;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  overflow: hidden;
}
iframe {
  width: 100%;
  height: 100%;
  border: none;
}
</style>
