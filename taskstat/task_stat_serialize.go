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

package taskstat

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

func (fm TaskStat) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "\n task time Cnt \n\n")
	for k, v := range fm.taskMap {
		fmt.Fprintf(
			&buf,
			" Avg ms : |%13.6f| StartCount : |%15v| EndCount : |%15v| SuccessCount : |%15v| failCount : |%15v| High ms(10s) : |%13.6f| Low ms(10s) : |%13.6f| funcName : %s\n",
			v.Avg(), v.StartCount, v.EndCount, v.SuccessCount, v.StartCount-v.SuccessCount, v.HighMS, v.LowMS, k)
	}
	fmt.Fprintf(&buf, "\n")
	return buf.String()
}

const (
	HTML_tableheader = `<tr>
<th>Avg ms</th>
<th>StartCount</th>
<th>EndCount</th>
<th>RunCount</th>
<th>SuccessCount</th>
<th>failCount</th>
<th>High ms(last 10s)</th>
<th>Low ms(last 10s)</th>
<th>funcName</th>
</tr>`
	HTML_row = `<tr>
<td>{{printf "%13.6f" $v.Avg }}</td>
<td>{{$v.StartCount}}</td>
<td>{{$v.EndCount}}</td>
<td>{{$v.RunCount}}</td>
<td>{{$v.SuccessCount}}</td>
<td>{{$v.FailCount }}</td>
<td>{{printf "%13.6f" $v.HighMS }}</td>
<td>{{printf "%13.6f" $v.LowMS}}</td>
<td>{{$i}}</td>
</tr>
`
)

func (fm TaskStat) ToWeb(w http.ResponseWriter, r *http.Request) error {
	tplIndex, err := template.New("index").Parse(`
	<html>
	<head>
	<title>Task Stat</title>
	</head>
	<body>

	<table border=1 style="border-collapse:collapse;">` +
		HTML_tableheader +
		`{{range $i, $v := .}}` +
		HTML_row +
		`{{end}}` +
		HTML_tableheader +
		`</table>
	<br/>
	</body>
	</html>
	`)
	if err != nil {
		return err
	}
	if err := tplIndex.Execute(w, fm.taskMap); err != nil {
		return err
	}
	return nil
}
