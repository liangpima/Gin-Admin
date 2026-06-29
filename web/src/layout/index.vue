<template>
  <div class="app-wrapper" :class="{ 'sidebar-opened': appStore.sidebar.opened }">
    <div v-if="isMobile && appStore.sidebar.opened" class="sidebar-overlay" @click="appStore.closeSidebar(true)" />
    <Sidebar class="sidebar-container" />
    <div class="main-container">
      <Navbar />
      <TagsView v-if="!isMobile" />
      <AppMain />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAppStore } from '@/store/modules/app'
import { useResponsive } from '@/hooks/useResponsive'
import Sidebar from './components/Sidebar.vue'
import Navbar from './components/Navbar.vue'
import TagsView from './components/TagsView.vue'
import AppMain from './components/AppMain.vue'

const appStore = useAppStore()
const { isMobile } = useResponsive()
</script>

<style lang="scss" scoped>
@use '@/assets/styles/responsive.scss' as *;

.app-wrapper {
  position: relative;
  height: 100%;
  width: 100%;
  display: flex;
  overflow: hidden;
}

.sidebar-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: var(--color-bg-mask);
  z-index: 999;
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.sidebar-container {
  width: var(--sidebar-width);
  height: 100vh;
  background: var(--color-bg-sidebar);
  transition: width var(--transition-normal);
  flex-shrink: 0;
  overflow: hidden;
  border-right: 1px solid var(--color-sidebar-logo-border);

  @include mobile {
    position: fixed;
    z-index: 1000;
    width: 0;
    overflow: hidden;

    &.is-open {
      width: var(--sidebar-width);
    }
  }
}

.main-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  overflow: hidden;
  background: var(--color-bg-page);
  transition: background-color var(--transition-slow);

  @include mobile {
    width: 100%;
  }
}
</style>
