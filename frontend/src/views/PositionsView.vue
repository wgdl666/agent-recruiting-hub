<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createPosition, deletePosition, fetchSkills, updatePosition } from '../api/client'
import type { EvalSkill, Position } from '../types'
import { usePositions } from '../composables/usePositions'

const { allPositions, jdPositionId, loadPositions } = usePositions()
const skills = ref<EvalSkill[]>([])
const loading = ref(false)
const dialogOpen = ref(false)
const editing = ref<Position | null>(null)
const form = reactive({ name: '', skill_id: '', description: '', is_open: true })

const skillLabel = computed(() => {
  const map = new Map(skills.value.map((s) => [s.id, s]))
  return (id: string) => map.get(id)?.name || id
})

onMounted(async () => {
  loading.value = true
  try {
    skills.value = await fetchSkills()
    await loadPositions()
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '加载岗位失败')
  } finally {
    loading.value = false
  }
})

function openCreate() {
  editing.value = null
  form.name = ''
  form.skill_id = skills.value[0]?.id || ''
  form.description = ''
  form.is_open = true
  dialogOpen.value = true
}

function openEdit(p: Position) {
  editing.value = p
  form.name = p.name
  form.skill_id = p.skill_id
  form.description = p.description || ''
  form.is_open = p.is_open
  dialogOpen.value = true
}

async function save() {
  if (!form.name.trim()) {
    ElMessage.warning('请填写岗位名称')
    return
  }
  if (!form.skill_id) {
    ElMessage.warning('请选择检验标准')
    return
  }
  try {
    if (editing.value) {
      await updatePosition(editing.value.id, {
        name: form.name.trim(),
        skill_id: form.skill_id,
        description: form.description,
        is_open: form.is_open,
      })
      ElMessage.success('已更新岗位')
    } else {
      await createPosition({
        name: form.name.trim(),
        skill_id: form.skill_id,
        description: form.description,
        is_open: form.is_open,
      })
      ElMessage.success('已新增岗位')
    }
    dialogOpen.value = false
    await loadPositions()
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  }
}

function openJd(p: Position) {
  // 只改共享的 jdPositionId，由外壳右侧抽屉承接，避免把岗位列表整页换成 JD。
  jdPositionId.value = p.id
}

async function toggleOpen(p: Position) {
  try {
    await updatePosition(p.id, { is_open: !p.is_open })
    await loadPositions()
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '更新失败')
  }
}

async function remove(p: Position) {
  const n = p.candidate_count ?? 0
  const extra = n > 0 ? `该岗下有 ${n} 位候选人，删岗后他们会从该岗解绑，候选人记录本身保留。` : '此操作不可恢复。'
  try {
    await ElMessageBox.confirm(`${extra}`, `删除岗位「${p.name}」？`, {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning',
      confirmButtonClass: 'el-button--danger',
    })
  } catch {
    return
  }
  try {
    await deletePosition(p.id)
    ElMessage.success('已删除岗位')
    await loadPositions()
  } catch (e: unknown) {
    ElMessage.error(e instanceof Error ? e.message : '删除失败')
  }
}
</script>

<template>
  <div class="positions-page">
    <div class="head">
      <div>
        <h2>岗位阶梯</h2>
        <p class="sub">点「查看 JD」或侧栏岗位名，在右侧打开详情。上传时选岗会套用对应检验标准。</p>
      </div>
      <el-button type="primary" @click="openCreate">新增岗位</el-button>
    </div>

    <el-table :data="allPositions" v-loading="loading" stripe size="small" :header-cell-style="{ background: '#fafafa' }">
      <el-table-column label="岗位" min-width="140">
        <template #default="{ row }">
          <el-button link type="primary" class="pos-name" @click="openJd(row)">{{ row.name }}</el-button>
        </template>
      </el-table-column>
      <el-table-column label="检验标准" min-width="160">
        <template #default="{ row }">
          <el-tag size="small" type="info">{{ row.skill_name || skillLabel(row.skill_id) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="招聘中" width="100">
        <template #default="{ row }">
          <el-tag :type="row.is_open ? 'success' : 'info'" size="small">
            {{ row.is_open ? '在招' : '已停' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="候选人" width="88">
        <template #default="{ row }">{{ row.candidate_count ?? 0 }}</template>
      </el-table-column>
      <el-table-column label="操作" width="280">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="openJd(row)">查看 JD</el-button>
          <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
          <el-button link size="small" @click="toggleOpen(row)">
            {{ row.is_open ? '停止招聘' : '重新开放' }}
          </el-button>
          <el-button link type="danger" size="small" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialogOpen" :title="editing ? '编辑岗位' : '新增岗位'" width="480px">
      <el-form label-width="96px" @submit.prevent="save">
        <el-form-item label="岗位名称" required>
          <el-input v-model="form.name" placeholder="如：实习生" />
        </el-form-item>
        <el-form-item label="检验标准" required>
          <el-select v-model="form.skill_id" placeholder="选择 Skill" class="full">
            <el-option v-for="s in skills" :key="s.id" :label="s.name" :value="s.id">
              <span>{{ s.name }}</span>
              <span v-if="s.description" class="opt-desc">{{ s.description }}</span>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="可选" />
        </el-form-item>
        <el-form-item label="招聘中">
          <el-switch v-model="form.is_open" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogOpen = false">取消</el-button>
        <el-button type="primary" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.positions-page { max-width: 960px; }
.head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 16px;
}
h2 { margin: 0 0 6px; font-size: 18px; }
.sub { margin: 0; font-size: 13px; color: #909399; line-height: 1.5; }
.pos-name { font-weight: 600; padding: 0; }
.full { width: 100%; }
.opt-desc {
  margin-left: 8px;
  color: #909399;
  font-size: 12px;
}
</style>
