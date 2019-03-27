package agent

import (
	"sort"
	"time"

	"github.com/golang/glog"

	"github.com/joyrex2001/nightshift/internal/scanner"
	"github.com/joyrex2001/nightshift/internal/schedule"
)

type event struct {
	at    time.Time
	obj   scanner.Object
	sched *schedule.Schedule
}

// Scale will process all scanned objects and scale them accordingly.
func (a *worker) Scale() {
	glog.Info("Scaling resources start...")
	a.now = time.Now()
	for _, obj := range a.objects {
		for _, e := range a.getEvents(obj) {
			glog.V(4).Infof("Scale event: %v", e)
			if e.obj.Scale != nil {
				repl, err := e.sched.GetReplicas()
				if err == nil {
					err = e.obj.Scale(repl)
				}
				if err != nil {
					glog.Errorf("Error scaling deployment: %s", err)
				}
			}
		}
	}
	a.past = a.now
	glog.V(4).Info("Scaling resources finished...")
}

// getEvents will return the events in chronological order that have to be
// done for the given object in the current tick.
func (a *worker) getEvents(obj scanner.Object) []event {
	var err error
	ev := []event{}
	for _, s := range obj.Schedule {
		for next := a.past; !next.After(a.now); next = next.AddDate(0, 0, 1) {
			next, err = s.GetNextTrigger(next)
			if err != nil {
				glog.Errorf("Error processing trigger: %s", err)
				continue
			}
			if a.now.After(next) || a.now == next {
				ev = append(ev, event{next, obj, s})
			}
		}
	}
	// order events by time
	sort.Slice(ev, func(i, j int) bool { return ev[i].at.Before(ev[j].at) })
	return ev
}
