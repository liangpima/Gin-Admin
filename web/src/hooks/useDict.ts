import { ref, reactive, onMounted } from 'vue'
import { getDictDataByType, type DictDataItem } from '@/api/dict'

/**
 * 字典选项（前端内部结构）。
 *
 * value 统一转成字符串：库里 value 是 varchar，而后端返回的可能是
 * "1"，页面里的行数据却是数字 1，直接比较永远不相等 ——
 * 这类「类型不一致导致查不到」的问题排查起来很费时间，故在此收敛。
 */
export interface DictOption {
  label: string
  value: string
  listClass: string
  cssClass: string
}

const cache = new Map<string, DictOption[]>()
const pending = new Map<string, Promise<DictOption[]>>()

function toOptions(items?: DictDataItem[] | null): DictOption[] {
  return (items ?? []).map((item) => ({
    label: item.label,
    value: String(item.value ?? ''),
    listClass: item.listClass ?? '',
    cssClass: item.cssClass ?? '',
  }))
}

/** 把任意取值归一成用于比对字典的字符串 */
export function normalizeDictValue(value: unknown): string {
  if (value === null || value === undefined) return ''
  return String(value)
}

/**
 * 加载某个类型的字典选项。
 *
 * - 命中模块级缓存直接返回，不发请求
 * - 同一类型的并发调用共用同一个请求（表格里几十个 DictTag 只会发一次）
 * - 失败**不写缓存**，下次调用会重试；同时吞掉异常，避免字典故障把页面搞崩
 */
export function loadDict(type: string): Promise<DictOption[]> {
  if (!type) return Promise.resolve([])

  const cached = cache.get(type)
  if (cached) return Promise.resolve(cached)

  const inflight = pending.get(type)
  if (inflight) return inflight

  const task = (async () => {
    try {
      const res = await getDictDataByType(type)
      const options = toOptions(res.data)
      cache.set(type, options)
      return options
    } catch (err) {
      console.warn(`[useDict] 字典 ${type} 加载失败，将降级显示原始值`, err)
      return []
    } finally {
      pending.delete(type)
    }
  })()

  pending.set(type, task)
  return task
}

/**
 * 清空字典缓存。
 * 在「数据字典」页新增/修改/删除选项后调用，避免其他页面继续用旧选项。
 */
export function clearDictCache(type?: string) {
  if (type) {
    cache.delete(type)
  } else {
    cache.clear()
  }
}

/**
 * 页面级字典用法：
 *
 * ```ts
 * const { getList, getLabel } = useDict('sys_user_status')
 * ```
 * 组件挂载时自动加载，模板里用 getList('sys_user_status') 渲染下拉。
 * 若只需在表格里回显标签，直接用 `<DictTag>` 组件更省事。
 */
export function useDict(...types: string[]) {
  const dicts = reactive<Record<string, DictOption[]>>({})
  const loading = ref(false)

  async function load() {
    if (types.length === 0) return
    loading.value = true
    try {
      const results = await Promise.all(types.map((t) => loadDict(t)))
      types.forEach((t, i) => {
        dicts[t] = results[i]
      })
    } finally {
      loading.value = false
    }
  }

  /** 重新拉取（忽略缓存） */
  async function reload() {
    clearDictCache()
    await load()
  }

  onMounted(load)

  /**
   * 取选项列表。可传 fallback：字典为空（未配置/被清空/加载失败）时返回它。
   *
   * 下拉框务必传 fallback —— 字典是可被运营删除的数据，清空后若直接返回空数组，
   * 筛选框会变成空下拉，用户连最基本的选项都选不了，属于功能级故障。
   * 表格里的标签回显不需要 fallback（DictTag 会降级显示原始值）。
   */
  const getList = (type: string, fallback: DictOption[] = []): DictOption[] => {
    const list = dicts[type] ?? []
    return list.length > 0 ? list : fallback
  }

  const find = (type: string, value: unknown): DictOption | undefined => {
    const key = normalizeDictValue(value)
    return getList(type).find((o) => o.value === key)
  }

  /**
   * 取标签文本。字典里找不到时**降级返回原始值** ——
   * 这样即使字典还没配好（或选项被停用），页面展示的也是真实数据，
   * 而不是一片空白。
   */
  const getLabel = (type: string, value: unknown): string => {
    const hit = find(type, value)
    return hit ? hit.label : normalizeDictValue(value)
  }

  /** 取 el-tag 的 type，取自字典的 listClass（表格回显样式） */
  const getTagType = (type: string, value: unknown): string => {
    return find(type, value)?.listClass || 'info'
  }

  return { dicts, loading, load, reload, getList, getLabel, getTagType, find }
}
