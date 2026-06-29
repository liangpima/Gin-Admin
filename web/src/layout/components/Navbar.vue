<template>
  <div class="navbar">
    <div class="left-menu">
      <el-icon class="hamburger" @click="toggleSidebar">
        <Fold v-if="appStore.sidebar.opened && !isMobile" />
        <Expand v-else />
      </el-icon>
      <Breadcrumb v-if="!isMobile" />
    </div>

    <div class="right-menu">
      <el-tooltip content="刷新页面（清除缓存）" placement="bottom">
        <span class="right-menu__icon" @click="handleRefresh">
          <el-icon :size="18"><Refresh /></el-icon>
        </span>
      </el-tooltip>

      <el-tooltip :content="isDark ? '切换亮色模式' : '切换暗色模式'" placement="bottom">
        <span class="right-menu__icon" @click="toggleTheme">
          <el-icon :size="18">
            <Moon v-if="!isDark" />
            <Sunny v-else />
          </el-icon>
        </span>
      </el-tooltip>

      <el-dropdown class="avatar-container" trigger="click">
        <div class="avatar-wrapper">
          <el-avatar :size="32" :src="userStore.userInfo?.avatar || undefined" class="avatar">
            {{ userStore.userInfo?.nickname?.charAt(0) || 'A' }}
          </el-avatar>
          <span class="username" v-if="!isMobile">{{ userStore.userInfo?.nickname || userStore.userInfo?.username }}</span>
          <el-icon class="avatar-arrow"><ArrowDown /></el-icon>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item divided @click="handleLogout">
              <el-icon><SwitchButton /></el-icon>
              退出登录
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAppStore } from '@/store/modules/app'
import { useUserStore } from '@/store/modules/user'
import { useResponsive } from '@/hooks/useResponsive'
import { useTheme } from '@/hooks/useTheme'
import Breadcrumb from './Breadcrumb.vue'

const router = useRouter()
const appStore = useAppStore()
const userStore = useUserStore()
const { isMobile } = useResponsive()
const { isDark, toggleTheme } = useTheme()

function handleRefresh() {
  location.reload()
}

function toggleSidebar() {
  appStore.toggleSidebar()
}

async function handleLogout() {
  await userStore.logout()
  router.push('/login')
}
</script>

<style lang="scss" scoped>
@use '@/assets/styles/responsive.scss' as *;

.navbar {
  height: var(--navbar-height);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--spacing-lg);
  background: var(--color-bg-navbar);
  box-shadow: var(--shadow-sm);
  border-bottom: 1px solid var(--color-border-lighter);
  transition: background-color var(--transition-slow), border-color var(--transition-slow);

  @include mobile {
    padding: 0 var(--spacing-md);
  }
}

.left-menu {
  display: flex;
  align-items: center;
}

.hamburger {
  font-size: 20px;
  cursor: pointer;
  margin-right: var(--spacing-md);
  color: var(--color-text-regular);
  padding: 4px;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);

  &:hover {
    background: var(--color-bg-hover);
    color: var(--color-primary);
  }
}

.right-menu {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.right-menu__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  cursor: pointer;
  color: var(--color-text-regular);
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);

  &:hover {
    background: var(--color-bg-hover);
    color: var(--color-primary);
  }
}

.avatar-wrapper {
  display: flex;
  align-items: center;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: var(--radius-md);
  transition: all var(--transition-fast);

  &:hover {
    background: var(--color-bg-hover);
  }

  .avatar {
    background: var(--color-primary);
    color: white;
    font-weight: var(--font-weight-semibold);
  }

  .username {
    margin-left: var(--spacing-sm);
    font-size: var(--font-size-base);
    color: var(--color-text-regular);
    font-weight: var(--font-weight-medium);

    @include mobile {
      display: none;
    }
  }

  .avatar-arrow {
    margin-left: var(--spacing-xs);
    font-size: 12px;
    color: var(--color-text-secondary);
  }
}
</style>
