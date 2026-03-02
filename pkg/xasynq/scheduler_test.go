package xasynq

import (
	"context"
	"log"
	"testing"
	"time"

	"go.uber.org/zap"
)

const TypeScheduledGet = "scheduled:get"

var zapLogger, _ = zap.NewProduction()

type ScheduledGetPayload struct {
	URL string `json:"url"`
}

func handleScheduledGetTask(ctx context.Context, p *ScheduledGetPayload) error {
	log.Printf("[ScheduledGet] Task for URL %s completed successfully", p.URL)
	return nil
}

func registerSchedulerTasks(scheduler *Scheduler) ([]string, error) {
	var entTryIDs []string

	payload1 := &ScheduledGetPayload{URL: "https://google.com"}
	entryID1, err := scheduler.RegisterTask("@every 2s", TypeScheduledGet, payload1)
	if err != nil {
		return nil, err
	}
	log.Printf("Registered periodic task with entry ID: %s", entryID1)
	entTryIDs = append(entTryIDs, entryID1)

	payload2 := &ScheduledGetPayload{URL: "https://bing.com"}
	entryID2, err := scheduler.RegisterTask("@every 3s", TypeScheduledGet, payload2)
	if err != nil {
		return nil, err
	}
	log.Printf("Registered periodic task with entry ID: %s", entryID2)
	entTryIDs = append(entTryIDs, entryID2)

	return entTryIDs, nil
}

func runServer(redisCfg RedisConfig) (*Server, error) {
	serverCfg := DefaultServerConfig(WithLogger(zapLogger)) // Uses critical, default, low queues
	srv := NewServer(redisCfg, serverCfg)
	srv.Use(LoggingMiddleware(WithLogger(zapLogger)))

	// register task handle function, there are three registration methods available
	RegisterTaskHandler(srv.Mux(), TypeScheduledGet, HandleFunc(handleScheduledGetTask))

	srv.Run()

	return srv, nil
}

func TestPeriodicTask(t *testing.T) {
	scheduler := NewScheduler(getRedisConfig(),
		WithSchedulerOptions(nil),
		WithSchedulerLogger(WithLogger(zapLogger)),
		WithSchedulerLogLevel(2),
	)
	entTryIDs, err := registerSchedulerTasks(scheduler)
	if err != nil {
		t.Log("register scheduler tasks failed", err)
	} else {
		scheduler.Run()
	}

	srv, err := runServer(getRedisConfig())
	if err != nil {
		t.Log("run server failed", err)
		return
	}

	time.Sleep(7 * time.Second)

	for _, entryID := range entTryIDs {
		err = scheduler.Unregister(entryID)
		if err != nil {
			t.Log("unregister scheduler task failed", err)
		} else {
			t.Log("unregister scheduler task success", entryID)
		}
	}

	srv.Shutdown()
}
