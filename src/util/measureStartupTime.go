package util

import (
	"log"
	"time"
)

type MST struct {
	time time.Time
}

func (m *MST) StartTimer() {
	m.time = time.Now()
}

func (m *MST) ElapsedTime() {
	log.Printf("Total startup time : %v\n", time.Since(m.time))
}
