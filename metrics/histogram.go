package metrics

import (
	"sync/atomic"
	"time"
)

// Bucket upper bounds in microseconds.
// Chosen for good resolution across typical DB query latencies (0.1ms–5s).
// The final bucket catches everything above the last bound.
var bucketBoundsUs = [numBounds]int64{
	100, 250, 500, 750,         // 0.1ms – 0.75ms
	1_000, 1_500, 2_000, 3_000, // 1ms – 3ms
	5_000, 7_500, 10_000, 15_000, // 5ms – 15ms
	20_000, 30_000, 50_000, 75_000, // 20ms – 75ms
	100_000, 200_000, 500_000, // 100ms – 500ms
	1_000_000, 5_000_000, // 1s – 5s
}

const numBounds = 21
const numBuckets = numBounds + 1

// Histogram collects latency samples in pre-defined buckets using atomic counters.
// Record is completely lock-free.
type Histogram struct {
	buckets [numBuckets]atomic.Int64
	sum     atomic.Int64 // total latency in microseconds
}

// HistogramSnapshot holds computed percentiles from a histogram window.
type HistogramSnapshot struct {
	P50   float64
	P99   float64
	Avg   float64
	Count int
}

// NewHistogram creates a new histogram.
func NewHistogram() *Histogram {
	return &Histogram{}
}

// Record adds a latency sample. Lock-free — uses only atomic operations.
func (h *Histogram) Record(d time.Duration) {
	us := d.Microseconds()

	// Linear scan — fast for the common case (low-latency queries hit early buckets)
	idx := numBounds // overflow bucket
	for i := 0; i < numBounds; i++ {
		if us < bucketBoundsUs[i] {
			idx = i
			break
		}
	}

	h.buckets[idx].Add(1)
	h.sum.Add(us)
}

// SnapshotAndReset reads all bucket counts, computes percentiles, and resets.
func (h *Histogram) SnapshotAndReset() HistogramSnapshot {
	// Swap all counters to zero and read their values.
	// Not perfectly atomic across all buckets, but the error is bounded
	// to samples recorded during the few nanoseconds of the swap loop.
	var counts [numBuckets]int64
	var totalCount int64
	for i := range counts {
		counts[i] = h.buckets[i].Swap(0)
		totalCount += counts[i]
	}
	totalSum := h.sum.Swap(0)

	if totalCount == 0 {
		return HistogramSnapshot{}
	}

	return HistogramSnapshot{
		P50:   percentileFromBuckets(counts[:], totalCount, 0.50),
		P99:   percentileFromBuckets(counts[:], totalCount, 0.99),
		Avg:   float64(totalSum) / float64(totalCount) / 1000.0, // µs → ms
		Count: int(totalCount),
	}
}

// percentileFromBuckets estimates a percentile by linear interpolation
// within the bucket that contains the target rank.
func percentileFromBuckets(counts []int64, total int64, p float64) float64 {
	target := float64(total) * p
	var cumulative float64

	for i, c := range counts {
		cumulative += float64(c)
		if cumulative >= target {
			loUs := bucketLowerUs(i)
			hiUs := bucketUpperUs(i)
			prev := cumulative - float64(c)
			frac := 0.0
			if c > 0 {
				frac = (target - prev) / float64(c)
			}
			return (float64(loUs) + frac*float64(hiUs-loUs)) / 1000.0 // µs → ms
		}
	}

	// Should not reach here; return last bound as fallback.
	return float64(bucketBoundsUs[numBounds-1]) / 1000.0
}

func bucketLowerUs(i int) int64 {
	if i == 0 {
		return 0
	}
	return bucketBoundsUs[i-1]
}

func bucketUpperUs(i int) int64 {
	if i >= numBounds {
		return bucketBoundsUs[numBounds-1] * 2 // overflow estimate
	}
	return bucketBoundsUs[i]
}
