<template>
  <el-tag v-if="option" :type="tagType" size="small">{{ option.label }}</el-tag>
  <span v-else class="dict-tag__raw">{{ displayText }}</span>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { loadDict, normalizeDictValue, type DictOption } from '@/hooks/useDict'

type TagType = 'primary' | 'success' | 'info' | 'warning' | 'danger'

const VALID_TAG_TYPES: TagType[] = ['primary', 'success', 'info', 'warning', 'danger']

const props = withDefaults(
  defineProps<{
    /** 字典类型编码，如 sys_user_status */
    type: string
    /** 当前值，数字或字符串都可以（内部统一转字符串比对） */
    value?: string | number | null
    /** 值为空时的占位文本 */
    emptyText?: string
  }>(),
  { value: '', emptyText: '-' },
)

const options = ref<DictOption[]>([])

async function load() {
  options.value = await loadDict(props.type)
}

onMounted(load)
watch(() => props.type, load)

const key = computed(() => normalizeDictValue(props.value))
const option = computed(() => options.value.find((o) => o.value === key.value))

/**
 * 取不到对应选项时降级显示**原始值**，而不是留空。
 *
 * 字典是运营侧可维护的数据：可能还没配、可能选项被停用、可能新加了状态值
 * 但字典没跟上。任何一种情况下，把真实值显示出来都比空白有用得多 ——
 * 空白会让人以为「这条数据没有状态」，而显示 3 至少能定位到问题。
 */
const displayText = computed(() => (key.value === '' ? props.emptyText : key.value))

/**
 * el-tag 的 type 只接受固定几个字面量，字典里配错（比如写了中文）时
 * 直接透传会得到一个没有样式的标签，故做一次白名单兜底。
 */
const tagType = computed<TagType>(() => {
  const v = option.value?.listClass ?? ''
  return VALID_TAG_TYPES.includes(v as TagType) ? (v as TagType) : 'info'
})
</script>

<style lang="scss" scoped>
.dict-tag__raw {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
</style>
