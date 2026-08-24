<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, QuestionFilled, Refresh } from '@element-plus/icons-vue'
import {
  importSeed, exportAll, exportFeishu, exportMarkdown, rescreenAll,
} from '../api/client'

const emit = defineEmits<{ refresh: [] }>()

const busy = ref(false)

async function withBusy(fn: () => Promise<void>) {
  if (busy.value) return
  busy.value = true
  try {
    await fn()
  } finally {
    busy.value = false
  }
}

async function onSeed() {
  await ElMessageBox.confirm(
    '将用 seed/batch.json 覆盖/补全候选人元数据，并关联 seed/resumes 下的 PDF。适合首次初始化，日常一般不需要。',
    '导入种子数据',
    { confirmButtonText: '导入', cancelButtonText: '取消', type: 'warning' },
  )
  await withBusy(async () => {
    const res = await importSeed()
    ElMessage.success(`已导入 ${res.imported} 条`)
    emit('refresh')
  })
}

async function onRescreen() {
  await ElMessageBox.confirm(
    '对已存 PDF 重新自动评分。你手动调整过的档位不会被覆盖；图片 PDF 可能仍是「待人工」。',
    '重评全部简历',
    { confirmButtonText: '开始重评', cancelButtonText: '取消' },
  )
  await withBusy(async () => {
    const res = await rescreenAll()
    ElMessage.success(`已重评 ${res.updated} 份`)
    emit('refresh')
  })
}

function onExportMd() {
  exportMarkdown()
  ElMessage.success('Markdown 报告已下载')
}

async function onFeishu() {
  await withBusy(async () => {
    const res = await exportFeishu()
    if (res.ok && res.lark?.data?.document?.url) {
      ElMessage.success('飞书文档已创建')
      window.open(res.lark.data.document.url, '_blank')
    } else {
      ElMessage.warning('飞书同步失败，请检查 lark-cli 登录；本地 data/export.md 可能已生成')
    }
  })
}

function onExportJson() {
  exportAll()
  ElMessage.success('JSON 已下载（可供 Cursor 读取）')
}

const helpItems = [
  { title: '导入种子', desc: '首次使用或库清空后，恢复 47 人预置筛选结果 + 简历' },
  { title: '重评简历', desc: '批量刷新自动分；手动档位不受影响' },
  { title: '导出 / 同步飞书', desc: '生成筛选报告，用于汇报或存档' },
  { title: '导出 JSON', desc: '全量数据导出，给 Cursor 或脚本用' },
]
function onCommand(cmd: string) {
  const map: Record<string, () => void> = {
    seed: onSeed,
    rescreen: onRescreen,
    md: onExportMd,
    feishu: onFeishu,
    json: onExportJson,
  }
  map[cmd]?.()
}
</script>

<template>
  <div class="toolbar-actions">
    <el-tooltip content="刷新列表">
      <el-button :icon="Refresh" circle :loading="busy" @click="emit('refresh')" />
    </el-tooltip>

    <el-popover placement="bottom-end" :width="320" trigger="click">
      <template #reference>
        <el-button :icon="QuestionFilled" circle />
      </template>
      <div class="help">
        <div class="help-title">管理工具说明</div>
        <div v-for="item in helpItems" :key="item.title" class="help-item">
          <strong>{{ item.title }}</strong>
          <span>{{ item.desc }}</span>
        </div>
        <div class="help-foot">日常只需：上传 → 点行开标签 → 改状态/档位</div>
      </div>
    </el-popover>

    <el-dropdown trigger="click" @command="onCommand">
      <el-button :loading="busy">
        数据与导出 <el-icon class="el-icon--right"><ArrowDown /></el-icon>
      </el-button>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item disabled class="menu-hint">初始化 / 维护</el-dropdown-item>
          <el-dropdown-item command="seed">导入种子数据…</el-dropdown-item>
          <el-dropdown-item command="rescreen">重评全部简历…</el-dropdown-item>
          <el-dropdown-item divided disabled class="menu-hint">导出</el-dropdown-item>
          <el-dropdown-item command="md">下载 Markdown 报告</el-dropdown-item>
          <el-dropdown-item command="feishu">同步到飞书文档</el-dropdown-item>
          <el-dropdown-item command="json">下载 JSON（Cursor）</el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
  </div>
</template>

<style scoped>
.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
}
.help-title { font-weight: 600; margin-bottom: 8px; }
.help-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin-bottom: 10px;
  font-size: 13px;
}
.help-item span { color: #909399; }
.help-foot {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid #ebeef5;
  font-size: 12px;
  color: #606266;
}
:deep(.menu-hint) {
  font-size: 12px;
  color: #909399 !important;
  cursor: default !important;
}
</style>
