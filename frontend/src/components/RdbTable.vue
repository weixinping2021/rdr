<template>
  <a-table :columns="columns" :data-source="data" bordered row-key="key">
      <template #bodyCell="{ column, record }">
    <a-tooltip v-if="column.ellipsis" :title="record[column.dataIndex]">
      <span>{{ record[column.dataIndex] }}</span>
    </a-tooltip>
  </template>
  </a-table>
</template>

<script setup>
import { defineProps } from 'vue'

// 接收父组件传入的数据
defineProps({
  data: {
    type: Array,
    default: () => []
  }
})

const columns = [
  { title: 'key', dataIndex: 'key', className: 'column-money', ellipsis: true},
  { title: '类型', dataIndex: 'type' ,width: 100},
  { title: '占有内存', dataIndex: 'readble_size',width: 200 },
  { title: '元素个数', dataIndex: 'elements',width: 100 },
  { title: '过期时间', dataIndex: 'expire',width: 200 },
  { title: 'db', dataIndex: 'db' ,width: 100},
]
</script>
