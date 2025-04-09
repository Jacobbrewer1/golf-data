package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jacobbrewer1/golf-data/cmd/retriever/runnables"
	logKeys "github.com/jacobbrewer1/golf-data/pkg/logging"
	"github.com/jacobbrewer1/golf-data/pkg/models"
	repo "github.com/jacobbrewer1/golf-data/pkg/repositories/retriever"
	"github.com/jacobbrewer1/web"
	"github.com/jacobbrewer1/web/logging"
	"github.com/jacobbrewer1/workerpool"
	"github.com/nats-io/nats.go/jetstream"
)

func (a *App) detailsTask(l *slog.Logger) web.AsyncTaskFunc {
	return func(ctx context.Context) {
		// Pick a random time between 60 - 180 minutes to run the task
		intervalNum, err := rand.Int(rand.Reader, big.NewInt(ticketMagicNumber()))
		if err != nil {
			l.Error("failed to generate random interval", slog.String(logging.KeyError, err.Error()))
			return
		}
		interval := time.Duration(intervalNum.Int64()) * time.Minute
		l.Info("generated random interval", slog.String(logKeys.KeyInterval, interval.String()))
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		if a.base.IsLeader() {
			l.Info("Am leader, running on startup")
			if err := detailsWorker(ctx, l, a.r, a.base.NatsJetStream(), a.base.WorkerPool()); err != nil {
				l.Error("failed to run details worker", slog.String(logging.KeyError, err.Error()))
			}
		}

		l.Info("Starting details worker loop")
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
				if err := detailsWorker(
					ctx,
					logging.LoggerWithComponent(l, "details-worker"),
					a.r,
					a.base.NatsJetStream(),
					a.base.WorkerPool(),
				); err != nil {
					l.Error("failed to run details worker", slog.String(logging.KeyError, err.Error()))
					continue
				}
			}

			if a.base.IsLeader() {
				continue
			}
			l.Info("not leader, waiting for leader change")
		}
	}
}

func detailsWorker(
	ctx context.Context,
	l *slog.Logger,
	r repo.Repository,
	publisher jetstream.JetStream,
	wp workerpool.Pool,
) error {
	courses, err := r.GetCourses()
	if err != nil && !errors.Is(err, repo.ErrNoCourses) {
		return fmt.Errorf("failed to get courses: %w", err)
	} else if len(courses) == 0 {
		l.Debug("no courses found")
		return nil
	}

	for _, course := range courses {
		courseId := course.Id
		l.Debug("courseId", slog.Int(logKeys.KeyCourseId, courseId))

		details, err := getEnglandGolfDetails(ctx, l, courseId)
		if err != nil {
			l.Error("failed to get details", slog.String(logging.KeyError, err.Error()))
			continue
		}

		if len(details) == 0 {
			l.Debug("no details found", slog.Int(logKeys.KeyCourseId, courseId))
			continue
		}

		for _, egDetail := range details {
			detail := egDetail.ToModel(courseId)

			detailsRunnable := runnables.NewPublishDetailsToNats(
				ctx,
				logging.LoggerWithComponent(l, "publish-details-to-nats"),
				publisher,
				detail,
			)

			if err := wp.BlockingSchedule(detailsRunnable); err != nil {
				l.Error("failed to schedule details runnable", slog.String(logging.KeyError, err.Error()))
				continue
			}

			holes := make([]*models.Hole, 0)
			for _, hole := range egDetail.Holes {
				holes = append(holes, hole.ToModel(detail.Id))
			}

			holesRunnable := runnables.NewPublishHolesToNats(
				ctx,
				logging.LoggerWithComponent(l, "publish-holes-to-nats"),
				publisher,
				holes,
			)

			if err := wp.BlockingSchedule(holesRunnable); err != nil { // nolint:revive // Traditional error handling
				l.Error("failed to schedule holes runnable", slog.String(logging.KeyError, err.Error()))
				continue
			}
		}
	}

	return nil
}

func getEnglandGolfDetails(ctx context.Context, l *slog.Logger, courseId int) ([]*EnglandGolfCourseDetailsResponse, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.Logger = logging.LoggerWithComponent(l, "retryablehttp-courses")
	retryClient.RetryMax = 3
	client := retryClient.StandardClient()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/clubs/getCourses", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	q := req.URL.Query()
	q.Add("courseId", strconv.Itoa(courseId))
	q.Add("gender", "M")
	q.Add("isNineHoles", strconv.FormatBool(false))
	q.Add("memberUid", "")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get details: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			l.Error("failed to close response body", slog.String(logging.KeyError, err.Error()))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get details: %s", resp.Status)
	}

	details := make([]*EnglandGolfCourseDetailsResponse, 0)
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return details, nil
}
