<template>
  <div class="app-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>支付订单</span>
        </div>
      </template>

      <div class="search-bar">
        <el-input v-model="queryParams.subject" placeholder="搜索订单标题" clearable style="width: 200px" @keyup.enter="loadData" @clear="loadData" />
        <el-select v-model="queryParams.channel" placeholder="支付渠道" clearable style="width: 140px" @change="loadData">
          <el-option label="微信支付" value="wechat" />
          <el-option label="支付宝" value="alipay" />
        </el-select>
        <el-select v-model="queryParams.status" placeholder="订单状态" clearable style="width: 140px" @change="loadData">
          <el-option label="待支付" value="0" />
          <el-option label="已支付" value="1" />
          <el-option label="已关闭" value="2" />
          <el-option label="已退款" value="3" />
        </el-select>
        <el-button type="primary" @click="loadData">搜索</el-button>
      </div>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="orderNo" label="订单号" width="200" />
        <el-table-column prop="subject" label="订单标题" min-width="150" />
        <el-table-column label="金额" width="100" align="right">
          <template #default="{ row }">
            <span style="color: #f56c6c; font-weight: bold">¥{{ (row.amount / 100).toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="渠道" width="100" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.channel === 'wechat'" type="success" size="small">微信</el-tag>
            <el-tag v-else-if="row.channel === 'alipay'" type="primary" size="small">支付宝</el-tag>
            <el-tag v-else size="small">{{ row.channel }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.status === 0" type="info" size="small">待支付</el-tag>
            <el-tag v-else-if="row.status === 1" type="success" size="small">已支付</el-tag>
            <el-tag v-else-if="row.status === 2" type="warning" size="small">已关闭</el-tag>
            <el-tag v-else-if="row.status === 3" type="danger" size="small">已退款</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="tradeNo" label="第三方交易号" width="180" />
        <el-table-column label="支付时间" width="180">
          <template #default="{ row }">
            {{ row.paidAt ? formatDate(row.paidAt) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button v-if="row.status === 0" type="danger" link size="small" @click="handleClose(row)">关闭</el-button>
            <el-button type="primary" link size="small" @click="handleDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        layout="total, prev, pager, next"
        style="margin-top: 16px; justify-content: flex-end"
        @current-change="loadData"
      />
    </el-card>

    <el-dialog v-model="detailVisible" title="订单详情" width="500px">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="订单号">{{ detailData.orderNo }}</el-descriptions-item>
        <el-descriptions-item label="订单标题">{{ detailData.subject }}</el-descriptions-item>
        <el-descriptions-item label="金额">¥{{ (detailData.amount / 100).toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="渠道">{{ detailData.channel === 'wechat' ? '微信支付' : '支付宝' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag v-if="detailData.status === 0" type="info">待支付</el-tag>
          <el-tag v-else-if="detailData.status === 1" type="success">已支付</el-tag>
          <el-descriptions-item v-else-if="detailData.status === 2" type="warning">已关闭</el-descriptions-item>
          <el-tag v-else type="danger">已退款</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="第三方交易号">{{ detailData.tradeNo || '-' }}</el-descriptions-item>
        <el-descriptions-item label="支付时间">{{ detailData.paidAt ? formatDate(detailData.paidAt) : '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDate(detailData.createdAt) }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPayOrderList, closePayOrder, type PayOrder } from '@/api/payment'

const loading = ref(false)
const tableData = ref<PayOrder[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const queryParams = reactive({
  subject: '',
  channel: '',
  status: '',
})

const detailVisible = ref(false)
const detailData = ref<PayOrder>({} as PayOrder)

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

async function loadData() {
  loading.value = true
  try {
    const res = await getPayOrderList({
      subject: queryParams.subject,
      channel: queryParams.channel,
      status: queryParams.status,
      page: page.value,
      pageSize: pageSize.value,
    })
    tableData.value = res.data.list
    total.value = res.data.total
  } finally {
    loading.value = false
  }
}

async function handleClose(row: PayOrder) {
  try {
    await ElMessageBox.confirm(`确定要关闭订单「${row.orderNo}」吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await closePayOrder(row.orderNo)
    ElMessage.success('订单已关闭')
    loadData()
  } catch {}
}

function handleDetail(row: PayOrder) {
  detailData.value = row
  detailVisible.value = true
}

onMounted(() => loadData())
</script>

<style lang="scss" scoped>
.search-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}
</style>
