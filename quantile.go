package main

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

type calculator struct {
	lock    sync.Mutex
	results []time.Duration
}

func (c *calculator) timer() *timer {
	return &timer{
		calc:  c,
		start: time.Now(),
	}
}

func (c *calculator) quantiles() []string {
	sort.Slice(c.results, func(i, j int) bool {
		return c.results[i] < c.results[j]
	})
	return []string{
		fmt.Sprintf("%.2f", c.results[0].Seconds()*1000),
		fmt.Sprintf("%.2f", c.results[int(math.Floor(float64(len(c.results))*0.5))].Seconds()*1000),
		fmt.Sprintf("%.2f", c.results[int(math.Floor(float64(len(c.results))*0.9))].Seconds()*1000),
		fmt.Sprintf("%.2f", c.results[len(c.results)-1].Seconds()*1000),
	}
}

type timer struct {
	calc  *calculator
	start time.Time
}

func (t *timer) done() {
	t.calc.lock.Lock()
	defer t.calc.lock.Unlock()
	t.calc.results = append(t.calc.results, time.Since(t.start))
}
