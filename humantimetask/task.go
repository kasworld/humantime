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

// 미리 지정 된 humantime에 실행 되어야 하는 task
package humantimetask

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/kasworld/timedtask/taskstat"
)

// An Task is something we manage in a tasktime queue.

const invalidTaskIndex = -1

type DoTaskFn func(*Task) error

type Task struct {
	fnName   string
	argument interface{}
	doTaskFn DoTaskFn  // Task do function
	tasktime time.Time // The tasktime of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

func New(tasktime time.Time, argument interface{}, doTaskFn DoTaskFn) *Task {
	taskfnname := runtime.FuncForPC(reflect.ValueOf(doTaskFn).Pointer()).Name()
	ft := Task{
		fnName:   taskfnname,
		doTaskFn: doTaskFn,
		tasktime: tasktime,
		argument: argument,
		index:    invalidTaskIndex,
	}
	return &ft
}

func (ft Task) String() string {
	return fmt.Sprintf(
		"HumanTimeTask[%v at %v]",
		ft.GetTaskFnName(), ft.tasktime.Format("2006-01-02T15:04:05Z07:00"))
}

func (ft *Task) PanicString() string {
	return fmt.Sprintf("HumanTimeTask[%#v]", ft)
}

func (ft *Task) GetTaskFn() DoTaskFn {
	return ft.doTaskFn
}

func (ft *Task) TaskTime() time.Time {
	return ft.tasktime
}

func (ft *Task) Argument() interface{} {
	return ft.argument
}

func (ft *Task) GetTaskFnName() string {
	return ft.fnName
}

func (ft *Task) IsValid() bool {
	return ft.index != invalidTaskIndex
}
func (ft *Task) RunWithStat(ts *taskstat.StatObj) error {
	defer RecoverPanic(ft)
	err := ft.GetTaskFn()(ft)
	ts.Commit()
	if err != nil {
		return fmt.Errorf("%v %v", ft, err)
	} else {
		ts.Success()
	}
	return nil
}

func RecoverPanic(obj *Task) {
	if r := recover(); r != nil {
		errMsg := fmt.Sprintf(
			"RecoverPanic %v\n\n%v\n\n%s\n\n%s",
			time.Now().UTC(),
			obj.PanicString(),
			r,
			string(debug.Stack()))
		os.Stderr.WriteString(errMsg)

		// syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}
}
