<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { BookOutlined, GithubOutlined, LinkOutlined } from '@ant-design/icons-vue'
import { fetchDoc, fetchDocList } from '../api/client'
import type { DocEntry } from '../types'
import { renderMarkdown } from '../utils/markdown'

const catalog = ref<DocEntry[]>([])
const activeSlug = ref('overview')
const loading = ref(false)
const html = ref('')
const docTitle = ref('')

async function loadDoc(slug: string) {
  loading.value = true
  try {
    const doc = await fetchDoc(slug)
    docTitle.value = doc.title
    html.value = renderMarkdown(doc.markdown)
    activeSlug.value = slug
  } finally {
    loading.value = false
  }
}

async function onMenuClick(info: { key: string | number }) {
  await loadDoc(String(info.key))
}

onMounted(async () => {
  catalog.value = await fetchDocList()
  const first = catalog.value[0]?.slug ?? 'overview'
  await loadDoc(first)
})
</script>

<template>
  <div class="docs-layout">
    <a-layout class="docs-shell">
      <a-layout-sider width="220" theme="light" class="docs-sider">
        <div class="sider-head">
          <BookOutlined />
          <span>开发知识库</span>
        </div>
        <a-menu
          :selected-keys="[activeSlug]"
          mode="inline"
          @click="onMenuClick"
        >
          <a-menu-item v-for="item in catalog" :key="item.slug">
            {{ item.title }}
          </a-menu-item>
        </a-menu>
        <div class="sider-links">
          <a
            href="https://github.com/wgdl666/agent-recruiting-hub"
            target="_blank"
            rel="noopener"
          >
            <GithubOutlined /> GitHub 仓库
          </a>
          <a href="https://recruit.wgdl.tech" target="_blank" rel="noopener">
            <LinkOutlined /> recruit.wgdl.tech
          </a>
        </div>
      </a-layout-sider>

      <a-layout-content class="docs-content">
        <a-spin :spinning="loading">
          <a-typography-title :level="3">{{ docTitle }}</a-typography-title>
          <a-alert
            type="info"
            show-icon
            message="给接手开发/部署的同学：改 docs/*.md 后 make deploy 即可更新本站内容，无需改前端代码。"
            class="docs-tip"
          />
          <article class="markdown-body" v-html="html" />
        </a-spin>
      </a-layout-content>
    </a-layout>
  </div>
</template>

<style scoped>
.docs-layout {
  margin: -16px;
  min-height: calc(100vh - 120px);
}
.docs-shell {
  background: #fff;
  min-height: inherit;
}
.docs-sider {
  border-right: 1px solid #f0f0f0;
  padding: 12px 0;
}
.sider-head {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 16px 12px;
  font-weight: 600;
  color: #262626;
}
.sider-links {
  margin-top: 16px;
  padding: 12px 16px 0;
  border-top: 1px solid #f0f0f0;
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 13px;
}
.sider-links a {
  color: #1677ff;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.docs-content {
  padding: 16px 24px 32px;
  max-width: 920px;
}
.docs-tip {
  margin: 12px 0 20px;
}
.markdown-body :deep(h1) {
  font-size: 1.5em;
  margin: 1.2em 0 0.6em;
}
.markdown-body :deep(h2) {
  font-size: 1.25em;
  margin: 1.4em 0 0.5em;
  padding-bottom: 0.3em;
  border-bottom: 1px solid #eee;
}
.markdown-body :deep(h3) {
  font-size: 1.1em;
  margin: 1.2em 0 0.4em;
}
.markdown-body :deep(p),
.markdown-body :deep(li) {
  line-height: 1.75;
  color: #434343;
}
.markdown-body :deep(pre) {
  background: #f6f8fa;
  padding: 12px 14px;
  border-radius: 6px;
  overflow-x: auto;
  font-size: 13px;
}
.markdown-body :deep(code) {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.9em;
}
.markdown-body :deep(p code),
.markdown-body :deep(li code) {
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 4px;
}
.markdown-body :deep(table) {
  border-collapse: collapse;
  width: 100%;
  margin: 12px 0;
  font-size: 14px;
}
.markdown-body :deep(th),
.markdown-body :deep(td) {
  border: 1px solid #e8e8e8;
  padding: 8px 10px;
}
.markdown-body :deep(th) {
  background: #fafafa;
}
.markdown-body :deep(blockquote) {
  margin: 12px 0;
  padding: 8px 16px;
  border-left: 4px solid #91caff;
  background: #f0f5ff;
  color: #595959;
}
</style>
