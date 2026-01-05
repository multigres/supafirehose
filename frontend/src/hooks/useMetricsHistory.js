import { useState, useCallback, useRef } from 'react';

const MAX_HISTORY_SIZE = 600; // 60 seconds at 100ms intervals
const DISPLAY_SIZE = 100; // Only show last 100 points in charts

// Efficient circular buffer implementation
class CircularBuffer {
  constructor(maxSize) {
    this.maxSize = maxSize;
    this.buffer = new Array(maxSize);
    this.head = 0;
    this.size = 0;
  }

  push(item) {
    this.buffer[this.head] = item;
    this.head = (this.head + 1) % this.maxSize;
    if (this.size < this.maxSize) {
      this.size++;
    }
  }

  // Get last N items (most recent)
  getLastN(n) {
    const count = Math.min(n, this.size);
    const result = new Array(count);

    for (let i = 0; i < count; i++) {
      // Calculate index going backwards from head
      const idx = (this.head - count + i + this.maxSize) % this.maxSize;
      result[i] = this.buffer[idx];
    }

    return result;
  }

  clear() {
    this.head = 0;
    this.size = 0;
  }

  getSize() {
    return this.size;
  }
}

export function useMetricsHistory() {
  // Use ref for the buffer to avoid recreating it
  const bufferRef = useRef(new CircularBuffer(MAX_HISTORY_SIZE));

  // Version counter to trigger re-renders when needed
  const [version, setVersion] = useState(0);

  // Cache the display data to avoid recomputation
  const displayCacheRef = useRef({ version: -1, data: [] });

  const addMetric = useCallback((metric) => {
    bufferRef.current.push(metric);
    // Increment version to trigger re-render
    setVersion((v) => v + 1);
  }, []);

  const clearHistory = useCallback(() => {
    bufferRef.current.clear();
    displayCacheRef.current = { version: -1, data: [] };
    setVersion(0);
  }, []);

  // Get display data (last 100 items) with caching
  const getDisplayData = useCallback(() => {
    if (displayCacheRef.current.version === version) {
      return displayCacheRef.current.data;
    }

    const data = bufferRef.current.getLastN(DISPLAY_SIZE);
    displayCacheRef.current = { version, data };
    return data;
  }, [version]);

  return {
    // Return a getter function instead of the array directly
    // This avoids creating new array references on every render
    getDisplayData,
    addMetric,
    clearHistory,
    version, // Expose version for dependency tracking
  };
}
