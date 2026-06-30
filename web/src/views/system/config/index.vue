<template>
  <div class="app-container">
    <el-card class="table-card">
      <template #header>
        <div class="card-header">
          <span>参数管理</span>
          <el-button type="primary" @click="handleAdd">新增配置</el-button>
        </div>
      </template>
      <el-table :data="tableData" v-loading="loading" border>
        <el-table-column prop="name" label="参数名称" min-width="100" />
        <el-table-column prop="key" label="参数键名" min-width="120" />
        <el-table-column prop="value" label="参数键值" min-width="100" show-overflow-tooltip />
        <el-table-column prop="type" label="系统内置" width="90">
          <template #default="{ row }">
            <el-tag :type="row.type === 0 ? 'danger' : 'info'" size="small">{{ row.type === 0 ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">{{ formatDateTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" layout="total, prev, pager, next" style="margin-top: 16px; justify-content: flex-end" @current-change="loadData" />
    </el-card>

    <FormDialog v-model="dialogVisible" :title="dialogTitle" :loading="submitLoading" @submit="handleSubmit">
      <el-form ref="formRef" :model="form" :rules="formRules" label-width="80px">
        <el-form-item label="参数名称" prop="name"><el-input v-model="form.name" placeholder="请输入参数名称" /></el-form-item>
        <el-form-item label="参数键名" prop="key"><el-input v-model="form.key" :disabled="!!form.id" placeholder="请输入参数键名" /></el-form-item>
        <el-form-item label="参数键值" prop="value"><el-input v-model="form.value" type="textarea" placeholder="请输入参数键值" /></el-form-item>
        <el-form-item label="系统内置">
          <el-radio-group v-model="form.type"><el-radio :value="0">是</el-radio><el-radio :value="1">否</el-radio></el-radio-group>
        </el-form-item>
      </el-form>
    </FormDialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { getConfigList, createConfig, updateConfig, deleteConfig } from '@/api/config'
import FormDialog from '@/components/FormDialog/index.vue'
import { formatDateTime } from '@/utils/format'

const loading = ref(false)
const submitLoading = ref(false)
const tableData = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref<FormInstance>()

const form = reactive({ id: 0, name: '', key: '', value: '', type: 1 })
const formRules = {
  name: [{ required: true, message: '请输入参数名称', trigger: 'blur' }],
  key: [{ required: true, message: '请输入参数键名', trigger: 'blur' }],
}

async function loadData() {
  loading.value = true
  try {
    const res = await getConfigList({ page: page.value, pageSize: pageSize.value })
    tableData.value = res.data.list
    total.value = res.data.total
  } finally { loading.value = false }
}

function handleAdd() {
  form.id = 0; form.name = ''; form.key = ''; form.value = ''; form.type = 1
  dialogTitle.value = '新增配置'; dialogVisible.value = true
}

function handleEdit(row: any) {
  Object.assign(form, row)
  dialogTitle.value = '编辑配置'; dialogVisible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true
  try {
    if (form.id) { await updateConfig(form) }
    else { await createConfig(form) }
    ElMessage.success('操作成功'); dialogVisible.value = false; loadData()
  } finally { submitLoading.value = false }
}

async function handleDelete(row: any) {
  await ElMessageBox.confirm('确认删除？', '提示', { type: 'warning' })
  await deleteConfig(row.id)
  ElMessage.success('删除成功'); loadData()
}

onMounted(() => { loadData() })
</script>

<style lang="scss" scoped>
.table-card {
  :deep(.el-card__header) {
    border-bottom-color: var(--color-border-lighter);
  }
}
</style>
