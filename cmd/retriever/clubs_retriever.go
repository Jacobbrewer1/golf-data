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
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/jacobbrewer1/golf-data/cmd/retriever/runnables"
	logKeys "github.com/jacobbrewer1/golf-data/pkg/logging"
	"github.com/jacobbrewer1/goredis"
	"github.com/jacobbrewer1/web"
	"github.com/jacobbrewer1/web/logging"
	"github.com/jacobbrewer1/workerpool"
)

const (
	clubSearchURL = "https://www.englandgolf.org/api/clubs/ClubSearch"
)

func clubsTask(l *slog.Logger, keydb func() goredis.Pool, wp func() workerpool.Pool) web.AsyncTaskFunc {
	return func(ctx context.Context) error {
		// Pick a random time between 15 - 60 minutes to run the task
		intervalNum, err := rand.Int(rand.Reader, big.NewInt(45))
		if err != nil {
			return fmt.Errorf("failed to generate random interval: %w", err)
		}
		interval := time.Duration(intervalNum.Int64()+15) * time.Minute
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		l.Info("ticker started", slog.String(logKeys.KeyInterval, interval.String()))

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				l.Debug("ticker ticked")
				if err := clubWorker(ctx, l, keydb(), wp()); err != nil {
					return fmt.Errorf("failed to run club worker: %w", err)
				}
			}
		}
	}
}

func clubWorker(
	ctx context.Context,
	l *slog.Logger,
	keydb goredis.Pool,
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
				runnable := runnables.NewClubToKeyDB(ctx, l, keydb, dbClub)
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
		PageSize:      150,
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
