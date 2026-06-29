<template>
  <section class="app-main">
    <router-view v-slot="{ Component }">
      <transition name="fade-transform" mode="out-in">
        <keep-alive :include="cachedViews">
          <component :is="Component" :key="route.path" />
        </keep-alive>
      </transition>
    </router-view>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useTagsViewStore } from '@/store/modules/tagsView'

const route = useRoute()
const tagsViewStore = useTagsViewStore()

const cachedViews = computed(() => tagsViewStore.cachedViews)
</script>

<style lang="scss" scoped>
@use '@/assets/styles/responsive.scss' as *;

.app-main {
  flex: 1;
  padding: var(--spacing-lg);
  overflow: auto;
  background: var(--color-bg-page);
  transition: background-color var(--transition-slow);

  @include mobile {
    padding: var(--spacing-md);
  }

  @include tablet {
    padding: var(--spacing-base);
  }
}
</style>
