<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Close } from '@element-plus/icons-vue'

const STORAGE_KEY = 'hub-workflow-guide-dismissed'
const visible = ref(true)

onMounted(() => {
  visible.value = localStorage.getItem(STORAGE_KEY) !== '1'
})

function dismiss() {
  visible.value = false
  localStorage.setItem(STORAGE_KEY, '1')
}
</script>

<template>
  <el-alert v-if="visible" type="info" class="workflow" :closable="false">
    <template #title>
      <span class="title-row">
        <span>日常流程</span>
        <el-button link type="primary" :icon="Close" @click="dismiss">不再提示</el-button>
      </span>
    </template>
    <ol class="steps">
      <li><strong>选批次</strong>：顶部选择「今日/本周/本月」批次，或查看全部；新上传自动归入当前周期</li>
      <li><strong>上传</strong>：拖入 PDF/ZIP，或点「上传简历」标签</li>
      <li><strong>推进</strong>：列表改状态，或切到「看板」拖拽卡片（简历筛选 → 待约面 → 面试中 → 通过/完成/淘汰）</li>
      <li><strong>面试</strong>：点击行打开详情；S 档可看定制题</li>
    </ol>
  </el-alert>
</template>

<style scoped>
.workflow { margin-bottom: 12px; }
.title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}
.steps {
  margin: 8px 0 0;
  padding-left: 20px;
  line-height: 1.8;
  font-size: 13px;
}
</style>
