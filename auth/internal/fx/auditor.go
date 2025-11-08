package fx

import (
	"sync/atomic"
)

type AuditorDB struct {
	MaxAgeDays         int
	MaxCount           int
	PercentageOfDelete float64
	LogDisabled        bool
}

type Auditor struct {
	Admin AuditorDB
}

type AuditorHolder struct{ v atomic.Value }

func NewAuditorHolder(cfg Auditor) *AuditorHolder {
	h := &AuditorHolder{}
	h.Reload(cfg)
	return h
}

func (h *AuditorHolder) Get() *Auditor {
	return h.v.Load().(*Auditor)
}

func (h *AuditorHolder) Reload(cfg Auditor) {
	svc := &cfg
	h.v.Store(svc)
	return
}
