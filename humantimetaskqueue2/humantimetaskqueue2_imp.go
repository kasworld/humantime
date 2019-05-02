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

// humantimetask를 관리 실행해주는 관리자.
package humantimetaskqueue2

import (
	"container/heap"
	"fmt"
	"time"

	"github.com/kasworld/actpersec"
	"github.com/kasworld/timedtask/humantimetask"
	"github.com/kasworld/timedtask/taskstat"
)

func (tq TaskQueue) IsPaused() bool {
	return tq.paused
}

func (tq *TaskQueue) GetActStat() *actpersec.ActPerSec {
	return tq.runStat
}

func (tq TaskQueue) GetTaskStat() *taskstat.TaskStat {
	return tq.taskStat
}

func (tq *TaskQueue) Peek() *humantimetask.Task {
	tq.mutex.RLock()
	if len(tq.pQueue) > 0 {
		t := tq.pQueue[0]
		tq.mutex.RUnlock()
		return t
	}
	tq.mutex.RUnlock()
	return nil
}

func (tq *TaskQueue) Pop() *humantimetask.Task {
	tq.mutex.Lock()
	if len(tq.pQueue) == 0 {
		tq.mutex.Unlock()
		return nil
	}
	t := tq.pQueue[0]
	_ = heap.Pop(&tq.pQueue)
	tq.mutex.Unlock()
	return t
}

func (tq *TaskQueue) Len() int {
	return tq.pQueue.Len()
}

func (tq *TaskQueue) Pause() {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()
	if tq.paused {
		return
	}
	tq.paused = true
	tq.tasktimer.Reset(timeDurationYear)
	tq.logger.TraceService("%v paused", tq)
}

func (tq *TaskQueue) Resume() {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()
	if !tq.paused {
		return
	}
	tq.paused = false
	tq.scheduleTimerAtRootTick()
	tq.logger.TraceService("%v resumed", tq)
}

func (tq *TaskQueue) UpdateTaskArgAndTime(t *humantimetask.Task, uparg interface{}, uptime time.Time) error {
	if t == nil {
		tq.logger.Fatal("failed to update nil task")
	}
	tq.mutex.Lock()
	defer tq.mutex.Unlock()
	return tq.update(t, uparg, uptime)
}

func (tq *TaskQueue) UpdateTaskTime(t *humantimetask.Task, uptime time.Time) error {
	if t == nil {
		tq.logger.Fatal("failed to update nil task")
	}
	tq.mutex.Lock()
	defer tq.mutex.Unlock()
	return tq.update(t, t.Argument(), uptime)
}

func (tq *TaskQueue) update(t *humantimetask.Task, uparg interface{}, uptime time.Time) error {
	if len(tq.pQueue) > 0 {
		oldRootTick := tq.pQueue[0].TaskTime()
		if err := tq.pQueue.Update(t, uparg, uptime, t.GetTaskFn()); err != nil {
			return err
		}
		newRootTick := tq.pQueue[0].TaskTime()
		if oldRootTick != newRootTick {
			tq.scheduleTimerAtRootTick()
		}
	} else {
		return fmt.Errorf("%v update failed, no items enqueued", tq)
	}
	return nil
}

func (tq *TaskQueue) Remove(t *humantimetask.Task) error {
	if t == nil {
		tq.logger.Fatal("failed to remove nil task")
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

func (tq *TaskQueue) Push(t *humantimetask.Task) {
	if t == nil {
		tq.logger.Fatal("%v tried to push nil task", tq)
	}

	tq.mutex.Lock()
	defer tq.mutex.Unlock()
	if t.IsValid() {
		tq.logger.Fatal("%v tried to push %v already pushed", tq, t)
	}
	heap.Push(&tq.pQueue, t)
	if tq.pQueue[0] == t {
		tq.scheduleTimerAtRootTick()
	}
}
