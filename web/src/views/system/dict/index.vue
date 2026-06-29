<template>
  <div class="app-container">
    <el-row :gutter="16">
      <el-col :xs="24" :sm="24" :md="10" class="dict-type-col">
        <el-card class="table-card">
          <template #header>
            <div class="card-header">
              <span>字典类型</span>
              <el-button type="primary" @click="handleAddType">新增</el-button>
            </div>
          </template>
          <el-table :data="typeList" v-loading="typeLoading" border stripe highlight-current-row @current-change="handleTypeChange">
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="name" label="字典名称" min-width="100" />
            <el-table-column prop="type" label="字典类型" min-width="100" />
            <el-table-column prop="status" label="状态" width="70">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '正常' : '停用' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="100">
              <template #default="{ row }">
                <el-button type="danger" link size="small" @click="handleDeleteType(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="24" :md="14">
        <el-card class="table-card">
          <template #header>
            <div class="card-header">
              <span>字典数据 {{ currentType ? `- ${currentType.name}` : '' }}</span>
              <el-button type="primary" :disabled="!currentType" @click="handleAddData">新增</el-button>
            </div>
          </template>
          <el-table :data="dataList" v-loading="dataLoading" border stripe>
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="label" label="字典标签" min-width="100" />
            <el-table-column prop="value" label="字典键值" min-width="80" />
            <el-table-column prop="sort" label="排序" width="60" />
            <el-table-column prop="status" label="状态" width="70">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '正常' : '停用' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="100">
              <template #default="{ row }">
                <el-button type="danger" link size="small" @click="handleDeleteData(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="typeDialogVisible" title="新增字典类型" width="600px" destroy-on-close>
      <el-form ref="typeFormRef" :model="typeForm" :rules="typeFormRules" label-width="80px">
        <el-form-item label="字典名称" prop="name"><el-input v-model="typeForm.name" placeholder="请输入字典名称" /></el-form-item>
        <el-form-item label="字典类型" prop="type"><el-input v-model="typeForm.type" placeholder="请输入字典类型" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="typeDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleTypeSubmit">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="dataDialogVisible" title="新增字典数据" width="600px" destroy-on-close>
      <el-form ref="dataFormRef" :model="dataForm" :rules="dataFormRules" label-width="80px">
        <el-form-item label="字典标签" prop="label"><el-input v-model="dataForm.label" placeholder="请输入字典标签" /></el-form-item>
        <el-form-item label="字典键值" prop="value"><el-input v-model="dataForm.value" placeholder="请输入字典键值" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="dataForm.sort" :min="0" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dataDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleDataSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import { getDictTypeList, createDictType, deleteDictType, getDictDataList, createDictData, deleteDictData } from '@/api/dict'

const typeLoading = ref(false)
const dataLoading = ref(false)
const typeList = ref<any[]>([])
const dataList = ref<any[]>([])
const currentType = ref<any>(null)

const typeDialogVisible = ref(false)
const dataDialogVisible = ref(false)
const typeFormRef = ref<FormInstance>()
const dataFormRef = ref<FormInstance>()

const typeForm = reactive({ name: '', type: '' })
const dataForm = reactive({ label: '', value: '', sort: 0 })
const typeFormRules = { name: [{ required: true, message: '必填' }], type: [{ required: true, message: '必填' }] }
const dataFormRules = { label: [{ required: true, message: '必填' }], value: [{ required: true, message: '必填' }] }

async function loadTypes() {
  typeLoading.value = true
  try {
    const res = await getDictTypeList({ page: 1, pageSize: 100 })
    typeList.value = res.data.list
  } finally { typeLoading.value = false }
}

async function loadData() {
  if (!currentType.value) return
  dataLoading.value = true
  try {
    const res = await getDictDataList({ dictType: currentType.value.type, page: 1, pageSize: 100 })
    dataList.value = res.data.list
  } finally { dataLoading.value = false }
}

function handleTypeChange(row: any) {
  currentType.value = row
  loadData()
}

function handleAddType() { typeForm.name = ''; typeForm.type = ''; typeDialogVisible.value = true }

async function handleTypeSubmit() {
  const valid = await typeFormRef.value?.validate().catch(() => false)
  if (!valid) return
  await createDictType(typeForm)
  ElMessage.success('操作成功'); typeDialogVisible.value = false; loadTypes()
}

async function handleDeleteType(row: any) {
  await ElMessageBox.confirm('确认删除？', '提示', { type: 'warning' })
  await deleteDictType(row.id)
  ElMessage.success('删除成功'); loadTypes()
}

function handleAddData() { dataForm.label = ''; dataForm.value = ''; dataForm.sort = 0; dataDialogVisible.value = true }

async function handleDataSubmit() {
  const valid = await dataFormRef.value?.validate().catch(() => false)
  if (!valid) return
  await createDictData({ ...dataForm, dictType: currentType.value.type })
  ElMessage.success('操作成功'); dataDialogVisible.value = false; loadData()
}

async function handleDeleteData(row: any) {
  await ElMessageBox.confirm('确认删除？', '提示', { type: 'warning' })
  await deleteDictData(row.id)
  ElMessage.success('删除成功'); loadData()
}

onMounted(() => { loadTypes() })
</script>

<style lang="scss" scoped>
@use '@/assets/styles/responsive.scss' as *;

.table-card {
  :deep(.el-card__header) {
    border-bottom-color: var(--color-border-lighter);
  }
}

@include mobile {
  .dict-type-col {
    margin-bottom: var(--spacing-md);
  }
}
</style>
