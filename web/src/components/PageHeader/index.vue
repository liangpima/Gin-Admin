<template>
  <div class="page-header">
    <div class="page-header__content">
      <h2 class="page-header__title">{{ title }}</h2>
      <p v-if="description" class="page-header__description">{{ description }}</p>
    </div>
    <div v-if="$slots.actions" class="page-header__actions">
      <slot name="actions" />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

<style lang="scss" scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-lg);
  background: var(--color-bg-card);
  border-radius: var(--card-radius);
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--color-border-lighter);
  margin-bottom: var(--spacing-base);
}

.page-header__content {
  flex: 1;
}

.page-header__title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-title);
  margin: 0;
  line-height: var(--line-height-tight);
}

.page-header__description {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  margin: var(--spacing-xs) 0 0;
}

.page-header__actions {
  display: flex;
  gap: var(--spacing-sm);
}
</style>
