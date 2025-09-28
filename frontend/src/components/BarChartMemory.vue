<template>
    <v-chart :option="chartOptions" autoresize style="height: 400px;" />
</template>

<script setup>
import { computed, defineProps } from "vue";
import { use } from "echarts/core";
import { CanvasRenderer } from "echarts/renderers";
import { BarChart } from "echarts/charts";
import { GridComponent, TitleComponent, TooltipComponent } from "echarts/components";
import VChart from "vue-echarts";

// 注册 ECharts 模块
use([CanvasRenderer, BarChart, GridComponent, TitleComponent, TooltipComponent]);

const props = defineProps({
  barData: {
    type: Array,
    default: () => []
  },
  columns:{
    type: Array,
    default: () => []
  }
});
function formatBytes(bytes) {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i];
}

const chartOptions = computed(() => ({
  title: {
    left: "center",
  },
  tooltip: {
    trigger: "axis",
    formatter: function (params) {
      return params[0].name + ': ' + formatBytes(params[0].value);
    }
  },
  xAxis: {
    type: "category",
    data: props.columns,
  },
  yAxis: {
    type: "value",
    axisLabel: {
    formatter: function (value) {
        return formatBytes(value);
      }
    }
  },
  series: [
    {
      name: "销量",
      type: "bar",
      data: props.barData,
      itemStyle: {
        color: "#1890ff", // 柱子颜色
      },
    },
  ],
}))
</script>
