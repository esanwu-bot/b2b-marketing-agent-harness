<template>
  <a-config-provider :locale="zhCN" :theme="themeConfig">
    <router-view />
  </a-config-provider>
</template>

<script setup lang="ts">
import zhCN from 'ant-design-vue/es/locale/zh_CN';
import { theme as antdTheme } from 'ant-design-vue';

const themeConfig = {
  token: { colorPrimary: '#2563eb', borderRadius: 8 },
  algorithm: antdTheme.defaultAlgorithm,
};
</script>
