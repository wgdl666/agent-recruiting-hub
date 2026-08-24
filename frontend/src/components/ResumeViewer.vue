<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ candidateId: number }>()

// 隐藏浏览器 PDF 内置缩略图侧栏，默认横向铺满
const resumeSrc = computed(
  () => `/api/candidates/${props.candidateId}/resume#navpanes=0&view=FitH`,
)
</script>

<template>
  <div class="viewer">
    <iframe :src="resumeSrc" title="resume" />
  </div>
</template>

<style scoped>
.viewer {
  height: calc(100vh - 200px);
  min-height: 560px;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  overflow: hidden;
  background: #525659;
}
iframe {
  width: 100%;
  height: 100%;
  border: none;
}
</style>
