package logger

import (
	"sync/atomic"
)

type Vitals struct {
	JobsProcessed     int64
	ParseSuccess      int64
	ParseFailure      int64
	DuplicateFound    int64
	InsertSuccess     int64
	InsertFailure     int64
	TransitionSuccess int64
}

var GlobalVitals Vitals

func (v *Vitals) IncJobs()          { atomic.AddInt64(&v.JobsProcessed, 1) }
func (v *Vitals) IncParseSuccess()  { atomic.AddInt64(&v.ParseSuccess, 1) }
func (v *Vitals) IncParseFailure()  { atomic.AddInt64(&v.ParseFailure, 1) }
func (v *Vitals) IncDuplicate()     { atomic.AddInt64(&v.DuplicateFound, 1) }
func (v *Vitals) IncInsertSuccess() { atomic.AddInt64(&v.InsertSuccess, 1) }
func (v *Vitals) IncInsertFailure() { atomic.AddInt64(&v.InsertFailure, 1) }
func (v *Vitals) IncTransition()    { atomic.AddInt64(&v.TransitionSuccess, 1) }

func (v *Vitals) GetSnapshot() map[string]int64 {
	return map[string]int64{
		"jobs_processed":     atomic.LoadInt64(&v.JobsProcessed),
		"parse_success":      atomic.LoadInt64(&v.ParseSuccess),
		"parse_failure":      atomic.LoadInt64(&v.ParseFailure),
		"duplicate_found":    atomic.LoadInt64(&v.DuplicateFound),
		"insert_success":     atomic.LoadInt64(&v.InsertSuccess),
		"insert_failure":     atomic.LoadInt64(&v.InsertFailure),
		"transition_success": atomic.LoadInt64(&v.TransitionSuccess),
	}
}
