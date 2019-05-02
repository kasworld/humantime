// Copyright 2015,2016,2017,2018,2019 SeukWon Kang (kasworld@gmail.com)
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package humantimetaskqueue

import (
	"context"
	"time"

	"github.com/kasworld/timedtask/humantimetask"
)

func (tq *TaskQueue) Run(ctx context.Context) {
	tq.log.TraceService("Start Run %v", tq)
	defer func() { tq.log.TraceService("End Run %v", tq) }()

	chProcessTask := time.After(tq.repeatWait)
	tk1sec := time.NewTicker(1 * time.Second)
	defer tk1sec.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-tk1sec.C:
			tq.runStat.UpdateLap()

		case <-chProcessTask:
			if tq.IsPaused() {
				chProcessTask = time.After(tq.repeatWait)
			} else {
				nextWaitDur := tq.processTasks()
				chProcessTask = time.After(nextWaitDur)
			}
		}
	}
}

func (tq *TaskQueue) FlushTaskTill(till time.Time) {
	tq.runTasksEndWaitGroup.Wait()
	processed := 0
	tq.log.TraceService("Start FlushTaskTill %v", tq)
	defer func() { tq.log.TraceService("End FlushTaskTill %v, %v", processed, tq) }()

	for {
		peeked := tq.Peek()
		if peeked == nil { // no task to do
			return
		}
		if till.Before(peeked.TaskTime()) { // no current task
			return
		}

		t := tq.Pop()
		if t == nil {
			continue
		}
		tq.runStat.Inc()
		tso := tq.taskStat.GetStatByFuncName(t.GetTaskFnName())
		if err := t.RunWithStat(tso); err != nil {
			tq.log.Error("%v", err)
		}
		processed++
	}
}

func (tq *TaskQueue) runWaitTask(t *humantimetask.Task) {
	defer tq.runTasksEndWaitGroup.Done()
	tq.runStat.Inc()
	tso := tq.taskStat.GetStatByFuncName(t.GetTaskFnName())
	if err := t.RunWithStat(tso); err != nil {
		tq.log.Error("%v", err)
	}
}

func (tq *TaskQueue) processTasks() time.Duration {
	tq.log.Debug("%v processTasks", tq)
	startTime := time.Now().UTC()

	for {
		thisTime := time.Now().UTC()

		peeked := tq.Peek()
		if peeked == nil { // no task to do
			nextWaitDur := tq.repeatWait - thisTime.Sub(startTime)
			return makeInDuration(nextWaitDur, 0, tq.repeatWait)
		}
		if startTime.Before(peeked.TaskTime()) { // no current task
			nextWaitDur := peeked.TaskTime().Sub(thisTime)
			return makeInDuration(nextWaitDur, 0, tq.repeatWait)
		}

		t := tq.Pop()
		if t == nil {
			tq.log.Warn("%v task nil %v", tq, t)
			continue
		}
		callDuration := thisTime.Sub(t.TaskTime())
		if callDuration > tq.popDelay {
			tq.log.Debug("%v Delayed Pop %v %v", tq, t, callDuration)
		}

		tq.runTasksEndWaitGroup.Add(1)
		go tq.runWaitTask(t)
	}
}

func makeInDuration(x time.Duration, r1, r2 time.Duration) time.Duration {
	if x < r1 {
		x = r1
	}
	if x >= r2 {
		x = r2 - 1
	}
	return x
}
