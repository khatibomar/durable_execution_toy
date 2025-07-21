package journal

import (
	"crypto/sha256"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"
)

type JournalEntryStatus int32

const (
	Running JournalEntryStatus = iota
	Failed
	Completed
)

type JournalEntry struct {
	StepIndex int                `json:"step_index"`
	Status    JournalEntryStatus `json:"status"`
	Timestamp time.Time          `json:"timestamp"`
	FuncHash  string             `json:"func_hash,omitempty"`
	RunResult string             `json:"run_result,omitempty"`
	ErrorMsg  string             `json:"error_msg,omitempty"`
}

type Journal struct {
	mu           sync.RWMutex
	entries      []JournalEntry
	ExecutionID  string    `json:"execution_id"`
	StartTime    time.Time `json:"start_time"`
	LastModified time.Time `json:"last_modified"`
}

func NewJournal() *Journal {
	return &Journal{
		entries:      make([]JournalEntry, 0),
		StartTime:    time.Now(),
		LastModified: time.Now(),
	}
}

func (j *Journal) GetFunctionHash(fn any) string {
	fnValue := reflect.ValueOf(fn)
	fnPtr := fnValue.Pointer()
	fnName := runtime.FuncForPC(fnPtr).Name()

	pc, file, line, ok := runtime.Caller(4)
	var sourceLocation string
	if ok {
		caller := runtime.FuncForPC(pc)
		if caller != nil {
			sourceLocation = fmt.Sprintf("%s:%s:%d", caller.Name(), file, line)
		}
	}

	hashInput := fmt.Sprintf("%d_%s_%s", fnPtr, fnName, sourceLocation)
	hash := sha256.Sum256([]byte(hashInput))
	return fmt.Sprintf("%x", hash[:8])
}

func (j *Journal) GetCompletedEntryByHash(fn any) (*JournalEntry, bool) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	funcHash := j.GetFunctionHash(fn)

	for i := range j.entries {
		entry := &j.entries[i]
		if entry.FuncHash == funcHash && entry.Status == Completed {
			return entry, true
		}
	}

	return nil, false
}

func (j *Journal) HasEntryByHash(fn any) bool {
	j.mu.RLock()
	defer j.mu.RUnlock()

	funcHash := j.GetFunctionHash(fn)

	for _, entry := range j.entries {
		if entry.FuncHash == funcHash {
			return true
		}
	}

	return false
}

func (j *Journal) StartStep(fn any) int {
	j.mu.Lock()
	defer j.mu.Unlock()

	stepNumber := len(j.entries) + 1
	funcHash := j.GetFunctionHash(fn)

	entry := JournalEntry{
		StepIndex: stepNumber,
		Status:    Running,
		Timestamp: time.Now(),
		FuncHash:  funcHash,
	}

	j.entries = append(j.entries, entry)
	j.LastModified = time.Now()

	return stepNumber
}

func (j *Journal) CompleteLastStep(result string) {
	j.mu.Lock()
	defer j.mu.Unlock()

	if len(j.entries) == 0 {
		fmt.Printf("[%s] Warning: No entries to complete\n", j.ExecutionID)
		return
	}

	lastEntry := &j.entries[len(j.entries)-1]
	lastEntry.Status = Completed
	lastEntry.RunResult = result
	lastEntry.ErrorMsg = ""
	j.LastModified = time.Now()
}

func (j *Journal) FailLastStep(err error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	if len(j.entries) == 0 {
		fmt.Printf("[%s] Warning: No entries to fail\n", j.ExecutionID)
		return
	}

	lastEntry := &j.entries[len(j.entries)-1]
	lastEntry.Status = Failed
	lastEntry.ErrorMsg = err.Error()
	j.LastModified = time.Now()
}

func (j *Journal) GetEntries() []JournalEntry {
	j.mu.RLock()
	defer j.mu.RUnlock()

	entries := make([]JournalEntry, len(j.entries))
	copy(entries, j.entries)
	return entries
}

func (j *Journal) GetLastEntry() *JournalEntry {
	j.mu.RLock()
	defer j.mu.RUnlock()
	if len(j.entries) == 0 {
		return nil
	}

	return &j.entries[len(j.entries)-1]
}
