import { memo } from 'react';
import { formatNumber, formatPercent } from '../utils/formatting';

function StatsPanelInner({ metrics }) {
  if (!metrics) {
    return (
      <div className="bg-slate-800 rounded-lg p-6">
        <h2 className="text-lg font-semibold text-white mb-4">Statistics</h2>
        <p className="text-slate-400">Waiting for data...</p>
      </div>
    );
  }

  return (
    <div className="bg-slate-800 rounded-lg p-6">
      <h2 className="text-lg font-semibold text-white mb-4">Statistics</h2>
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard
          label="Total Queries"
          value={formatNumber(metrics.totals.queries)}
        />
        <StatCard
          label="Total Errors"
          value={formatNumber(metrics.totals.errors)}
          variant={metrics.totals.errors > 0 ? 'error' : 'default'}
        />
        <StatCard
          label="Error Rate"
          value={formatPercent(metrics.totals.error_rate)}
          variant={metrics.totals.error_rate > 0.01 ? 'error' : 'default'}
        />
        <StatCard
          label="Active Connections"
          value={`${metrics.pool.active_connections}`}
        />
      </div>
    </div>
  );
}

const StatCard = memo(function StatCard({ label, value, variant = 'default' }) {
  const valueColor = variant === 'error' ? 'text-red-400' : 'text-white';

  return (
    <div className="bg-slate-700/50 rounded-lg p-4">
      <div className="text-sm text-slate-400 mb-1">{label}</div>
      <div className={`text-2xl font-mono font-bold ${valueColor}`}>{value}</div>
    </div>
  );
});

export const StatsPanel = memo(StatsPanelInner);
