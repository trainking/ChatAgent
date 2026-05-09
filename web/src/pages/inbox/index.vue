<template>
  <div class="inbox-page">
    <el-card>
      <template #header><span>{{ $t('inbox.title') }}</span></template>
      <el-empty :description="$t('inbox.placeholder')" />
    </el-card>
  </div>
</template>
