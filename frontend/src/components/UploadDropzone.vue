<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import { fetchPositions, uploadFiles } from '../api/client'
import type { Position } from '../types'

const props = withDefaults(
  defineProps<{ batchId?: number; periodType?: string }>(),
  { batchId: 0, periodType: 'daily' },
)
const emit = defineEmits<{ done: [] }>()
const dragging = ref(false)
const loading = ref(false)
const positions = ref<Position[]>([])
// 故意不预填唯一选项：上传前必须点选招聘岗位，红色 * 才有约束意义。
const positionId = ref<number | ''>('')

const selected = computed(() => positions.value.find((p) => p.id === positionId.value))

onMounted(async () => {
  try {
    positions.value = await fetchPositions(true)
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '加载招聘岗位失败')
  }
})

function requirePosition(): boolean {
  if (positionId.value) return true
  ElMessage.warning('请先选择招聘岗位')
  return false
}

async function handleFiles(fileList: File[]) {
  if (!fileList.length) return
  if (!requirePosition()) return
  loading.value = true
  try {
    const res = await uploadFiles(fileList, 'upload', {
      batchId: props.batchId,
      periodType: props.periodType,
      positionId: Number(positionId.value),
    })
    ElMessage.success(`已评估 ${res.imported} 份简历`)
    if (res.errors?.length) {
      ElMessage.warning(`${res.errors.length} 个文件失败`)
    }
    emit('done')
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '上传失败')
  } finally {
    loading.value = false
  }
}

function onDrop(e: DragEvent) {
  e.preventDefault()
  dragging.value = false
  const files = Array.from(e.dataTransfer?.files || [])
  handleFiles(files)
}

function onChange(uploadFile: { raw?: File }) {
  if (uploadFile.raw) handleFiles([uploadFile.raw])
}
</script>

<template>
  <div class="upload-block">
    <el-form label-width="88px" class="skill-form" @submit.prevent>
      <el-form-item required>
        <template #label>招聘岗位</template>
        <el-select
          v-model="positionId"
          placeholder="请选择正在招聘的岗位"
          class="skill-select"
        >
          <el-option
            v-for="p in positions"
            :key="p.id"
            :label="p.name"
            :value="p.id"
          />
        </el-select>
      </el-form-item>
    </el-form>
    <p class="skill-hint">
      <template v-if="selected">
        检验标准：{{ selected.skill_name || selected.skill_id }}
        <span v-if="selected.description"> · {{ selected.description }}</span>
      </template>
      <template v-else>从左侧岗位阶梯中选择；选岗后自动套用该岗的检验标准（Skill）。未选择不能上传。</template>
    </p>

    <div
      class="dropzone"
      :class="{ dragging, loading, disabled: !positionId }"
      @dragover.prevent="dragging = true"
      @dragleave.prevent="dragging = false"
      @drop="onDrop"
    >
      <el-upload
        drag
        :auto-upload="false"
        :show-file-list="false"
        accept=".pdf,.zip"
        :on-change="onChange"
        :disabled="loading || !positionId"
      >
        <el-icon class="icon"><UploadFilled /></el-icon>
        <div class="title">拖入 PDF 简历或 ZIP 压缩包</div>
        <div class="hint">自动解压、文本提取、工程+Agent 启发式评分；图片 PDF 会尝试 OCR</div>
      </el-upload>
      <el-skeleton v-if="loading" animated :rows="2" style="margin-top: 12px" />
    </div>
  </div>
</template>

<style scoped>
.skill-form {
  margin-bottom: 0;
}
.skill-select {
  width: 280px;
}
.skill-hint {
  margin: -8px 0 12px 88px;
  font-size: 12px;
  color: #909399;
}
.dropzone {
  border-radius: 12px;
  padding: 8px;
  transition: background 0.2s;
}
.dropzone.dragging { background: #ecf5ff; }
.dropzone.disabled {
  opacity: 0.55;
}
.icon { font-size: 48px; color: #409eff; }
.title { font-size: 16px; margin-top: 8px; }
.hint { color: #909399; font-size: 13px; margin-top: 4px; }
</style>
