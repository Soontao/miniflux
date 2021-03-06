// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package worker // import "miniflux.app/worker"

import (
	"fmt"
	"time"

	"github.com/xiaonanln/keylock"
	"miniflux.app/config"
	"miniflux.app/logger"
	"miniflux.app/metric"
	"miniflux.app/model"
	feedHandler "miniflux.app/reader/handler"
	"miniflux.app/storage"
)

// Worker refreshes a feed in the background.
type Worker struct {
	id    int
	store *storage.Storage
}

var jobLock = keylock.NewKeyLock()

// Run wait for a job and refresh the given feed.
func (w *Worker) Run(c chan model.Job) {
	logger.Debug("[Worker] #%d started", w.id)

	for {
		job := <-c
		logger.Debug("[Worker #%d] Received feed #%d for user #%d", w.id, job.FeedID, job.UserID)
		sID := fmt.Sprintf("%v", job.FeedID)

		jobLock.Lock(sID)
		startTime := time.Now()
		refreshErr := feedHandler.RefreshFeed(w.store, job.UserID, job.FeedID)
		jobLock.Unlock(sID)

		if config.Opts.HasMetricsCollector() {
			status := "success"
			if refreshErr != nil {
				status = "error"
			}
			metric.BackgroundFeedRefreshDuration.WithLabelValues(status).Observe(time.Since(startTime).Seconds())
		}

		if refreshErr != nil {
			logger.Error("[Worker] Refreshing the feed #%d returned this error: %v", job.FeedID, refreshErr)
		}
	}
}
