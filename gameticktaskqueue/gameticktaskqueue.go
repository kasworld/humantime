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

// gameticktask를 관리 실행해주는 관리자.
package gameticktaskqueue

import (
	"fmt"
	"sync"
	"time"

	"github.com/kasworld/actpersec"
	"github.com/kasworld/gametick"
	"github.com/kasworld/timedtask/gameticktask"
	"github.com/kasworld/timedtask/gameticktaskqueuei"
	"github.com/kasworld/timedtask/loggeri"
	"github.com/kasworld/timedtask/taskstat"
)

var _ gameticktaskqueuei.TaskQueueI = &TaskQueue{}

type TaskQueue struct {
	mutex sync.RWMutex
	log   loggeri.LoggerI

	runTasksEndWaitGroup sync.WaitGroup // 실행중인 task가 모두 끝났음을 보장
	paused               bool
	runStat              *actpersec.ActPerSec
	pQueue               gameticktask.TaskList
	Name                 string
	repeatWait           time.Duration
	popDelay             gametick.GameTick
	taskStat             *taskstat.TaskStat
}

func New(
	name string,
	popDelay time.Duration,
	repeatWait time.Duration,
	logger loggeri.LoggerI) *TaskQueue {

	tq := &TaskQueue{
		pQueue:     make(gameticktask.TaskList, 0),
		Name:       name,
		popDelay:   gametick.FromTimeDurationToTickType(popDelay),
		repeatWait: repeatWait,
		taskStat:   taskstat.New(),
		runStat:    actpersec.New(),
		log:        logger,
	}
	return tq
}

func (tq TaskQueue) String() string {
	if tq.paused {
		return fmt.Sprintf(
			"GameTickTaskQueue[%v paused %v %v]",
			tq.Name, tq.Len(), tq.runStat)
	} else {
		return fmt.Sprintf(
			"GameTickTaskQueue[%v running %v %v]",
			tq.Name, tq.Len(), tq.runStat)
	}
}
