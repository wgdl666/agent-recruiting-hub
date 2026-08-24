<script setup lang="ts">
import { ArrowLeft } from '@element-plus/icons-vue'
import UploadDropzone from '../components/UploadDropzone.vue'
import BatchSelector from '../components/pipeline/BatchSelector.vue'
import { useTabs } from '../composables/useTabs'
import { useActiveBatch } from '../composables/useActiveBatch'

const { switchTab } = useTabs()
const { activeBatchId, uploadPeriodType } = useActiveBatch()

function onDone() {
  switchTab('home')
}
</script>

<template>
  <div class="upload-page">
    <el-button :icon="ArrowLeft" link @click="switchTab('home')">返回候选人列表</el-button>
    <a-card :bordered="false" class="batch-card">
      <BatchSelector />
    </a-card>
    <el-card shadow="never" class="card">
      <template #header>
        <div class="header">
          <span>上传并评估简历</span>
          <el-tag type="info" size="small">新候选人默认进入「简历筛选」，并归入当前周期批次</el-tag>
        </div>
      </template>
      <UploadDropzone
        :batch-id="activeBatchId"
        :period-type="uploadPeriodType"
        @done="onDone"
      />
      <el-divider />
      <div class="tips">
        <p><strong>支持：</strong>单个 PDF，或 ZIP（自动解压其中的 PDF）</p>
        <p><strong>评估后：</strong>回到列表按档位/状态筛选，点击行在新标签打开详情</p>
        <p><strong>图片 PDF：</strong>自动分可能不准，以人工档位和简历内容为准</p>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.upload-page { max-width: 720px; }
.batch-card { margin-top: 8px; margin-bottom: 8px; }
.card { margin-top: 8px; }
.header {
  display: flex;
  align-items: center;
  gap: 12px;
}
.tips {
  font-size: 13px;
  color: #606266;
  line-height: 1.8;
}
.tips p { margin: 0 0 4px; }
</style>
