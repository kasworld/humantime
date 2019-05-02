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

// task (gametask, humantask) 실시간 통계
package taskstat

import (
	"sync"
	"time"
)

type StatObj struct {
	startTime time.Time
	statRef   *Stat
}

func (so *StatObj) Commit() {
	so.statRef.commit(so.startTime)
}

func (so *StatObj) Success() {
	so.statRef.success()
}

type Stat struct {
	mutex          sync.Mutex
	totalDur       time.Duration
	StartCount     int64
	SuccessCount   int64
	EndCount       int64
	HighMS         float64
	LowMS          float64
	lastUpdateTime time.Time
}

func (st *Stat) Open() *StatObj {
	st.mutex.Lock()
	st.StartCount++
	st.mutex.Unlock()
	return &StatObj{
		startTime: time.Now().UTC(),
		statRef:   st,
	}
}

func (st *Stat) commit(startTime time.Time) {
	cur := time.Now().UTC()
	dur := cur.Sub(startTime)

	st.mutex.Lock()
	st.EndCount++
	st.totalDur += dur

	if cur.Sub(st.lastUpdateTime) > time.Duration(10)*time.Second {
		st.lastUpdateTime = cur
		st.HighMS = 0.0
		st.LowMS = 100000.0
	}

	if st.HighMS < dur.Seconds()*1000 {
		st.HighMS = dur.Seconds() * 1000
	}
	if st.LowMS > dur.Seconds()*1000 {
		st.LowMS = dur.Seconds() * 1000
	}
	st.mutex.Unlock()
}
func (st *Stat) success() {
	st.mutex.Lock()
	st.SuccessCount++
	st.mutex.Unlock()
}
func (st Stat) FailCount() int64 {
	return st.EndCount - st.SuccessCount
}

func (st Stat) RunCount() int64 {
	return st.StartCount - st.EndCount
}

func (st Stat) Avg() float64 {
	if st.EndCount != 0 {
		return float64(st.totalDur.Nanoseconds()) / float64(st.EndCount*1000000)
	} else {
		return 0.0
	}
}

type TaskStat struct {
	mutex   sync.RWMutex
	taskMap map[string]*Stat
}

func New() *TaskStat {
	return &TaskStat{
		taskMap: make(map[string]*Stat),
	}
}

func (fm *TaskStat) GetStatByFuncName(fnname string) *StatObj {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()
	taskstat, ok := fm.taskMap[fnname]
	if ok {
		stobj := taskstat.Open()
		return stobj
	} else {
		taskstat = &Stat{
			LowMS:          100000.0,
			lastUpdateTime: time.Now().UTC(),
		}
		stobj := taskstat.Open()
		fm.taskMap[fnname] = taskstat
		return stobj
	}
}
