<template>
  <div class="tags-view-container">
    <el-scrollbar class="tags-scrollbar">
      <div class="tags-view-wrapper">
        <router-link
          v-for="tag in visitedViews"
          :key="tag.path"
          :to="tag.path"
          :class="isActive(tag) ? 'active' : ''"
          class="tags-view-item"
        >
          <span class="tags-view-item__text">{{ tag.title }}</span>
          <el-icon
            v-if="!tag.meta?.affix"
            class="tags-view-item__close"
            @click.prevent.stop="closeTag(tag)"
          >
            <Close />
          </el-icon>
        </router-link>
      </div>
    </el-scrollbar>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useTagsViewStore } from '@/store/modules/tagsView'

const route = useRoute()
const router = useRouter()
const tagsViewStore = useTagsViewStore()

const visitedViews = computed(() => tagsViewStore.visitedViews)

function isActive(tag: any): boolean {
  return tag.path === route.path
}

function closeTag(tag: any) {
  tagsViewStore.delView(tag)
  if (isActive(tag)) {
    const views = tagsViewStore.visitedViews
    if (views.length > 0) {
      const lastView = views[views.length - 1]
      router.push(lastView.path).catch(() => {})
    } else {
      router.push('/').catch(() => {})
    }
  }
}
</script>

<style lang="scss" scoped>
.tags-view-container {
  height: var(--tagsview-height);
  background: var(--color-bg-card);
  border-bottom: 1px solid var(--color-border-lighter);
  transition: background-color var(--transition-slow), border-color var(--transition-slow);
}

.tags-view-wrapper {
  display: flex;
  align-items: center;
  padding: 0 var(--spacing-sm);
  height: var(--tagsview-height);
  gap: 6px;
}

.tags-view-item {
  display: inline-flex;
  align-items: center;
  padding: 0 var(--spacing-md);
  height: 28px;
  border: 1px solid var(--color-border-lighter);
  color: var(--color-text-regular);
  background: var(--color-bg-card);
  font-size: var(--font-size-xs);
  text-decoration: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
  white-space: nowrap;

  &:hover {
    color: var(--color-primary);
    border-color: var(--color-primary-200);
    background: var(--color-primary-50);
  }

  &.active {
    background-color: var(--color-bg-tag-active);
    color: #ffffff;
    border-color: var(--color-bg-tag-active);
    font-weight: var(--font-weight-medium);

    .tags-view-item__close {
      color: rgba(255, 255, 255, 0.7);

      &:hover {
        color: #fff;
        background: rgba(255, 255, 255, 0.2);
      }
    }
  }
}

.tags-view-item__text {
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tags-view-item__close {
  margin-left: 4px;
  cursor: pointer;
  border-radius: 50%;
  font-size: 12px;
  padding: 1px;
  transition: all var(--transition-fast);

  &:hover {
    background-color: rgba(0, 0, 0, 0.12);
  }
}
</style>
