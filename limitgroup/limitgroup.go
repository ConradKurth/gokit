package limitgroup

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

// Group is our handler for the limiter
type Group struct {
	sema  chan struct{}
	group *errgroup.Group
}

// New will create a new limit group of size
func New(limit int) *Group {
	group, _ := errgroup.WithContext(context.Background())
	return &Group{
		sema:  make(chan struct{}, limit),
		group: group,
	}
}

// Go will start up a limited go routine
func (g *Group) Go(ctx context.Context, f func() error) {

	g.sema <- struct{}{}
	g.group.Go(func() (err error) {
		defer func() {
			// if we already have an error from the recover, don't do anything
			r := recover()
			if err == nil && r != nil {
				e, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				} else {
					err = e
				}
				err = fmt.Errorf("panic in limitgroup: %w", err)
			}
			<-g.sema
		}()
		err = f()
		return
	})
}

// Wait will wait for all the routines to finish returning an error
func (g *Group) Wait() error {
	return g.group.Wait()
}
