export function formatNumber(num) {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(2) + 'M';
  }
  if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K';
  }
  return num.toFixed(0);
}

export function formatQPS(qps) {
  if (qps >= 1000) {
    return (qps / 1000).toFixed(1) + 'K';
  }
  return qps.toFixed(0);
}

export function formatLatency(ms) {
  if (ms >= 1000) {
    return (ms / 1000).toFixed(2) + 's';
  }
  return ms.toFixed(1) + 'ms';
}

export function formatPercent(rate) {
  return (rate * 100).toFixed(3) + '%';
}

export function formatBytes(bytes) {
  if (bytes === 0 || bytes === undefined || bytes === null) {
    return '0 B';
  }
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const k = 1024;
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  const value = bytes / Math.pow(k, i);
  return value.toFixed(i > 0 ? 2 : 0) + ' ' + units[i];
}
