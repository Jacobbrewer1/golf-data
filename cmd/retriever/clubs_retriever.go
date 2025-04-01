package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"runtime"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jacobbrewer1/golf-data/cmd/retriever/runnables"
	logKeys "github.com/jacobbrewer1/golf-data/pkg/logging"
	"github.com/jacobbrewer1/web"
	"github.com/jacobbrewer1/web/logging"
	"github.com/jacobbrewer1/workerpool"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	clubSearchURL = "https://www.englandgolf.org/api/clubs/ClubSearch"
)

func (a *App) clubsTask() web.AsyncTaskFunc {
	return func(ctx context.Context) {
		// Pick a random time between 15 - 60 minutes to run the task
		intervalNum, err := rand.Int(rand.Reader, big.NewInt(45))
		if err != nil {
			a.base.Logger().Error("failed to generate random interval", slog.String(logging.KeyError, err.Error()))
			return
		}
		interval := time.Duration(intervalNum.Int64()+15) * time.Minute
		a.base.Logger().Debug("generated random interval", slog.String(logKeys.KeyInterval, interval.String()))
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-a.base.LeaderChange():
				if !a.base.IsLeader() {
					a.base.Logger().Info("not leader, waiting for leader change")
					continue
				}

				for {
					select {
					case <-ctx.Done():
						return
					case <-a.base.LeaderChange():
						// Do nothing as this logic is handled in the outer loop
					case <-ticker.C:
						a.base.Logger().Debug("ticker ticked")
						if err := clubWorker(ctx, a.base.Logger(), a.base.NatsJetStream(), a.base.WorkerPool()); err != nil {
							a.base.Logger().Error("failed to run club worker", slog.String(logging.KeyError, err.Error()))
							continue
						}
					}

					if a.base.IsLeader() {
						continue
					}
					a.base.Logger().Info("not leader, waiting for leader change")
					break
				}
			}
		}
	}
}

func clubWorker(
	ctx context.Context,
	l *slog.Logger,
	publisher jetstream.JetStream,
	wp workerpool.Pool,
) error {
	pageNum := 1
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			clubs, err := getEnglandGolfClubs(ctx, l, pageNum)
			if err != nil {
				return fmt.Errorf("failed to get england golf clubs: %w", err)
			}

			if len(clubs) == 0 {
				l.Info("no more clubs to fetch, fetching complete")
				return nil
			}

			for _, club := range clubs {
				dbClub := club.ToModel()
				runnable := runnables.NewPublishClubToNats(ctx, l, publisher, dbClub)
				if err := wp.BlockingSchedule(runnable); err != nil { // nolint:revive // Traditional error handling
					l.Error("failed to schedule club runnable", slog.String(logging.KeyError, err.Error()))
					return fmt.Errorf("failed to schedule club runnable: %w", err)
				}
			}

			pageNum++
		}
	}
}

func getEnglandGolfClubs(ctx context.Context, l *slog.Logger, pageNum int) ([]*EnglandGolfClubResponse, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.Logger = logging.LoggerWithComponent(l, "retryablehttp")
	retryClient.RetryMax = 3
	client := retryClient.StandardClient()

	type body struct {
		UserLatitude  any   `json:"userLatitude"`
		UserLongitude any   `json:"userLongitude"`
		AmenityIds    []any `json:"amenityIds"`
		ProgrammeIds  []any `json:"programmeIds"`
		PageNumber    int   `json:"pageNumber"`
		PageSize      int   `json:"pageSize"`
	}

	bdy := &body{
		UserLatitude:  nil,
		UserLongitude: nil,
		AmenityIds:    make([]any, 0),
		ProgrammeIds:  make([]any, 0),
		PageNumber:    pageNum,
		PageSize:      runtime.NumCPU(), // Allows for the number of clubs to be fetched to be equal to the number of CPUs
	}

	dataBuf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(dataBuf).Encode(bdy); err != nil {
		return nil, fmt.Errorf("failed to encode request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, clubSearchURL, dataBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error doing request: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			l.Error("failed to close response body", slog.String(logging.KeyError, err.Error()))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	clubs := make([]*EnglandGolfClubResponse, 0)
	if err := json.NewDecoder(resp.Body).Decode(&clubs); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return clubs, nil
}
