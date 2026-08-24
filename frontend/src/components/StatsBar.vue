<script setup lang="ts">
import type { Stats } from '../types'

defineProps<{ stats: Stats | null }>()
</script>

<template>
  <div v-if="stats">
    <el-row :gutter="12" class="stats">
      <el-col :span="4"><el-statistic title="总计" :value="stats.total" /></el-col>
      <el-col :span="4"><el-statistic title="简历筛选" :value="stats.by_status?.screening || 0" /></el-col>
      <el-col :span="4"><el-statistic title="面试中" :value="stats.by_status?.interviewing || 0" /></el-col>
      <el-col :span="3"><el-statistic title="待约面" :value="stats.by_status?.to_interview || 0" /></el-col>
      <el-col :span="3"><el-statistic title="通过" :value="stats.by_status?.passed || 0" /></el-col>
      <el-col :span="3"><el-statistic title="淘汰" :value="stats.by_status?.rejected || 0" /></el-col>
      <el-col :span="3"><el-statistic title="面试完成" :value="stats.by_status?.completed || 0" /></el-col>
      <el-col :span="3"><el-statistic title="S 档" :value="stats.by_tier.S || 0" /></el-col>
      <el-col :span="3"><el-statistic title="A 档" :value="stats.by_tier.A || 0" /></el-col>
      <el-col :span="3"><el-statistic title="档位淘汰" :value="stats.by_tier['淘汰'] || 0" /></el-col>
    </el-row>
  </div>
</template>

<style scoped>
.stats { margin-bottom: 16px; }
</style>
