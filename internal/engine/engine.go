package engine

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/khatibomar/durable_execution_toy/internal/journal"
)

type Void struct{}

type Context interface {
	Run(fn func() (any, error)) (any, error)
}

type ExecutionContext struct {
	journal     *journal.Journal
	executionID string
}

func NewContext() Context {
	executionID := uuid.New().String()
	j := journal.NewJournal()
	j.ExecutionID = executionID
	return &ExecutionContext{
		journal:     j,
		executionID: executionID,
	}
}

func (ec *ExecutionContext) Run(fn func() (any, error)) (any, error) {
	if existingResult, exists := ec.journal.GetCompletedEntryByHash(fn); exists {
		var result any
		if existingResult.RunResult != "" {
			if existingResult.RunResult == `{}` || existingResult.RunResult == "null" {
				result = Void{}
			} else if err := json.Unmarshal([]byte(existingResult.RunResult), &result); err != nil {
				result = existingResult.RunResult
			}
		}
		fmt.Printf("Step %d already completed in last execution with result %s, we will use it\n", existingResult.StepIndex, result)

		return result, nil
	}

	var stepNumber int
	lastEntry := ec.journal.GetLastEntry()
	if lastEntry == nil || lastEntry.Status == journal.Completed {
		stepNumber = ec.journal.StartStep(fn)
	} else {
		stepNumber = lastEntry.StepIndex
	}
	fmt.Printf("Executing step %d\n", stepNumber)

	result, err := fn()
	if err != nil {
		ec.journal.FailLastStep(err)
		return nil, err
	}

	var serializedResult string
	if result != nil {
		if _, ok := result.(Void); ok {
			serializedResult = "{}"
		} else if resultBytes, err := json.Marshal(result); err == nil {
			serializedResult = string(resultBytes)
		} else {
			serializedResult = fmt.Sprintf("%v", result)
		}
	}

	ec.journal.CompleteLastStep(serializedResult)

	return result, nil
}

func (ec *ExecutionContext) PrintState() {
	entries := ec.journal.GetEntries()
	fmt.Printf("\nExecution State:\n")
	for _, entry := range entries {
		status := "COMPLETED"
		if entry.Status == journal.Failed {
			status = "FAILED"
		} else if entry.Status == journal.Running {
			status = "RUNNING"
		}
		fmt.Printf("Step %d: %s\n", entry.StepIndex, status)
		if entry.ErrorMsg != "" {
			fmt.Printf("  Error: %s\n", entry.ErrorMsg)
		}
	}
	fmt.Printf("Total steps: %d\n", len(entries))
}

func Run[T any](ctx Context, fn func() (T, error)) (T, error) {
	result, err := ctx.Run(func() (any, error) {
		return fn()
	})

	if err != nil {
		var zero T
		return zero, err
	}

	if typedResult, ok := result.(T); ok {
		return typedResult, nil
	}

	var zero T
	if _, ok := any(zero).(Void); ok {
		if _, ok := result.(Void); ok {
			return any(Void{}).(T), nil
		}
	}

	return zero, fmt.Errorf("type assertion failed: expected %T, got %T", zero, result)
}
