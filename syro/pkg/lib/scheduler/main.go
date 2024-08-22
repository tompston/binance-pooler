// Package job holds some util functions for cron jobs, which check if a cron
// job is already running before starting it again. This is useful for jobs
// that take a long time to run, and should not be run again if they are
// already running. This is done to avoid possible concurrency issues.
package scheduler

import (
	"fmt"
	"sync"
	"syro/pkg/lib/errgroup"
	"time"

	"github.com/robfig/cron/v3"
)

// Scheduler is a wrapper around the cron scheduler that allows for the
// registration of jobs and the optional storage of job status and
// execution logs.
type Scheduler struct {
	cron    *cron.Cron
	Jobs    []*Job
	Storage Storage // Storage interface for monitoring the job execution (optional)
}

func NewScheduler(cron *cron.Cron, storage Storage) *Scheduler {
	return &Scheduler{cron: cron, Storage: storage}
}

// Register the cron job to the scheduler.
func (s *Scheduler) Register(j *Job) error {
	if s == nil {
		return fmt.Errorf("scheduler cannot be nil")
	}

	if j == nil {
		return fmt.Errorf("job cannot be nil")
	}

	// if the name of the job is already taken, return an error
	for _, job := range s.Jobs {
		if job != nil && job.Name == j.Name {
			return fmt.Errorf("job with name %v already exists", j.Name)
		}
	}

	return s.register(j)
}

// Start starts the cron scheduler.
//
// NOTE: Need to specify for how long the scheduler should run after
// calling this function (e.g. time.Sleep(1 * time.Hour) or forever)
func (s *Scheduler) Start() { s.cron.Start() }

// Job represents a cron job that can be registered with the cron scheduler.
type Job struct {
	Freq string       // Frequency of the job in cron format
	Name string       // Name of the job
	Func func() error // Function to be executed by the job
}

type Storage interface {
	// AllJobs returns a list of all registered jobs
	AllJobs() ([]JobInfo, error)
	// RegisterJob registers the details of the selected job
	RegisterJob(name, frequency string, status JobStatus, err error) error
	// RegisterExecution registers the execution of a job if the storage is specified
	RegisterExecution(*ExecutionLog) error
	// FindExecutions returns a list of job executions that match the filter
	FindExecutions(filter ExecutionFilter, limit int64, skip int64) ([]ExecutionLog, error)
}

// JobInfo stores information about the registered job
type JobInfo struct {
	CreatedAt       time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" bson:"updated_at"`
	Name            string    `json:"name" bson:"name"`
	Status          string    `json:"status" bson:"status"`
	Frequency       string    `json:"frequency" bson:"frequency"`
	Error           string    `json:"error" bson:"error"`
	ExitedWithError bool      `json:"exited_with_error" bson:"exited_with_error"`
}

// ExecutionLog stores information about the job execution
type ExecutionLog struct {
	Name          string        `json:"name" bson:"name"`
	InitializedAt time.Time     `json:"initialized_at" bson:"initialized_at"`
	FinishedAt    time.Time     `json:"finished_at" bson:"finished_at"`
	ExecutionTime time.Duration `json:"execution_time" bson:"execution_time"`
	Error         string        `json:"error" bson:"error"`
}

type ExecutionFilter struct {
	From         time.Time    `json:"from" bson:"from"`
	To           time.Time    `json:"to" bson:"to"`
	ExecutionLog ExecutionLog `json:"execution_log" bson:"execution_log"`
}

// newExecutionLog creates a new ExecutionLog instance.
func newExecutionLog(name string, initializedAt time.Time, err error) *ExecutionLog {
	log := &ExecutionLog{
		Name:          name,
		InitializedAt: initializedAt,
		FinishedAt:    time.Now().UTC(),
		ExecutionTime: time.Since(initializedAt),
	}

	// Avoid panics if the error is nil
	if err != nil {
		log.Error = err.Error()
	}

	return log
}

type JobStatus string

const (
	JobStatusInitialized JobStatus = "initialized"
	JobStatusRunning     JobStatus = "running"
	JobStatusDone        JobStatus = "done"
)

// register adds a new job to the cron scheduler and wraps the job function with a
// mutex lock to prevent the execution of the job if it is already running. If
// the function recieves a valid implementation of the Storage interface then
// this will also handle the registration and monitoring of the job.
func (s *Scheduler) register(j *Job) error {
	if j == nil {
		return fmt.Errorf("job cannot be nil")
	}

	if s.cron == nil {
		return fmt.Errorf("cron cannot be nil")
	}

	if j.Freq == "" {
		return fmt.Errorf("frequency has to be specified")
	}

	if j.Name == "" {
		return fmt.Errorf("name has to be specified")
	}

	if j.Func == nil {
		return fmt.Errorf("job function cannot be nil")
	}

	name := j.Name
	freq := j.Freq

	if s.Storage != nil {
		if err := s.Storage.RegisterJob(name, freq, JobStatusInitialized, nil); err != nil {
			return err
		}
	}

	// Accumulate errors in the c.AddJob function, because the cron.Job param does not return anything
	errors := errgroup.New()

	_, err := s.cron.AddJob(freq, newJobLock(func() {

		if s.Storage != nil {
			if err := s.Storage.RegisterJob(name, freq, JobStatusRunning, nil); err != nil {
				errors.Add(err)
			}
		}

		now := time.Now().UTC()

		// Passed in job function which should be executed by the cron job
		err := j.Func()

		if s.Storage != nil {
			if err := s.Storage.RegisterExecution(newExecutionLog(name, now, err)); err != nil {
				errors.Add(err)
			}

			if err := s.Storage.RegisterJob(name, freq, JobStatusDone, err); err != nil {
				errors.Add(err)
			}
		}

	}, name))

	if err != nil {
		return err
	}

	return errors.ToErr()
}

// jobLock is a mutex lock that prevents the execution of a
// job if it is already running.
type jobLock struct {
	jobFunc  func()
	jobName  string
	jobMutex sync.Mutex
}

func newJobLock(jobFunc func(), name string) *jobLock {
	return &jobLock{jobName: name, jobFunc: jobFunc}
}

func (j *jobLock) Run() {
	if j.tryLock() {
		defer j.jobMutex.Unlock()
		j.jobFunc()
	} else {
		fmt.Printf("job %v already running. Skipping...\n", j.jobName)
	}
}

func (j *jobLock) tryLock() bool { return j.jobMutex.TryLock() }
