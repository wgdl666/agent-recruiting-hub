<script setup lang="ts">
import { computed } from 'vue'
import type { PipelineStats } from '../../types'
import { PIPELINE_ORDER, statusLabel } from '../../constants/status'

const props = defineProps<{ stats: PipelineStats | null; loading?: boolean }>()

const columns = computed(() =>
  PIPELINE_ORDER.map((status) => ({
    status,
    label: statusLabel(status),
    count: props.stats?.by_status?.[status] ?? 0,
  })),
)
</script>

<template>
  <a-spin :spinning="loading">
    <div class="funnel">
      <a-statistic
        title="本批合计"
        :value="stats?.total ?? 0"
        class="total"
      />
      <a-divider type="vertical" style="height: 48px" />
      <div class="steps">
        <div v-for="col in columns" :key="col.status" class="step">
          <div class="step-count">{{ col.count }}</div>
          <div class="step-label">{{ col.label }}</div>
        </div>
      </div>
      <a-divider type="vertical" style="height: 48px" />
      <div class="tiers">
        <a-tag color="red">S {{ stats?.by_tier?.S ?? 0 }}</a-tag>
        <a-tag color="orange">A {{ stats?.by_tier?.A ?? 0 }}</a-tag>
        <a-tag>淘汰 {{ stats?.by_tier?.淘汰 ?? 0 }}</a-tag>
      </div>
    </div>
  </a-spin>
</template>

<style scoped>
.funnel {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
  padding: 10px 12px;
  background: #fafafa;
  border-radius: 8px;
  margin-bottom: 12px;
}
.total { min-width: 80px; }
.steps {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  flex: 1;
}
.step {
  text-align: center;
  min-width: 56px;
}
.step-count {
  font-size: 20px;
  font-weight: 600;
  line-height: 1.2;
}
.step-label {
  font-size: 12px;
  color: #666;
  margin-top: 2px;
}
.tiers {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}
</style>
