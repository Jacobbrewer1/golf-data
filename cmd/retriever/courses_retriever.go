package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jacobbrewer1/golf-data/cmd/retriever/runnables"
	logKeys "github.com/jacobbrewer1/golf-data/pkg/logging"
	repo "github.com/jacobbrewer1/golf-data/pkg/repositories/retriever"
	"github.com/jacobbrewer1/web"
	"github.com/jacobbrewer1/web/logging"
	"github.com/jacobbrewer1/workerpool"
	"github.com/nats-io/nats.go/jetstream"
)

func (a *App) coursesTask(l *slog.Logger) web.AsyncTaskFunc {
	return func(ctx context.Context) {
		// Pick a random time between 15 - 60 minutes to run the task
		intervalNum, err := rand.Int(rand.Reader, big.NewInt(45))
		if err != nil {
			l.Error("failed to generate random interval", slog.String(logging.KeyError, err.Error()))
			return
		}
		interval := time.Duration(intervalNum.Int64()+15) * time.Second // TODO: Change to minutes
		l.Debug("generated random interval", slog.String(logKeys.KeyInterval, interval.String()))
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		//l.Info("leader change detected, running on startup")
		//if err := courseWorker(ctx, l, a.r, a.base.NatsJetStream(), a.base.WorkerPool()); err != nil {
		//	l.Error("failed to run club worker", slog.String(logging.KeyError, err.Error()))
		//}

		l.Info("Starting course worker loop")
		for {
			select {
			case <-ctx.Done():
				return
			case <-a.base.LeaderChange():
				// Do nothing as this logic is handled in the outer loop
			case <-ticker.C:
				if !a.base.IsLeader() {
					break
				}

				l.Debug("ticker ticked")
				if err := courseWorker(
					ctx,
					logging.LoggerWithComponent(l, "course-worker"),
					a.r,
					a.base.NatsJetStream(),
					a.base.WorkerPool(),
				); err != nil {
					l.Error("failed to run club worker", slog.String(logging.KeyError, err.Error()))
					continue
				}
			}

			if a.base.IsLeader() {
				continue
			}
			l.Info("not leader, waiting for leader change")
			break
		}
	}
}

func courseWorker(
	ctx context.Context,
	l *slog.Logger,
	r repo.Repository,
	publisher jetstream.JetStream,
	wp workerpool.Pool,
) error {
	clubs, err := r.GetClubs()
	if err != nil {
		return fmt.Errorf("failed to get clubs: %w", err)
	}

	for _, club := range clubs {
		l.Debug("processing club", slog.String(logKeys.KeyClubId, fmt.Sprintf("%d", club.Id)))
		courses, err := getEnglandGolfCourses(ctx, l, club.Id)
		if err != nil {
			return fmt.Errorf("failed to get courses: %w", err)
		}

		for _, course := range courses {
			runnable := runnables.NewPublishCourseToNats(
				ctx,
				logging.LoggerWithComponent(l, "publish-course-to-nats"),
				publisher,
				course.ToModel(club.Id),
			)

			if err := wp.Schedule(runnable); err != nil {
				l.Error("failed to schedule course runnable", slog.String(logKeys.KeyError, err.Error()))
				continue
			}

			l.Debug("course runnable scheduled", slog.String(logKeys.KeyClubId, fmt.Sprintf("%d", club.Id)))
		}
	}

	l.Info("course worker completed")
	return nil
}

func getEnglandGolfCourses(ctx context.Context, l *slog.Logger, clubId int) ([]*EnglandGolfCourseResponse, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.Logger = logging.LoggerWithComponent(l, "retryablehttp-courses")
	retryClient.RetryMax = 3
	client := retryClient.StandardClient()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.englandgolf.org/api/clubs/getCourses", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("clubId", fmt.Sprintf("%d", clubId))
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get courses: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			l.Error("failed to close response body", slog.String(logging.KeyError, err.Error()))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get courses: %s", resp.Status)
	}

	courses := make([]*EnglandGolfCourseResponse, 0)
	if err := json.NewDecoder(resp.Body).Decode(&courses); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return courses, nil
}
