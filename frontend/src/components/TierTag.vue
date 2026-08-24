<script setup lang="ts">
defineProps<{ tier: string }>()
</script>

<template>
  <el-tag :type="tier === 'S' ? 'danger' : tier === 'A' ? 'warning' : tier === 'B' ? '' : tier === 'C' ? 'info' : 'info'" effect="dark" size="small">
    {{ tier }}
  </el-tag>
</template>
