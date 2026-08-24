<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import { uploadFiles } from '../api/client'

const props = withDefaults(
  defineProps<{ batchId?: number; periodType?: string }>(),
  { batchId: 0, periodType: 'daily' },
)
const emit = defineEmits<{ done: [] }>()
const dragging = ref(false)
const loading = ref(false)

async function handleFiles(fileList: File[]) {
  if (!fileList.length) return
  loading.value = true
  try {
    const res = await uploadFiles(fileList, 'upload', {
      batchId: props.batchId,
      periodType: props.periodType,
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
  <div
    class="dropzone"
    :class="{ dragging, loading }"
    @dragover.prevent="dragging = true"
    @dragleave.prevent="dragging = false"
    @drop="onDrop"
  >
    <el-upload drag :auto-upload="false" :show-file-list="false" accept=".pdf,.zip" :on-change="onChange" :disabled="loading">
      <el-icon class="icon"><UploadFilled /></el-icon>
      <div class="title">拖入 PDF 简历或 ZIP 压缩包</div>
      <div class="hint">自动解压、文本提取、工程+Agent 启发式评分；图片 PDF 会尝试 OCR</div>
    </el-upload>
    <el-skeleton v-if="loading" animated :rows="2" style="margin-top: 12px" />
  </div>
</template>

<style scoped>
.dropzone {
  border-radius: 12px;
  padding: 8px;
  transition: background 0.2s;
}
.dropzone.dragging { background: #ecf5ff; }
.icon { font-size: 48px; color: #409eff; }
.title { font-size: 16px; margin-top: 8px; }
.hint { color: #909399; font-size: 13px; margin-top: 4px; }
</style>
