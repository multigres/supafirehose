import { useRef, useEffect, memo, useMemo } from 'react';
import uPlot from 'uplot';
import 'uplot/dist/uPlot.min.css';

// Chart colors matching our theme
const COLORS = {
  readP50: '#22c55e',
  readP99: '#86efac',
  writeP50: '#3b82f6',
  writeP99: '#93c5fd',
  grid: '#334155',
  axis: '#94a3b8',
  bg: '#1e293b',
};

function getOptions(width) {
  return {
    width: width,
    height: 256,
    padding: [10, 10, 0, 0],
    cursor: {
      show: true,
      drag: { x: false, y: false },
    },
    legend: {
      show: true,
    },
    scales: {
      x: { time: false },
      y: { auto: true, range: [0, null] },
    },
    axes: [
      {
        show: false, // Hide x-axis
      },
      {
        stroke: COLORS.axis,
        grid: { stroke: COLORS.grid, width: 1 },
        ticks: { stroke: COLORS.grid },
        size: 50,
        values: (_, vals) => vals.map((v) => v.toFixed(1)),
      },
    ],
    series: [
      {}, // x-axis series (index)
      {
        label: 'Read P50',
        stroke: COLORS.readP50,
        width: 2,
      },
      {
        label: 'Read P99',
        stroke: COLORS.readP99,
        width: 1,
        dash: [5, 5],
      },
      {
        label: 'Write P50',
        stroke: COLORS.writeP50,
        width: 2,
      },
      {
        label: 'Write P99',
        stroke: COLORS.writeP99,
        width: 1,
        dash: [5, 5],
      },
    ],
  };
}

function LatencyChartInner({ displayData }) {
  const containerRef = useRef(null);
  const chartRef = useRef(null);
  const resizeObserverRef = useRef(null);

  // Transform data for uPlot format: [[x values], [y1 values], [y2 values], ...]
  const chartData = useMemo(() => {
    if (!displayData || displayData.length === 0) {
      return [[], [], [], [], []];
    }

    const len = displayData.length;
    const xData = new Array(len);
    const readP50 = new Array(len);
    const readP99 = new Array(len);
    const writeP50 = new Array(len);
    const writeP99 = new Array(len);

    for (let i = 0; i < len; i++) {
      const m = displayData[i];
      xData[i] = i;
      readP50[i] = m.reads.latency_p50_ms;
      readP99[i] = m.reads.latency_p99_ms;
      writeP50[i] = m.writes.latency_p50_ms;
      writeP99[i] = m.writes.latency_p99_ms;
    }

    return [xData, readP50, readP99, writeP50, writeP99];
  }, [displayData]);

  // Initialize chart
  useEffect(() => {
    if (!containerRef.current) return;

    const width = containerRef.current.clientWidth;
    const opts = getOptions(width);
    const initialData = chartData[0].length > 0 ? chartData : [[0], [0], [0], [0], [0]];

    chartRef.current = new uPlot(opts, initialData, containerRef.current);

    // Handle resize
    resizeObserverRef.current = new ResizeObserver((entries) => {
      for (const entry of entries) {
        if (chartRef.current) {
          chartRef.current.setSize({
            width: entry.contentRect.width,
            height: 256,
          });
        }
      }
    });
    resizeObserverRef.current.observe(containerRef.current);

    return () => {
      if (chartRef.current) {
        chartRef.current.destroy();
        chartRef.current = null;
      }
      if (resizeObserverRef.current) {
        resizeObserverRef.current.disconnect();
      }
    };
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // Only run once on mount

  // Update chart data
  useEffect(() => {
    if (chartRef.current && chartData[0].length > 0) {
      chartRef.current.setData(chartData);
    }
  }, [chartData]);

  return (
    <div className="bg-slate-800 rounded-lg p-6">
      <h2 className="text-lg font-semibold text-white mb-4">Latency (ms)</h2>
      <div className="h-64" ref={containerRef} />
    </div>
  );
}

export const LatencyChart = memo(LatencyChartInner);
