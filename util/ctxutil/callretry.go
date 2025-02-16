package ctxutil

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Allows fn to return early on ctx cancel. If fn does not return in time, lateFn will run at the end of fn (async).
func Call(ctx context.Context, prefix string, fn func() error, lateFn func(error)) error {
	buildErr := func(e error) error {
		return fmt.Errorf("%v: %v", prefix, e)
	}

	var d = struct {
		sync.Mutex
		fn struct {
			done bool
			err  error
		}
		ctxDone bool
	}{}

	// run fn in go routine
	ctx2, cancel := context.WithCancel(ctx)
	id := addCall(prefix) // keep track of fn()
	go func() {
		defer doneCall(id) // keep track of fn()
		defer cancel()

		err := fn()

		d.Lock()
		defer d.Unlock()
		d.fn.done = true
		if err != nil {
			err = buildErr(err)
		}
		d.fn.err = err
		if d.ctxDone {
			if lateFn != nil {
				lateFn(err)
			} else {
				// err is lost
			}
		}
	}()

	<-ctx2.Done()

	d.Lock()
	defer d.Unlock()
	d.ctxDone = true
	if d.fn.done {
		return d.fn.err
	} else {
		// context was canceled and fn has not returned yet
		return buildErr(ctx2.Err())
	}
}

func Retry(ctx context.Context, retryPause time.Duration, prefix string, fn func() error, lateFn func(error)) error {
	var err error
	for {
		err = Call(ctx, prefix, fn, lateFn)
		if err != nil {
			//  keep retrying
		} else {
			return nil // done
		}
		select {
		case <-ctx.Done():
			return err // err is non-nil
		default: // non-blocking select
			time.Sleep(retryPause) // sleep before next retry
		}
	}
}

type cdata struct {
	t time.Time
	s string
}

var cmu sync.Mutex
var callm = map[int]*cdata{}
var ci = 0

func addCall(s string) int {
	cmu.Lock()
	defer cmu.Unlock()
	ci++
	callm[ci] = &cdata{s: s, t: time.Now()}
	return ci
}

func doneCall(v int) {
	cmu.Lock()
	defer cmu.Unlock()
	delete(callm, v)
}

func CallsState() string {
	cmu.Lock()
	defer cmu.Unlock()
	u := []string{}
	now := time.Now()
	for _, d := range callm {
		s := fmt.Sprintf("%v: %v ago", d.s, now.Sub(d.t))
		u = append(u, s)
	}
	return fmt.Sprintf("%v entries\n%v\n", len(u), strings.Join(u, "\n"))
}
