<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { updateCandidate } from '../api/client'
import type { Candidate, CandidateDetail } from '../types'
import TierTag from './TierTag.vue'

const props = defineProps<{ candidate: Candidate; compact?: boolean }>()
const emit = defineEmits<{ updated: [CandidateDetail] }>()

const tiers = ['S', 'A', '淘汰']
const tier = ref(props.candidate.tier)
const saving = ref(false)

watch(() => props.candidate.tier, (v) => { tier.value = v })

async function onChange(val: string) {
  if (val === props.candidate.tier) return
  saving.value = true
  try {
    const updated = await updateCandidate(props.candidate.id, { tier: val })
    tier.value = updated.tier
    ElMessage.success(`${props.candidate.name} → ${val} 档`)
    emit('updated', updated)
  } catch {
    tier.value = props.candidate.tier
    ElMessage.error('档位更新失败')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="tier-editor" @click.stop>
    <el-select
      v-if="!compact"
      v-model="tier"
      size="small"
      :disabled="saving"
      style="width: 88px"
      @change="onChange"
    >
      <el-option v-for="t in tiers" :key="t" :label="t" :value="t" />
    </el-select>
    <el-dropdown v-else trigger="click" @command="onChange">
      <span class="tier-trigger">
        <TierTag :tier="tier" />
        <span v-if="candidate.tier_manual" class="lock" title="仅锁定档位；上传/重评仍会更新自动分">🔒</span>
      </span>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item v-for="t in tiers" :key="t" :command="t">{{ t }}</el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
  </div>
</template>

<style scoped>
.tier-editor { display: inline-flex; align-items: center; }
.tier-trigger { display: inline-flex; align-items: center; gap: 4px; cursor: pointer; }
.lock { font-size: 10px; opacity: 0.7; }
</style>
