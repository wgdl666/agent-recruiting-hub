<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import { fetchSkills, uploadFiles } from '../api/client'
import type { EvalSkill } from '../types'

const props = withDefaults(
  defineProps<{ batchId?: number; periodType?: string }>(),
  { batchId: 0, periodType: 'daily' },
)
const emit = defineEmits<{ done: [] }>()
const dragging = ref(false)
const loading = ref(false)
const skills = ref<EvalSkill[]>([])
// 故意不预填唯一选项：上传前必须点选评估岗位，红色 * 才有约束意义。
const skillId = ref('')

onMounted(async () => {
  try {
    skills.value = await fetchSkills()
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '加载评估岗位失败')
  }
})

function requireSkill(): boolean {
  if (skillId.value) return true
  ElMessage.warning('请先选择评估岗位标准')
  return false
}

async function handleFiles(fileList: File[]) {
  if (!fileList.length) return
  if (!requireSkill()) return
  loading.value = true
  try {
    const res = await uploadFiles(fileList, 'upload', {
      batchId: props.batchId,
      periodType: props.periodType,
      skillId: skillId.value,
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
    <el-form label-width="120px" class="skill-form" @submit.prevent>
      <el-form-item required>
        <template #label>评估岗位标准</template>
        <el-select
          v-model="skillId"
          placeholder="请选择评估岗位标准"
          class="skill-select"
        >
          <el-option
            v-for="s in skills"
            :key="s.id"
            :label="s.name"
            :value="s.id"
          />
        </el-select>
      </el-form-item>
    </el-form>
    <p class="skill-hint">对应招聘 Skill；当前仅「实习生」岗。未选择不能上传。</p>

    <div
      class="dropzone"
      :class="{ dragging, loading, disabled: !skillId }"
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
        :disabled="loading || !skillId"
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
  margin: -8px 0 12px 120px;
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
