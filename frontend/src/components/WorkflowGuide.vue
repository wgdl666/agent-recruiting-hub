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
      <li><strong>列表筛人</strong>：左侧「列表」按状态/档位翻人，点行打开详情</li>
      <li><strong>看板跟进</strong>：左侧「看板」选今日/本周/本月批次，看漏斗并拖拽改状态</li>
      <li><strong>上传</strong>：左侧「上传」，或列表里快速拖入 PDF/ZIP</li>
      <li><strong>知识库</strong>：开发、部署、API 文档 — 见左侧「知识库」</li>
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
