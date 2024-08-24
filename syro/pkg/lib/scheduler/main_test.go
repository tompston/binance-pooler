package scheduler

// go test -timeout 30s syro/pkg/lib/scheduler -v -count=1

// go test -timeout 30s syro/pkg/lib/scheduler -count=1

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"syro/pkg/app/db"
	"syro/pkg/lib/mongodb"
	"syro/pkg/lib/utils"
	"syro/pkg/lib/validate"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
)

func TestJobRun(t *testing.T) {
	counter := int32(0)
	j := newJobLock(func() {
		time.Sleep(100 * time.Millisecond)
		atomic.AddInt32(&counter, 1)
	}, "testJob")

	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			j.Run()
		}()
	}
	wg.Wait()

	if atomic.LoadInt32(&counter) != 1 {
		t.Fatalf("Expected job function to run once, but it ran %d times", counter)
	}
}

func TestJobRunLocking(t *testing.T) {
	counter := int32(0)
	j := newJobLock(func() {
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt32(&counter, 1)
	}, "testJob")

	// first run
	go j.Run()

	// sleep for a moment to allow the first job to start running
	time.Sleep(10 * time.Millisecond)

	// second run
	go j.Run()

	// sleep for a moment to allow the second job to attempt to start
	time.Sleep(10 * time.Millisecond)

	// At this point, the second job should have attempted to start and failed,
	// so the counter should still be 0
	if atomic.LoadInt32(&counter) != 0 {
		t.Fatalf("Expected job function not to run, but it ran %d times", counter)
	}

	// Wait for the first job to complete
	time.Sleep(50 * time.Millisecond)

	// Now the counter should be 1
	if atomic.LoadInt32(&counter) != 1 {
		t.Fatalf("Expected job function to run once, but it ran %d times", counter)
	}
}

func TestCronRegistration(t *testing.T) {

	t.Run("return err if job is nil", func(t *testing.T) {
		sched := NewScheduler(cron.New(), "")
		err := sched.addJob(nil)

		if err == nil {
			t.Fatalf("Expected error, got nil")
		}

		if err.Error() != "job cannot be nil" {
			t.Fatalf("Expected error, got %v", err)
		}
	})

	t.Run("return err if cron is nil", func(t *testing.T) {
		sched := NewScheduler(nil, "")

		err := sched.addJob(&Job{
			Freq: "@every 1m",
			Name: "test",
		})

		if err == nil {
			t.Fatalf("Expected error, got nil")
		}

		if err.Error() != "cron cannot be nil" {
			t.Fatalf("Expected error, got %v", err)
		}
	})

	t.Run("return err if freq is nil", func(t *testing.T) {
		sched := NewScheduler(cron.New(), "")

		err := sched.addJob(&Job{
			Freq: "",
		})

		if err.Error() != "frequency has to be specified" {
			t.Fatalf("Expected error, got %v", err)
		}
	})
}

func TestMongoStorage(t *testing.T) {
	conn, err := mongodb.New("localhost", 27017, "", "")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Disconnect(context.Background())

	cronListColl := mongodb.Coll(conn, db.TEST_DB, "cron_list")
	cronHistoryColl := mongodb.Coll(conn, db.TEST_DB, "cron_history")

	// Remove the previous data
	if err := cronListColl.Drop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := cronHistoryColl.Drop(context.Background()); err != nil {
		t.Fatal(err)
	}

	storage, err := NewMongoStorage(cronListColl, cronHistoryColl)
	if err != nil {
		t.Fatal(err)
	}

	const SOURCE = "test-source"

	t.Run("test job registration", func(t *testing.T) {
		if err := storage.RegisterJob(SOURCE, "cron-job-1", "@every 1m", JobStatusInitialized, nil); err != nil {
			t.Fatal(err)
		}

		jobs, err := storage.AllJobs()
		if err != nil {
			t.Fatal(err)
		}

		if len(jobs) != 1 {
			t.Fatalf("Expected 1 job, got %d", len(jobs))
		}

		registeredJob := jobs[0]

		fmt.Printf("registeredJob: %v\n", registeredJob)

		decoded, err := utils.DecodeStructToStrings(registeredJob)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Printf("decoded.BSON: %v\n", decoded.BSON)

		expectedSubstrings := []string{
			`"created_at":{"$date":"`,
			`"updated_at":{"$date":`,
			`"name":"cron-job-1"`,
			`"status":"initialized"`,
			`"frequency":"@every 1m"`,
			`"error":""`,
			`"exited_with_error":false`,
			`"source":"test-source"`,
		}

		if err := validate.StringIncludes(decoded.BSON, expectedSubstrings); err != nil {
			t.Fatalf("bson did not have the expected fields: %v", err)
		}

		if registeredJob.Name != "cron-job-1" {
			t.Fatalf("Expected job name to be 'cron-job-1', got %s", registeredJob.Name)
		}

		if registeredJob.Frequency != "@every 1m" {
			t.Fatalf("Expected frequency to be '@every 1m', got %s", registeredJob.Frequency)
		}

		// c := cron.New()

		sched := NewScheduler(cron.New(), "test-pooler").WithStorage(storage)

		const CRON_NAME = "cron-job-1"

		// Register the job
		if err := sched.addJob(&Job{
			Freq: "@every 1s",
			Name: CRON_NAME,
			Func: func() error {
				fmt.Println("Running job...")
				return nil
			},
		}); err != nil {
			t.Fatal(err)
		}

		// c.Start()

		// start the cron
		sched.Start()

		// sleep for a moment to allow the cron to run and register executions
		time.Sleep(2 * time.Second)

		t.Run("test execution finder - find expected documents with correct field names", func(t *testing.T) {
			execHistory, err := storage.FindExecutions(ExecutionFilter{
				Limit: 10,
			})
			if err != nil {
				t.Fatal(err)
			}

			if len(execHistory) == 0 {
				t.Fatalf("Expected at least one execution, got 0")
			}

			for _, exec := range execHistory {
				if exec.Name != CRON_NAME {
					t.Fatalf("Expected execution name to be '%v', got %s", CRON_NAME, exec.Name)
				}

				if exec.ExecutionTime > 2*time.Second {
					t.Fatalf("Expected execution time to be less than 2s, got %v", exec.ExecutionTime)
				}

				decodedExec, err := utils.DecodeStructToStrings(exec)
				if err != nil {
					t.Fatal(err)
				}

				// fmt.Printf("decoded.JSON: %v\n", decodedExec.JSON)

				if err := validate.StringIncludes(decodedExec.BSON, []string{
					`"name":"` + CRON_NAME,
					`"error":""`,
					`"finished_at":{"$date":"`,
					`"initialized_at":{"$date":"`,
				}); err != nil {
					t.Fatalf("bson did not have the expected fields: %v", err)
				}
			}
		})

		t.Run("test execution finder - limit works", func(t *testing.T) {
			execHistory, err := storage.FindExecutions(ExecutionFilter{
				Limit: 1,
			})
			if err != nil {
				t.Fatal(err)
			}

			if len(execHistory) != 1 {
				t.Fatalf("Expected one execution, got %v", len(execHistory))
			}
		})

		t.Run("test execution finder - name filter works for non existing crons", func(t *testing.T) {
			execHistory, err := storage.FindExecutions(ExecutionFilter{
				Limit: 1,
				ExecutionLog: ExecutionLog{
					Name: "does-not-exist-qweqwe",
				},
			})
			if err != nil {
				t.Fatal(err)
			}

			if len(execHistory) != 0 {
				t.Fatalf("Expected 0 executions because cron name does not exist, got %v", len(execHistory))
			}
		})

		t.Run("test execution finder - name filter works existing crons", func(t *testing.T) {
			execHistory, err := storage.FindExecutions(ExecutionFilter{
				Limit: 1,
				ExecutionLog: ExecutionLog{
					Name: "does-not-exist-qweqwe",
				},
			})
			if err != nil {
				t.Fatal(err)
			}

			if len(execHistory) != 0 {
				t.Fatalf("Expected 0 executions because cron name does not exist, got %v", len(execHistory))
			}
		})
	})
}
