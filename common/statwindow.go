package common

import (
	"fmt"
	"sort"
	"strconv"
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

func (h *StatHistogram) Xile(x int) *[]StatHistogramEntry {

	firstValidVal := -1
	for i, v := range *h {
		if v > 0 {
			firstValidVal = i
			break
		}
	}
	totalValid := len(*h) - firstValidVal
	fmt.Println("FIRST VALID " + strconv.Itoa(firstValidVal))
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
		for hy := 0; hy < binSize && arrPtr < len(*h); hy++ {
			ms := (*h)[arrPtr].Microseconds()
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
		ptr:    0,
	}
}

func (sw *StatWindow) Add(t time.Duration) {
	sw.Lock()
	defer sw.Unlock()
	sw.window[sw.ptr] = t
	if sw.ptr >= len(sw.window) {
		sw.ptr = 0
	} else {
		sw.ptr = sw.ptr + 1
	}
}

func (sw *StatWindow) MakeHistogram() *StatHistogram {
	arr := make(StatHistogram, len(sw.window))
	copy(arr, sw.window)
	sort.Slice(arr, func(i, j int) bool {
		return arr[i] < arr[j]
	})
	return &arr
}
