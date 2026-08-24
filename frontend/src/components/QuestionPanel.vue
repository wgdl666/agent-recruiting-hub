<script setup lang="ts">
import type { InterviewQuestion } from '../types'

defineProps<{ questions: InterviewQuestion[] }>()

function levelColor(level?: string) {
  if (level === 'L1') return 'blue'
  if (level === 'L2') return 'orange'
  if (level === 'L3') return 'red'
  return 'default'
}
</script>

<template>
  <el-card shadow="never" class="panel">
    <template #header>
      <span>定制面试题（L1→L2→L3 深挖）</span>
    </template>
    <el-empty v-if="!questions.length" description="暂无定制题，S 档候选人会自动生成" />
    <el-collapse v-else class="qa-list">
      <el-collapse-item v-for="q in questions" :key="q.id" :name="q.id">
        <template #title>
          <div class="qa-title">
            <a-tag v-if="q.level" :color="levelColor(q.level)" size="small">{{ q.level }}</a-tag>
            <span>{{ q.question }}</span>
          </div>
        </template>
        <div v-if="q.answer" class="answer">
          <div class="answer-label">参考答案（面试官）</div>
          <p>{{ q.answer }}</p>
        </div>
        <el-text v-else type="info" size="small">暂无参考答案</el-text>
      </el-collapse-item>
    </el-collapse>
  </el-card>
</template>

<style scoped>
.panel { margin-top: 16px; }
.qa-list { border: none; }
.qa-title {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  line-height: 1.5;
  padding-right: 8px;
}
.answer {
  padding: 4px 0 8px;
  font-size: 13px;
  line-height: 1.7;
  color: #606266;
}
.answer-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}
:deep(.el-collapse-item__header) {
  height: auto;
  min-height: 44px;
  line-height: 1.5;
  padding: 8px 0;
}
</style>
