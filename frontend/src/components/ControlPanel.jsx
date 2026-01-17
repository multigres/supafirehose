import { useState, useEffect } from 'react';

export function ControlPanel({ config, scenarios, running, onConfigChange, onStart, onStop, onReset }) {
  const [localConfig, setLocalConfig] = useState(config);

  useEffect(() => {
    setLocalConfig(config);
  }, [config]);

  const handleSliderChange = (key, value) => {
    const newConfig = { ...localConfig, [key]: value };
    setLocalConfig(newConfig);
    onConfigChange(newConfig);
  };

  const handleScenarioChange = (scenario) => {
    const newConfig = { ...localConfig, scenario };
    // Clear custom_table if not custom scenario
    if (scenario !== 'custom') {
      newConfig.custom_table = '';
    }
    setLocalConfig(newConfig);
    onConfigChange(newConfig);
  };

  const handleCustomTableChange = (customTable) => {
    const newConfig = { ...localConfig, custom_table: customTable };
    setLocalConfig(newConfig);
    // Don't auto-submit custom table - wait for blur or enter
  };

  const submitCustomTable = () => {
    onConfigChange(localConfig);
  };

  return (
    <div className="bg-slate-800 rounded-lg p-6 space-y-6">
      <h2 className="text-lg font-semibold text-white">Control Panel</h2>

      <div className="space-y-4">
        {/* Scenario Selector */}
        <div className="space-y-2">
          <label className="block text-sm text-slate-300">Schema Scenario</label>
          <select
            value={localConfig.scenario || 'simple'}
            onChange={(e) => handleScenarioChange(e.target.value)}
            className="w-full bg-slate-700 border border-slate-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            {scenarios.map((s) => (
              <option key={s.name} value={s.name}>
                {s.name} - {s.description}
              </option>
            ))}
          </select>
        </div>

        {/* Custom Table Input */}
        {localConfig.scenario === 'custom' && (
          <div className="space-y-2">
            <label className="block text-sm text-slate-300">Table Name</label>
            <input
              type="text"
              value={localConfig.custom_table || ''}
              onChange={(e) => handleCustomTableChange(e.target.value)}
              onBlur={submitCustomTable}
              onKeyDown={(e) => e.key === 'Enter' && submitCustomTable()}
              placeholder="schema.table_name"
              className="w-full bg-slate-700 border border-slate-600 rounded-lg px-3 py-2 text-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 placeholder-slate-500"
            />
          </div>
        )}

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
          max={500000}
          step={1000}
          onChange={(v) => handleSliderChange('read_qps', v)}
        />

        <SliderControl
          label="Write QPS"
          value={localConfig.write_qps}
          min={0}
          max={500000}
          step={1000}
          onChange={(v) => handleSliderChange('write_qps', v)}
        />

        <SliderControl
          label="Churn Rate"
          value={localConfig.churn_rate}
          min={0}
          max={10000}
          step={100}
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
