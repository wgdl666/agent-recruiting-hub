<script setup lang="ts">
import type { InterviewQuestion } from '../types'

defineProps<{ questions: InterviewQuestion[] }>()
</script>

<template>
  <el-card shadow="never" class="panel">
    <template #header>
      <span>定制面试题（L1→L2→L3 深挖）</span>
    </template>
    <el-empty v-if="!questions.length" description="暂无定制题，S 档候选人会自动生成" />
    <ol v-else class="list">
      <li v-for="q in questions" :key="q.id">{{ q.question }}</li>
    </ol>
  </el-card>
</template>

<style scoped>
.panel { margin-top: 16px; }
.list { padding-left: 20px; line-height: 1.8; }
.list li { margin-bottom: 12px; }
</style>
