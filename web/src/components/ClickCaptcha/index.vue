<template>
  <div class="click-captcha">
    <div class="click-captcha__header">
      <span class="click-captcha__tip">请依次点击：<strong>{{ chars }}</strong></span>
      <el-button link type="primary" size="small" @click="refresh" :loading="loading">
        <el-icon><Refresh /></el-icon> 换一张
      </el-button>
    </div>
    <div class="click-captcha__canvas" ref="canvasRef" @click="handleClick">
      <img
        v-if="bgUrl"
        ref="imgRef"
        :src="bgUrl"
        class="click-captcha__img"
        draggable="false"
      />
      <div
        v-for="(point, idx) in displayPoints"
        :key="idx"
        class="click-captcha__mark"
        :style="{ left: point.dx + 'px', top: point.dy + 'px' }"
      >
        <span>{{ idx + 1 }}</span>
      </div>
      <div v-if="result === 'success'" class="click-captcha__overlay click-captcha__overlay--success">
        <el-icon :size="36"><CircleCheck /></el-icon>
        <span class="click-captcha__overlay-msg">验证成功</span>
      </div>
      <div v-if="result === 'fail'" class="click-captcha__overlay click-captcha__overlay--fail">
        <el-icon :size="36"><CircleClose /></el-icon>
        <span class="click-captcha__overlay-msg">{{ message }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Refresh, CircleCheck, CircleClose } from '@element-plus/icons-vue'
import { getCaptcha, verifyCaptcha, type CaptchaPoint } from '@/api/captcha'

const emit = defineEmits<{
  (e: 'success', token: string): void
}>()

const loading = ref(false)
const bgUrl = ref('')
const bgWidth = ref(640)
const bgHeight = ref(200)
const chars = ref('')
const token = ref('')
const clickedPoints = ref<CaptchaPoint[]>([])
const result = ref<'success' | 'fail' | ''>('')
const message = ref('')
const canvasRef = ref<HTMLDivElement>()
const imgRef = ref<HTMLImageElement>()

const displayPoints = computed(() => {
  const img = imgRef.value
  if (!img) return []
  const imgW = img.clientWidth
  const imgH = img.clientHeight
  const sx = imgW / bgWidth.value
  const sy = imgH / bgHeight.value
  return clickedPoints.value.map((p) => ({
    dx: p.x * sx,
    dy: p.y * sy,
  }))
})

async function loadCaptcha() {
  loading.value = true
  result.value = ''
  message.value = ''
  clickedPoints.value = []
  try {
    const res = await getCaptcha()
    const data = res.data
    token.value = data.token
    bgUrl.value = data.bg
    bgWidth.value = data.bgWidth
    bgHeight.value = data.bgHeight
    chars.value = data.chars
  } catch {
    message.value = '获取验证码失败'
  } finally {
    loading.value = false
  }
}

function refresh() {
  loadCaptcha()
}

function handleClick(e: MouseEvent) {
  if (result.value === 'success') return

  const img = imgRef.value
  if (!img) return

  const rect = img.getBoundingClientRect()
  const displayX = e.clientX - rect.left
  const displayY = e.clientY - rect.top

  const scaleX = bgWidth.value / rect.width
  const scaleY = bgHeight.value / rect.height
  const x = Math.round(displayX * scaleX)
  const y = Math.round(displayY * scaleY)

  clickedPoints.value.push({ x, y })

  if (clickedPoints.value.length === chars.value.length) {
    verify()
  }
}

async function verify() {
  try {
    const res = await verifyCaptcha({
      token: token.value,
      points: clickedPoints.value,
    })
    const data = res.data
    if (data.success) {
      result.value = 'success'
      emit('success', data.token)
    } else {
      result.value = 'fail'
      message.value = data.message || '验证失败'
      setTimeout(() => {
        refresh()
      }, 1500)
    }
  } catch {
    result.value = 'fail'
    message.value = '验证请求失败'
    setTimeout(() => {
      refresh()
    }, 1500)
  }
}

onMounted(() => {
  loadCaptcha()
})
</script>

<style lang="scss" scoped>
.click-captcha {
  border: 1px solid var(--el-border-color);
  border-radius: 6px;
  overflow: visible;
  background: #fff;
  width: 100%;
}

.click-captcha__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 12px;
  background: #f5f7fa;
  border-bottom: 1px solid var(--el-border-color-lighter);
  font-size: 13px;
  color: var(--el-text-color-regular);

  strong {
    color: var(--el-color-primary);
    font-size: 16px;
    letter-spacing: 4px;
    margin-left: 4px;
  }
}

.click-captcha__canvas {
  position: relative;
  cursor: crosshair;
  user-select: none;
  line-height: 0;
}

.click-captcha__img {
  display: block;
  width: 100%;
  height: auto;
}

.click-captcha__mark {
  position: absolute;
  width: 24px;
  height: 24px;
  margin-left: -12px;
  margin-top: -12px;
  background: var(--el-color-primary);
  color: #fff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.3);
  pointer-events: none;
  z-index: 2;
}

.click-captcha__overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  z-index: 3;

  &--success {
    background: rgba(103, 194, 58, 0.85);
    color: #fff;
  }

  &--fail {
    background: rgba(245, 108, 108, 0.85);
    color: #fff;
  }
}

.click-captcha__overlay-msg {
  font-size: 14px;
  font-weight: 500;
}
</style>
