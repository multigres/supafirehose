import { useRef, useEffect, memo, useMemo } from 'react';
import uPlot from 'uplot';
import 'uplot/dist/uPlot.min.css';

// Chart colors matching our theme
const COLORS = {
  readErrors: '#ef4444',
  writeErrors: '#f97316',
  grid: '#334155',
  axis: '#94a3b8',
};

function getOptions(width) {
  return {
    width: width,
    height: 192,
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
        size: 40,
      },
    ],
    series: [
      {}, // x-axis series (index)
      {
        label: 'Read Errors',
        stroke: COLORS.readErrors,
        width: 2,
      },
      {
        label: 'Write Errors',
        stroke: COLORS.writeErrors,
        width: 2,
      },
    ],
  };
}

function ErrorChartInner({ displayData }) {
  const containerRef = useRef(null);
  const chartRef = useRef(null);
  const resizeObserverRef = useRef(null);

  // Transform data and check for errors
  const { chartData, hasErrors } = useMemo(() => {
    if (!displayData || displayData.length === 0) {
      return { chartData: [[], [], []], hasErrors: false };
    }

    const len = displayData.length;
    const xData = new Array(len);
    const readErrors = new Array(len);
    const writeErrors = new Array(len);
    let foundErrors = false;

    for (let i = 0; i < len; i++) {
      const m = displayData[i];
      xData[i] = i;
      readErrors[i] = m.reads.errors;
      writeErrors[i] = m.writes.errors;
      if (m.reads.errors > 0 || m.writes.errors > 0) {
        foundErrors = true;
      }
    }

    return {
      chartData: [xData, readErrors, writeErrors],
      hasErrors: foundErrors,
    };
  }, [displayData]);

  // Initialize chart when errors first appear
  useEffect(() => {
    if (!containerRef.current || !hasErrors) return;

    // Destroy existing chart if any
    if (chartRef.current) {
      chartRef.current.destroy();
      chartRef.current = null;
    }

    const width = containerRef.current.clientWidth;
    const opts = getOptions(width);

    chartRef.current = new uPlot(opts, chartData, containerRef.current);

    // Handle resize
    resizeObserverRef.current = new ResizeObserver((entries) => {
      for (const entry of entries) {
        if (chartRef.current) {
          chartRef.current.setSize({
            width: entry.contentRect.width,
            height: 192,
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
  }, [hasErrors, chartData]);

  // Update chart data
  useEffect(() => {
    if (chartRef.current && chartData[0].length > 0 && hasErrors) {
      chartRef.current.setData(chartData);
    }
  }, [chartData, hasErrors]);

  return (
    <div className="bg-slate-800 rounded-lg p-6">
      <h2 className="text-lg font-semibold text-white mb-4">Errors</h2>
      <div className="h-48">
        {hasErrors ? (
          <div ref={containerRef} className="w-full h-full" />
        ) : (
          <div className="h-full flex items-center justify-center text-slate-500">
            No errors
          </div>
        )}
      </div>
    </div>
  );
}

export const ErrorChart = memo(ErrorChartInner);
