// Copyright (C) 2022 Storj Labs, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package present

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const keepAliveInterval = 30 * time.Second

// keepAlive takes a ctx and a ping method. It returns a context rctx derived
// from ctx and a stop method. The derived rctx is canceled if ping returns
// a non-nil error. In the background after return, keepAlive will call ping
// with ctx every keepAliveInterval until:
//  1) ping returns an error, at which point rctx is canceled.
//  2) stop is called. in this case rctx is left alone.
//  3) ctx is canceled. rctx is also canceled as a consequence.
// stop is a no-op if the keepAlive loop has already been stopped. stop returns
// the first error that ping returned, if ping returned an error before stop
// was called.
func keepAlive(ctx context.Context, ping func(context.Context) error) (
	rctx context.Context, stop func() error) {

	ping = catchPanics(ping)

	rctx, cancel := context.WithCancel(ctx)
	done := make(chan bool)
	var once sync.Once
	var mu sync.Mutex
	var pingErr error
	var stopped bool

	ticker := time.NewTicker(keepAliveInterval)
	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				mu.Lock()
				if stopped || pingErr != nil {
					mu.Unlock()
					return
				}
				if err := ping(ctx); err != nil {
					pingErr = err
					mu.Unlock()
					cancel()
					return
				}
				mu.Unlock()
			}
		}
	}()

	return rctx, func() error {
		once.Do(func() { close(done) })
		// this mutex serves two purposes:
		// 1) it protects pingErr but
		// 2) it makes sure that there isn't an active ping call going
		//    by the time stop returns.
		mu.Lock()
		defer mu.Unlock()
		stopped = true
		return pingErr
	}
}

func catchPanics(cb func(context.Context) error) func(context.Context) error {
	return func(ctx context.Context) (err error) {
		defer func() {
			if rec := recover(); rec != nil {
				if rerr, ok := rec.(error); ok {
					err = rerr
				} else {
					err = fmt.Errorf("%v", rec)
				}
			}
		}()
		return cb(ctx)
	}
}
