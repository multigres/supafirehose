import { useState, useEffect, useMemo, useRef } from 'react';
import { useWebSocket } from './hooks/useWebSocket';
import { useMetricsHistory } from './hooks/useMetricsHistory';
import { getStatus, getScenarios, updateConfig, start, stop, reset } from './api/client';
import { ConnectionStatus } from './components/ConnectionStatus';
import { ControlPanel } from './components/ControlPanel';
import { StatsPanel } from './components/StatsPanel';
import { LatencyChart } from './components/LatencyChart';
import { ThroughputChart } from './components/ThroughputChart';
import { ErrorList } from './components/ErrorList';

function App() {
  const [config, setConfig] = useState({
    connections: 10,
    read_qps: 100,
    write_qps: 10,
    churn_rate: 0,
    scenario: 'simple',
    custom_table: '',
  });
  const [scenarios, setScenarios] = useState([]);
  const [running, setRunning] = useState(false);
  const [latestMetrics, setLatestMetrics] = useState(null);

  const { isConnected, lastMessage } = useWebSocket('/ws/metrics');
  const { getDisplayData, addMetric, clearHistory, version } = useMetricsHistory();

  // Track the last processed message to avoid duplicate processing
  const lastProcessedRef = useRef(null);

  // Memoize display data to prevent unnecessary re-computations
  const displayData = useMemo(() => getDisplayData(), [getDisplayData]);

  // Fetch initial status and scenarios
  useEffect(() => {
    getStatus().then((status) => {
      setRunning(status.running);
      setConfig(status.config);
    }).catch(console.error);

    getScenarios().then((data) => {
      setScenarios(data.scenarios || []);
    }).catch(console.error);
  }, []);

  // Update metrics from WebSocket - use ref to track changes
  // This pattern is intentional: we need to sync external WebSocket data to React state
  useEffect(() => {
    if (lastMessage && lastMessage !== lastProcessedRef.current) {
      lastProcessedRef.current = lastMessage;
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setLatestMetrics(lastMessage);
      addMetric(lastMessage);
    }
  }, [lastMessage, addMetric]);

  const handleConfigChange = async (newConfig) => {
    try {
      const response = await updateConfig(newConfig);
      setConfig(response.config);
    } catch (error) {
      console.error('Failed to update config:', error);
    }
  };

  const handleStart = async () => {
    try {
      await start();
      setRunning(true);
    } catch (error) {
      console.error('Failed to start:', error);
    }
  };

  const handleStop = async () => {
    try {
      await stop();
      setRunning(false);
    } catch (error) {
      console.error('Failed to stop:', error);
    }
  };

  const handleReset = async () => {
    try {
      await reset();
      clearHistory();
      setLatestMetrics(null);
    } catch (error) {
      console.error('Failed to reset:', error);
    }
  };

  return (
    <div className="min-h-screen bg-slate-900 text-slate-100">
      {/* Header */}
      <header className="bg-slate-800 border-b border-slate-700 px-6 py-4">
        <div className="max-w-7xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3">
            <img src="/logo.png" alt="SupaFirehose" className="h-10 w-10" />
            <h1 className="text-2xl font-bold text-white">SUPAFIREHOSE</h1>
          </div>
          <div className="flex items-center gap-4">
            <ConnectionStatus isConnected={isConnected} />
            <div
              className={`px-3 py-1 rounded-full text-sm font-medium ${
                running
                  ? 'bg-green-500/20 text-green-400'
                  : 'bg-slate-600/50 text-slate-400'
              }`}
            >
              {running ? 'RUNNING' : 'STOPPED'}
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto p-6 space-y-6">
        {/* Top Section: Controls + Charts */}
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
          {/* Control Panel */}
          <div className="lg:col-span-1">
            <ControlPanel
              config={config}
              scenarios={scenarios}
              running={running}
              onConfigChange={handleConfigChange}
              onStart={handleStart}
              onStop={handleStop}
              onReset={handleReset}
            />
          </div>

          {/* Charts */}
          <div className="lg:col-span-3 space-y-6">
            <LatencyChart displayData={displayData} version={version} />
            <ThroughputChart displayData={displayData} version={version} />
          </div>
        </div>

        {/* Bottom Section: Stats + Errors */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <StatsPanel metrics={latestMetrics} />
          <ErrorList errors={latestMetrics?.recent_errors} />
        </div>
      </main>
    </div>
  );
}

export default App;
