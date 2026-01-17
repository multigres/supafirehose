const API_BASE = '/api';

export async function getStatus() {
  const response = await fetch(`${API_BASE}/status`);
  return response.json();
}

export async function getScenarios() {
  const response = await fetch(`${API_BASE}/scenarios`);
  return response.json();
}

export async function updateConfig(config) {
  const response = await fetch(`${API_BASE}/config`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(config),
  });
  return response.json();
}

export async function start() {
  const response = await fetch(`${API_BASE}/start`, { method: 'POST' });
  return response.json();
}

export async function stop() {
  const response = await fetch(`${API_BASE}/stop`, { method: 'POST' });
  return response.json();
}

export async function reset() {
  const response = await fetch(`${API_BASE}/reset`, { method: 'POST' });
  return response.json();
}
