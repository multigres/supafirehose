import { useState, useEffect } from 'react';

export function ControlPanel({ config, running, onConfigChange, onStart, onStop, onReset }) {
  const [localConfig, setLocalConfig] = useState(config);

  useEffect(() => {
    setLocalConfig(config);
  }, [config]);

  const handleSliderChange = (key, value) => {
    const newConfig = { ...localConfig, [key]: value };
    setLocalConfig(newConfig);
    onConfigChange(newConfig);
  };

  return (
    <div className="bg-slate-800 rounded-lg p-6 space-y-6">
      <h2 className="text-lg font-semibold text-white">Control Panel</h2>

      <div className="space-y-4">
        <SliderControl
          label="Connections"
          value={localConfig.connections}
          min={1}
          max={20000}
          step={10}
          onChange={(v) => handleSliderChange('connections', v)}
        />

        <SliderControl
          label="Read QPS"
          value={localConfig.read_qps}
          min={0}
          max={25000}
          step={100}
          onChange={(v) => handleSliderChange('read_qps', v)}
        />

        <SliderControl
          label="Write QPS"
          value={localConfig.write_qps}
          min={0}
          max={5000}
          step={50}
          onChange={(v) => handleSliderChange('write_qps', v)}
        />

        <SliderControl
          label="Churn Rate"
          value={localConfig.churn_rate}
          min={0}
          max={2000}
          step={10}
          onChange={(v) => handleSliderChange('churn_rate', v)}
          hint="connections/sec"
        />
      </div>

      <div className="flex gap-3">
        {running ? (
          <button
            onClick={onStop}
            className="flex-1 bg-red-600 hover:bg-red-700 text-white py-2 px-4 rounded-lg font-medium transition-colors"
          >
            Stop
          </button>
        ) : (
          <button
            onClick={onStart}
            className="flex-1 bg-green-600 hover:bg-green-700 text-white py-2 px-4 rounded-lg font-medium transition-colors"
          >
            Start
          </button>
        )}
        <button
          onClick={onReset}
          className="flex-1 bg-slate-600 hover:bg-slate-700 text-white py-2 px-4 rounded-lg font-medium transition-colors"
        >
          Reset
        </button>
      </div>
    </div>
  );
}

function SliderControl({ label, value, min, max, step = 1, onChange, hint }) {
  const formatValue = (v) => {
    if (v >= 1000000) {
      return (v / 1000000).toFixed(1) + 'M';
    }
    if (v >= 1000) {
      return (v / 1000).toFixed(v >= 10000 ? 0 : 1) + 'K';
    }
    return v;
  };

  return (
    <div className="space-y-2">
      <div className="flex justify-between text-sm">
        <span className="text-slate-300">
          {label}
          {hint && <span className="text-slate-500 ml-1">({hint})</span>}
        </span>
        <span className="text-white font-mono">{formatValue(value)}</span>
      </div>
      <input
        type="range"
        min={min}
        max={max}
        step={step}
        value={value}
        onChange={(e) => onChange(parseInt(e.target.value, 10))}
        className="w-full h-2 bg-slate-700 rounded-lg appearance-none cursor-pointer accent-blue-500"
      />
    </div>
  );
}
