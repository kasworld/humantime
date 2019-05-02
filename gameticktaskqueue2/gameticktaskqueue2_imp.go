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

	"github.com/kasworld/gametick"
	"github.com/kasworld/timedtask/gameticktask"
	"github.com/kasworld/timedtask/taskstat"
)

func (tq *TaskQueue) FlushTaskTill(till gametick.GameTick) {
	tq.runTasksEndWaitGroup.Wait()
	processed := 0
	tq.log.TraceService("Start FlushTaskTill %v", tq)
	defer func() { tq.log.TraceService("End FlushTaskTill %v, %v", processed, tq) }()
	for {
		peeked := tq.Peek()
		if peeked == nil { // no task to do
			return
		}
		if till < peeked.TaskGameTick() { // no current task
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

func (tq *TaskQueue) runWaitTask(t *gameticktask.Task) {
	defer tq.runTasksEndWaitGroup.Done()
	tq.runStat.Inc()
	tso := tq.taskStat.GetStatByFuncName(t.GetTaskFnName())
	if err := t.RunWithStat(tso); err != nil {
		tq.log.Error("%v", err)
	}
}

func (tq *TaskQueue) IsPaused() bool {
	return tq.paused
}

func (tq *TaskQueue) GetTaskStat() *taskstat.TaskStat {
	return tq.taskStat
}

func (tq *TaskQueue) Peek() *gameticktask.Task {
	tq.mutex.RLock()
	if len(tq.pQueue) > 0 {
		t := tq.pQueue[0]
		tq.mutex.RUnlock()
		return t
	}
	tq.mutex.RUnlock()
	return nil
}

func (tq *TaskQueue) Pop() *gameticktask.Task {
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
