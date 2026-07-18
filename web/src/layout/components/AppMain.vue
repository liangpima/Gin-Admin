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
  padding: 20px;
  overflow: auto;
  background: var(--color-bg-page);

  @include mobile {
    padding: 12px;
  }

  @include tablet {
    padding: 16px;
  }
}
</style>
