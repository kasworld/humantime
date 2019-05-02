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

const (
	timeDurationYear time.Duration = time.Hour * 24 * 365
)

type TaskQueue struct {
	mutex                sync.RWMutex
	runTasksEndWaitGroup sync.WaitGroup // 실행중인 task가 모두 끝났음을 보장

	logger   loggeri.LoggerI
	Name     string
	runStat  *actpersec.ActPerSec
	taskStat *taskstat.TaskStat
	pQueue   humantimetask.TaskList
	paused   bool

	popDelay  time.Duration
	tasktimer *time.Timer
}

func New(name string, popDelay time.Duration, logger loggeri.LoggerI) *TaskQueue {
	tq := &TaskQueue{
		logger:    logger,
		Name:      name,
		runStat:   actpersec.New(),
		taskStat:  taskstat.New(),
		pQueue:    make(humantimetask.TaskList, 0),
		popDelay:  popDelay,
		tasktimer: time.NewTimer(timeDurationYear), // after a year
	}
	return tq
}

func (tq TaskQueue) String() string {
	pstr := TrueString(tq.paused, "paused", "running")
	return fmt.Sprintf(
		"HumanTimeTaskQueue2[%v %s %v %v]",
		tq.Name, pstr, tq.Len(), tq.runStat)
}

func TrueString(b bool, truestr, falsestr string) string {
	if b {
		return truestr
	} else {
		return falsestr
	}
}
