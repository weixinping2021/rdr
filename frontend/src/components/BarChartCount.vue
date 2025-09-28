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

const chartOptions = computed(() => ({
  title: {
    left: "center",
  },
  tooltip: {
    trigger: "axis",
  },
  xAxis: {
    type: "category",
    data: props.columns,
  },
  yAxis: {
    type: "value",
  },
  series: [
    {
      name: "数量",
      type: "bar",
      data: props.barData,
      itemStyle: {
        color: "#1890ff", // 柱子颜色
      },
    },
  ],
}))
</script>
