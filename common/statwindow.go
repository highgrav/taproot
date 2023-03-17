package common

import (
	"sort"
	"sync"
	"time"
)

type StatHistogram []time.Duration
type StatHistogramEntry struct {
	Quantile int   `json:"quantile"`
	Minµs    int64 `json:"minMicros"`
	Maxµs    int64 `json:"maxMicros"`
	Avgµs    int64 `json:"avgMicros"`
	Count    int   `json:"count"`
}

/*
Xile() takes a slice of time.Duration values and partition them into x number of bins, where each bin represents a range of values within the data. The function then calculates statistical information for each bin, including the minimum, maximum, and average values, as well as the number of values in each bin.
*/
func (h StatHistogram) Xile(x int) *[]StatHistogramEntry {

	firstValidVal := -1
	for i, v := range h {
		if v > 0 {
			firstValidVal = i
			break
		}
	}
	totalValid := len(h) - firstValidVal
	binSize := int(totalValid / x)
	arrPtr := firstValidVal
	if x > int(totalValid/binSize) {
		x = int(totalValid / binSize)
	}
	gram := make([]StatHistogramEntry, 0)
	for hx := 0; hx < x; hx++ {
		she := StatHistogramEntry{
			Quantile: hx,
			Minµs:    9999999999,
			Maxµs:    -9999999999,
			Avgµs:    0,
			Count:    0,
		}
		// This end-of-array check isn't necessary, but I'm paranoid enough to keep it in.
		for hy := 0; hy < binSize && arrPtr < len(h); hy++ {
			ms := h[arrPtr].Microseconds()
			if ms > she.Maxµs {
				she.Maxµs = ms
			}
			if ms < she.Minµs {
				she.Minµs = ms
			}
			she.Avgµs = ((she.Avgµs * int64(she.Count)) + ms) / int64(she.Count+1)
			she.Count = she.Count + 1
			arrPtr = arrPtr + 1
		}
		gram = append(gram, she)
	}
	return &gram
}

// StatWindow is a circular buffer that calculates histograms. Used for capturing Pxx latency in the
// per-endpoint metrics middleware.
type StatWindow struct {
	sync.Mutex
	name   string
	window []time.Duration
	ptr    int // points to the current head
}

func NewStatWindow(name string, length int) *StatWindow {
	return &StatWindow{
		name:   name,
		window: make([]time.Duration, length),
		ptr:    -1,
	}
}

func (sw *StatWindow) Add(t time.Duration) {
	sw.Lock()
	defer sw.Unlock()
	if sw.ptr >= len(sw.window)-1 || sw.ptr < 0 {
		sw.ptr = 0
	} else {
		sw.ptr = sw.ptr + 1
	}
	sw.window[sw.ptr] = t

}

func (sw *StatWindow) MakeHistogram() *StatHistogram {
	arr := make(StatHistogram, len(sw.window))
	copy(arr, sw.window)
	sort.Slice(arr, func(i, j int) bool {
		return arr[i] < arr[j]
	})
	return &arr
}
