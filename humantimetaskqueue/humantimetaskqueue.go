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
package humantimetaskqueue

import (
	"container/heap"
	"fmt"
	"sync"
	"time"

	"github.com/kasworld/actpersec"
	"github.com/kasworld/timedtask/humantimetask"
	"github.com/kasworld/timedtask/humantimetaskqueuei"
	"github.com/kasworld/timedtask/loggeri"
	"github.com/kasworld/timedtask/taskstat"
)

var _ humantimetaskqueuei.TaskQueueI = &TaskQueue{}

type TaskQueue struct {
	mutex sync.RWMutex
	log   loggeri.LoggerI `webformhide:"" stringformhide:""`

	runTasksEndWaitGroup sync.WaitGroup // 실행중인 task가 모두 끝났음을 보장
	paused               bool
	runStat              *actpersec.ActPerSec
	pQueue               humantimetask.TaskList
	Name                 string
	repeatWait           time.Duration
	popDelay             time.Duration
	taskStat             *taskstat.TaskStat
}

func New(name string, popDelay time.Duration, repeatWait time.Duration, l loggeri.LoggerI) *TaskQueue {
	tq := &TaskQueue{
		log:        l,
		pQueue:     make(humantimetask.TaskList, 0),
		Name:       name,
		popDelay:   popDelay,
		repeatWait: repeatWait,
		taskStat:   taskstat.New(),
		runStat:    actpersec.New(),
	}
	return tq
}

func (tq TaskQueue) String() string {
	if tq.paused {
		return fmt.Sprintf(
			"HumanTimeTaskQueue[%v paused %v %v]",
			tq.Name, tq.Len(), tq.runStat)
	} else {
		return fmt.Sprintf(
			"HumanTimeTaskQueue[%v running %v %v]",
			tq.Name, tq.Len(), tq.runStat)
	}
}

func (tq *TaskQueue) Pause() {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()
	tq.paused = true
}

func (tq *TaskQueue) Resume() {
	tq.mutex.Lock()
	defer tq.mutex.Unlock()
	tq.paused = false
}

func (tq TaskQueue) IsPaused() bool {
	return tq.paused
}

func (tq *TaskQueue) GetActStat() *actpersec.ActPerSec {
	return tq.runStat
}

func (tq TaskQueue) GetTaskStat() *taskstat.TaskStat {
	return tq.taskStat
}

func (tq *TaskQueue) UpdateTaskArgAndTime(t *humantimetask.Task, uparg interface{}, uptick time.Time) error {
	if t == nil {
		tq.log.Fatal("failed to update nil task")
	}
	tq.mutex.Lock()
	defer tq.mutex.Unlock()
	return tq.pQueue.Update(t, uparg, uptick, t.GetTaskFn())
}

func (tq *TaskQueue) UpdateTaskTime(t *humantimetask.Task, uptime time.Time) error {
	if t == nil {
		tq.log.Fatal("failed to update nil task")
	}
	tq.mutex.Lock()
	defer tq.mutex.Unlock()
	return tq.pQueue.Update(t, t.Argument(), uptime, t.GetTaskFn())
}

func (tq *TaskQueue) Remove(t *humantimetask.Task) error {
	if t == nil {
		tq.log.Fatal("failed to remove nil task")
	}

	tq.mutex.Lock()
	defer tq.mutex.Unlock()
	return tq.pQueue.Remove(t)
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

func (tq *TaskQueue) Push(t *humantimetask.Task) {
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

func (tq *TaskQueue) Len() int {
	return tq.pQueue.Len()
}
