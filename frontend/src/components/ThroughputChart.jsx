import { useRef, useEffect, memo, useMemo } from 'react';
import uPlot from 'uplot';
import 'uplot/dist/uPlot.min.css';
import { formatQPS } from '../utils/formatting';

// Chart colors matching our theme
const COLORS = {
  reads: '#22c55e',
  writes: '#3b82f6',
  grid: '#334155',
  axis: '#94a3b8',
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
        size: 60,
        values: (_, vals) => vals.map((v) => formatQPS(v)),
      },
    ],
    series: [
      {}, // x-axis series (index)
      {
        label: 'Reads',
        stroke: COLORS.reads,
        fill: COLORS.reads + '99', // Add transparency
        width: 2,
      },
      {
        label: 'Writes',
        stroke: COLORS.writes,
        fill: COLORS.writes + '99',
        width: 2,
      },
    ],
  };
}

function ThroughputChartInner({ displayData }) {
  const containerRef = useRef(null);
  const chartRef = useRef(null);
  const resizeObserverRef = useRef(null);

  // Transform data for uPlot format: [[x values], [reads], [writes]]
  const chartData = useMemo(() => {
    if (!displayData || displayData.length === 0) {
      return [[], [], []];
    }

    const len = displayData.length;
    const xData = new Array(len);
    const reads = new Array(len);
    const writes = new Array(len);

    for (let i = 0; i < len; i++) {
      const m = displayData[i];
      xData[i] = i;
      reads[i] = m.reads.qps;
      writes[i] = m.writes.qps;
    }

    return [xData, reads, writes];
  }, [displayData]);

  // Initialize chart
  useEffect(() => {
    if (!containerRef.current) return;

    const width = containerRef.current.clientWidth;
    const opts = getOptions(width);
    const initialData = chartData[0].length > 0 ? chartData : [[0], [0], [0]];

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
      <h2 className="text-lg font-semibold text-white mb-4">Throughput (QPS)</h2>
      <div className="h-64" ref={containerRef} />
    </div>
  );
}

export const ThroughputChart = memo(ThroughputChartInner);
