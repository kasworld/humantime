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

package humantimetaskqueue2

import (
	"context"
	"time"

	"github.com/kasworld/timedtask/humantimetask"
)

func (tq *TaskQueue) Run(ctx context.Context) {
	tq.logger.TraceService("Start Run %v", tq)
	defer func() { tq.logger.TraceService("End Run %v", tq) }()

	tk1sec := time.NewTicker(1 * time.Second)
	defer tk1sec.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-tq.tasktimer.C:
			tq.processTasks()

			tq.mutex.RLock()
			if tq.paused {
				tq.tasktimer.Reset(timeDurationYear)
			} else {
				tq.scheduleTimerAtRootTick()
			}
			tq.mutex.RUnlock()

		case <-tk1sec.C:
			tq.runStat.UpdateLap()

		}
	}
}

func (tq *TaskQueue) processTasks() {
	tq.logger.Debug("%v processTasks", tq)
	startTime := time.Now().UTC()

	for {
		thisTime := time.Now().UTC()

		peeked := tq.Peek()
		if peeked == nil { // no task to do
			return
		}
		if startTime.Before(peeked.TaskTime()) { // no current task
			return
		}

		t := tq.Pop()
		if t == nil {
			tq.logger.Warn("%v task nil %v", tq, t)
			continue
		}
		delay := thisTime.Sub(t.TaskTime())
		if delay > tq.popDelay {
			tq.logger.Warn("%v Delayed Pop %v %v", tq, t, delay)
		}

		tq.runTasksEndWaitGroup.Add(1)
		go tq.runWaitTask(t)
	}
}

func (tq *TaskQueue) runWaitTask(t *humantimetask.Task) {
	defer tq.runTasksEndWaitGroup.Done()
	tq.runStat.Inc()
	tso := tq.taskStat.GetStatByFuncName(t.GetTaskFnName())
	t.RunWithStat(tso)
}

func (tq *TaskQueue) scheduleTimerAtRootTick() {
	if tq.paused {
		return
	}
	d := timeDurationYear
	if len(tq.pQueue) > 0 {
		t := tq.pQueue[0].TaskTime()
		d = t.Sub(time.Now())
	}
	tq.tasktimer.Reset(d)
}

func (tq *TaskQueue) FlushTaskTill(till time.Time) {
	tq.runTasksEndWaitGroup.Wait()
	processed := 0
	tq.logger.TraceService("Start FlushTaskTill %v", tq)
	defer func() { tq.logger.TraceService("End FlushTaskTill %v, %v", processed, tq) }()

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
		t.RunWithStat(tso)
		processed++
	}
}
