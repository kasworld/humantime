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

// humantimetaskqueue interface
package humantimetaskqueuei

import (
	"context"
	"time"

	"github.com/kasworld/timedtask/humantimetask"
	"github.com/kasworld/timedtask/taskstat"
)

type TaskQueueI interface {
	Pause()
	Resume()
	GetTaskStat() *taskstat.TaskStat
	UpdateTaskTime(t *humantimetask.Task, uptime time.Time) error
	Remove(t *humantimetask.Task) error
	Push(t *humantimetask.Task)
	Len() int
	Run(ctx context.Context)
	FlushTaskTill(till time.Time)
}
