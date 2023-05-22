package timestamp

import (
	"fmt"
	"time"
)

var startTimeStamp [][]time.Time = make([][]time.Time, 50)
var endTimeStamp [][]time.Time = make([][]time.Time, 50)

func AddStartTime(id int) {
	startTimeStamp[id] = append(startTimeStamp[id], time.Now())
}

func AddEndTime(id int) {
	endTimeStamp[id] = append(endTimeStamp[id], time.Now())
}

func LenStartTimeStamp(id int) int {
	return len(startTimeStamp[id])
}

func LenEndTimeStamp(id int) int {
	return len(endTimeStamp[id])
}

func GetDiffAvg(id int) int64 {
	diffList := GetDiffList(id)

	if len(diffList) == 0 {
		return 0
	}
	total := int64(0)
	for _, v := range diffList {
		total += v
	}

	return total / int64(len(diffList))
}

func GetDiffList(id int) []int64 {
	var diffList []int64

	min := len(startTimeStamp[id])
	if min > len(endTimeStamp[id]) {
		min = len(endTimeStamp[id])
	}

	for i := 0; i < min; i++ {
		diffList = append(diffList, endTimeStamp[id][i].Sub(startTimeStamp[id][i]).Microseconds())
	}

	return diffList
}

func PrintStartStamp(id int) {
	stamp := startTimeStamp[id]

	curT := time.Now()
	baseT := time.Date(curT.Year(), curT.Month(), curT.Day(), curT.Hour(), 0, 0, 0, time.Local)

	for _, t := range stamp {
		fmt.Println(t.Sub(baseT).Milliseconds())
	}
}

func PrintEndStamp(id int) {
	stamp := endTimeStamp[id]

	curT := time.Now()
	baseT := time.Date(curT.Year(), curT.Month(), curT.Day(), curT.Hour(), 0, 0, 0, time.Local)

	for _, t := range stamp {
		fmt.Println(t.Sub(baseT).Milliseconds())
	}
}
