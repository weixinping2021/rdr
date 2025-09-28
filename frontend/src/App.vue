<template>
  <a-layout style="min-height: 100vh">

    <a-layout-sider>
      <div class="logo" />
      <a-spin :spinning="spinning"> <a-button @click="openFiles" block style="margin: 16px 0;">
          选择文件
        </a-button></a-spin>
      <a-menu v-model:selectedKeys="current" mode="inline" theme="dark" :items="items1" @click="handleClick"></a-menu>
    </a-layout-sider>
    <a-layout>
      <a-layout-content style="margin: 0 16px">
        <template v-if="items1.length === 0">
          <img src="./assets/images/redis.png" style="display:block;margin:40px auto;max-width:300px;" />
        </template>
        <template v-else>
          <a-row :gutter="24" :style="{ marginTop: '24px' }">
            <a-col :span="24">
              <a-card :bordered="false">
                <a-descriptions :title="file_name">
                  <a-descriptions-item label="key总数量">{{ totalkeys }}</a-descriptions-item>
                  <a-descriptions-item label="总内存">{{ total_memory }}</a-descriptions-item>
                </a-descriptions>
              </a-card>
            </a-col>
          </a-row>
          <a-row :gutter="24" :style="{ marginTop: '24px' }">
            <a-col :sm="24" :md="12" :xl="12" :style="{ marginBottom: '24px' }">
              <a-card title="各类型Key内存占用情况" :bordered="false">
                <BarChartMemory :barData="mbardata" :columns="mcolumns" />
              </a-card>
            </a-col>
            <a-col :sm="24" :md="12" :xl="12" :style="{ marginBottom: '24px' }">
              <a-card title="各类型Key数量分布情况" :bordered="false">
                <BarChartCount :barData="bardata" :columns="columns" />
              </a-card>
            </a-col>
          </a-row>
          <a-row :gutter="24" :style="{ marginTop: '24px' }">
            <a-col :sm="24" :md="12" :xl="12" :style="{ marginBottom: '24px' }">
              <a-card title="Key过期时间分布(内存)" :bordered="false">
                <BarChartMemory :barData="embardata" :columns="emcolumns" />
              </a-card>
            </a-col>
            <a-col :sm="24" :md="12" :xl="12" :style="{ marginBottom: '24px' }">
              <a-card title="Key过期时间分布(数量)" :bordered="false">
                <BarChartCount :barData="ebardata" :columns="ecolumns" />
              </a-card>
            </a-col>
          </a-row>
          <a-row :gutter="24" :style="{ marginTop: '24px' }">
            <a-col :sm="24" :md="24" :xl="24" :style="{ marginBottom: '24px' }">
              <a-card style="width: 100%" :tab-list="tabListNoTitle" :active-tab-key="noTitleKey"
                @tabChange="key => onTabChange(key, 'noTitleKey')">
                <p v-if="noTitleKey === 'article'">
                  <RdbTable :data="data" />
                </p>
                <p v-else-if="noTitleKey === 'app'">
                  <RdbTable :data="prefixdata" />
                </p>
              </a-card>
            </a-col>
          </a-row>
        </template>
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>
<script setup>
import { ref, computed } from 'vue';
import { ParseRDB, GetParsedKeys, OpenFile } from '../wailsjs/go/main/App'
import { message } from 'ant-design-vue'
import BarChartMemory from './components/BarChartMemory.vue'
import BarChartCount from './components/BarChartCount.vue'
import RdbTable from './components/RdbTable.vue'

const noTitleKey = ref('article');
const current = ref(['1']);
const spinning = ref(false);
const data = ref([])

const bardata = ref([])
const columns = ref([])

const mbardata = ref([])
const mcolumns = ref([])

const ebardata = ref([])
const ecolumns = ref([])

const embardata = ref([])
const emcolumns = ref([])

const totalkeys = ref()
const total_memory = ref()
const file_name = ref()

const file_names = ref([])
const prefixdata = ref([])

const tabListNoTitle = [
  {
    key: 'article',
    tab: 'Top 500 BigKey(按内存)',
  },
  {
    key: 'app',
    tab: 'Top 500 Key前缀(按内存)',
  }
];

const items1 = computed(() =>
  file_names.value.map((name, idx) => ({
    key: String(idx + 1),
    label: name,
    title: name
  }))
)

// 点击调用 Go 后端
const handleClick = async e => {
  try {
    const selectedItem = items1.value.find(item => item.key === e.key);
    console.log(selectedItem.label)
    const result = await GetParsedKeys(selectedItem.label);
    //console.log(result)
    totalkeys.value = result.total_keys
    total_memory.value = result.total_memory
    file_name.value = result.file_name
    bardata.value = Object.values(result.type_stats).map(stat => stat.count)
    //console.log(bardata.value)
    columns.value = Object.keys(result.type_stats)
    //console.log(columns.value)
    mbardata.value = Object.values(result.type_stats).map(stat => stat.memory)
    //console.log(mbardata.value)
    mcolumns.value = Object.keys(result.type_stats)
    //console.log(mcolumns.value)

    ebardata.value = Object.values(result.expire_stats).map(stat => stat.count)
    //console.log(ebardata.value)
    ecolumns.value = Object.keys(result.expire_stats)
    //console.log(ecolumns.value)
    embardata.value = Object.values(result.expire_stats).map(stat => stat.memory)
    //console.log(embardata.value)
    emcolumns.value = Object.keys(result.expire_stats)
    //console.log(emcolumns.value)
    data.value = result.top_keys
    prefixdata.value = result.top_prefix_keys
  } catch (err) {
    console.error("分析失败", err);
  }
};
const openFiles = async () => {
  const newFiles = await OpenFile();
  file_names.value = [...file_names.value, ...newFiles];
  console.log(file_names)
  spinning.value = true
  const result = await ParseRDB(newFiles);
  message.info(result);
  handleClick({ key: items1.value[0].key })
  spinning.value = false
}

const onTabChange = (value, type) => {
  console.log(value, type);
  if (type === 'key') {
    key.value = value;
  } else if (type === 'noTitleKey') {
    noTitleKey.value = value;
  }
};

</script>
<style scoped>
#components-layout-demo-side .logo {
  height: 32px;
  margin: 16px;
  background: rgba(255, 255, 255, 0.3);
}

.site-layout .site-layout-background {
  background: #fff;
}
</style>