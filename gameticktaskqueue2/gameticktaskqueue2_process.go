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

package gameticktaskqueue2

import (
	"container/heap"
	"context"
	"fmt"
	"time"

	"github.com/kasworld/gametick"
	"github.com/kasworld/globalgametick"
	"github.com/kasworld/timedtask/gameticktask"
)

func (tq *TaskQueue) Run(ctx context.Context) {
	tq.log.TraceService("Start Run %v", tq)
	defer func() { tq.log.TraceService("End Run %v", tq) }()

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
	startTick := globalgametick.GetGameTick()

	for {
		thisTick := globalgametick.GetGameTick()

		peeked := tq.Peek()
		if peeked == nil {
			return
		}
		if startTick < peeked.TaskGameTick() {
			return
		}

		t := tq.Pop()
		if t == nil {
			continue
		}
		callDuration := thisTick - t.TaskGameTick()
		if callDuration > tq.popDelay {
			tq.log.Debug("%v Delayed Pop %v %v", tq, t, callDuration)
		}

		tq.runTasksEndWaitGroup.Add(1)
		go tq.runWaitTask(t)
	}
}

func (tq *TaskQueue) scheduleTimerAtRootTick() {
	if tq.paused {
		return
	}
	d := timeDurationYear
	if len(tq.pQueue) > 0 {
		t := tq.pQueue[0].TaskGameTick()
		d = (t - globalgametick.GetGameTick()).ToTimeDuration()
	}
	tq.tasktimer.Reset(d)
}

func (tq *TaskQueue) Pause() error {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	if tq.paused {
		return nil
	}
	tq.paused = true
	tq.tasktimer.Reset(timeDurationYear)
	tq.log.TraceService("%v paused", tq)
	return nil
}

func (tq *TaskQueue) Resume() error {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	if !tq.paused {
		return nil
	}

	tq.paused = false
	tq.scheduleTimerAtRootTick()
	tq.log.TraceService("%v resumed", tq)
	return nil
}

func (tq *TaskQueue) UpdateTaskArgAndTick(t *gameticktask.Task, uparg interface{}, uptick gametick.GameTick) error {
	if t == nil {
		tq.log.Fatal("failed to update nil task")
	}

	tq.mutex.Lock()
	defer tq.mutex.Unlock()
	return tq.update(t, uparg, uptick)
}

func (tq *TaskQueue) UpdateTaskTick(t *gameticktask.Task, uptick gametick.GameTick) error {
	if t == nil {
		tq.log.Fatal("failed to update nil task")
	}

	tq.mutex.Lock()
	defer tq.mutex.Unlock()
	return tq.update(t, t.Argument(), uptick)
}

func (tq *TaskQueue) update(t *gameticktask.Task, uparg interface{}, uptick gametick.GameTick) error {

	if len(tq.pQueue) > 0 {
		oldRootTick := tq.pQueue[0].TaskGameTick()
		if err := tq.pQueue.Update(t, uparg, uptick, t.GetTaskFn()); err != nil {
			return err
		}
		newRootTick := tq.pQueue[0].TaskGameTick()
		if oldRootTick != newRootTick {
			tq.scheduleTimerAtRootTick()
		}
	} else {
		return fmt.Errorf("%v update failed, no items enqueued", tq)
	}
	return nil
}

func (tq *TaskQueue) Remove(t *gameticktask.Task) error {
	if t == nil {
		tq.log.Fatal("failed to remove nil task")
	}

	tq.mutex.Lock()
	defer tq.mutex.Unlock()
	if len(tq.pQueue) > 0 {
		oldroot := tq.pQueue[0]
		if err := tq.pQueue.Remove(t); err != nil {
			return err
		}
		if oldroot == t {
			tq.scheduleTimerAtRootTick()
		}
	} else {
		return fmt.Errorf("%v remove failed, no items enqueued", tq)
	}
	return nil
}

func (tq *TaskQueue) Push(t *gameticktask.Task) {
	if t == nil {
		tq.log.Fatal("%v tried to push nil task", tq)
	}

	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	if t.IsValid() {
		tq.log.Fatal("%v tried to push %v already pushed", tq, t)
	}
	heap.Push(&tq.pQueue, t)
	if tq.pQueue[0] == t {
		tq.scheduleTimerAtRootTick()
	}
}
