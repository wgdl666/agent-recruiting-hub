<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { updateCandidate } from '../api/client'
import type { Candidate, CandidateDetail } from '../types'
import { STATUS_OPTIONS, statusLabel, statusType } from '../constants/status'

const props = defineProps<{ candidate: Pick<Candidate, 'id' | 'status'>; compact?: boolean }>()
const emit = defineEmits<{ updated: [CandidateDetail] }>()

const value = ref(props.candidate.status || 'screening')
const saving = ref(false)

watch(() => props.candidate.status, (v) => { value.value = v || 'screening' })

async function onChange(val: string) {
  if (val === props.candidate.status) return
  saving.value = true
  try {
    const updated = await updateCandidate(props.candidate.id, { status: val })
    value.value = updated.status
    ElMessage.success(`状态 → ${statusLabel(val)}`)
    emit('updated', updated)
  } catch {
    value.value = props.candidate.status
    ElMessage.error('状态更新失败')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="status-editor" @click.stop>
    <el-select
      v-if="!compact"
      v-model="value"
      size="small"
      :disabled="saving"
      style="width: 110px"
      @change="onChange"
    >
      <el-option v-for="o in STATUS_OPTIONS" :key="o.value" :label="o.label" :value="o.value" />
    </el-select>
    <el-dropdown v-else trigger="click" @command="onChange">
      <el-tag size="small" :type="statusType(value)" class="status-tag">{{ statusLabel(value) }}</el-tag>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item v-for="o in STATUS_OPTIONS" :key="o.value" :command="o.value">
            {{ o.label }}
          </el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
  </div>
</template>

<style scoped>
.status-editor { display: inline-flex; }
.status-tag { cursor: pointer; }
</style>
