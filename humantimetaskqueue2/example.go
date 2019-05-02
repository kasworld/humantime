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

// +build ignore

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/kasworld/humantimetask/humantimetask"
	"github.com/kasworld/humantimetask/humantimetaskqueue2"
	"github.com/kasworld/log/globalloglevels/globallogger"
)

func main() {
	q := humantimetaskqueue2.New("TQ", time.Second, globallogger.GlobalLogger)
	for i := 0; i < 10; i++ {
		q.Push(
			humantimetask.New(
				time.Now().Add(time.Duration(i)*time.Second),
				i, exTaskFn,
			),
		)
	}
	fmt.Printf("%v", q)
	ctx := context.Background()
	go q.Run(ctx)
	time.Sleep(time.Second * 10)
}

func exTaskFn(tk *humantimetask.Task) error {
	fmt.Printf("exTaskFn %v\n", tk.Argument())
	return nil
}
