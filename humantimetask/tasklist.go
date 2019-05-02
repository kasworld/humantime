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

package humantimetask

import (
	"container/heap"
	"fmt"
	"time"
)

// heap implementation

type TaskList []*Task

func (fh TaskList) Len() int { return len(fh) }

func (fh TaskList) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, tasktime so we use greater than here.
	return fh[i].tasktime.Before(fh[j].tasktime)
}

func (fh TaskList) Swap(i, j int) {
	fh[i], fh[j] = fh[j], fh[i]
	fh[i].index = i
	fh[j].index = j
}

func (fh *TaskList) Push(x interface{}) {
	n := len(*fh)
	item := x.(*Task)
	item.index = n
	*fh = append(*fh, item)
}

func (fh *TaskList) Pop() interface{} {
	old := *fh
	n := len(old)
	item := old[n-1]
	item.index = invalidTaskIndex // for safety
	*fh = old[0 : n-1]
	return item
}

func (fh *TaskList) Fix() {
	heap.Fix(fh, 0)
}

func (fh *TaskList) Update(
	item *Task, argument interface{}, tasktime time.Time, fn DoTaskFn) error {

	if item.index == invalidTaskIndex {
		return fmt.Errorf("not found item in queue: %v", item)
	}
	oldtick := item.tasktime

	item.argument = argument
	item.tasktime = tasktime
	item.doTaskFn = fn

	if oldtick != tasktime {
		heap.Fix(fh, item.index)
	}
	return nil
}

func (fh *TaskList) Remove(item *Task) error {
	if item.index == invalidTaskIndex {
		return fmt.Errorf("not found item in queue: %v", item)
	}
	heap.Remove(fh, item.index)
	return nil
}
