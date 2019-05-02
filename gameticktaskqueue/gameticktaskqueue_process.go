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

package gameticktaskqueue

import (
	"container/heap"
	"context"
	"time"

	"github.com/kasworld/gametick"
	"github.com/kasworld/globalgametick"
	"github.com/kasworld/timedtask/gameticktask"
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
				nextWait := tq.processTasks()
				chProcessTask = time.After(nextWait)
			}
		}
	}
}

func (tq *TaskQueue) processTasks() time.Duration {
	startTick := globalgametick.GetGameTick()
	repeatWaitTick := gametick.FromTimeDurationToTickType(tq.repeatWait)

	for {
		thisTick := globalgametick.GetGameTick()

		peeked := tq.Peek()
		if peeked == nil { // no task to do
			nextWait := repeatWaitTick - (thisTick - startTick)
			return gametick.MakeIn(nextWait, 0, repeatWaitTick).ToTimeDuration()
		}
		if startTick < peeked.TaskGameTick() { // no current task
			nextWait := peeked.TaskGameTick() - thisTick
			return gametick.MakeIn(nextWait, 0, repeatWaitTick).ToTimeDuration()
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

func (tq *TaskQueue) Pause() error {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	tq.paused = true
	tq.log.TraceService("%v paused", tq)
	return nil
}

func (tq *TaskQueue) Resume() error {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()

	tq.paused = false
	tq.log.TraceService("%v resumed", tq)
	return nil
}

func (tq *TaskQueue) UpdateTaskArgAndTick(t *gameticktask.Task, uparg interface{}, uptick gametick.GameTick) error {
	if t == nil {
		tq.log.Fatal("failed to update nil task")
	}

	tq.mutex.Lock()
	defer tq.mutex.Unlock()
	return tq.pQueue.Update(t, uparg, uptick, t.GetTaskFn())
}

func (tq *TaskQueue) UpdateTaskTick(t *gameticktask.Task, uptick gametick.GameTick) error {
	if t == nil {
		tq.log.Fatal("failed to update nil task")
	}

	tq.mutex.Lock()
	defer tq.mutex.Unlock()
	return tq.pQueue.Update(t, t.Argument(), uptick, t.GetTaskFn())
}

func (tq *TaskQueue) Remove(t *gameticktask.Task) error {
	if t == nil {
		tq.log.Fatal("failed to remove nil task")
	}

	tq.mutex.Lock()
	defer tq.mutex.Unlock()
	return tq.pQueue.Remove(t)
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
}
