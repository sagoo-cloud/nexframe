package utils

import (
	"fmt"
	"time"
)

// GetWeekDay 获取本周的开始时间和结束时间
func GetWeekDay() (string, string) {
	now := time.Now()
	weekday := now.Weekday()

	// Calculate the start of the week (Monday)
	start := now.AddDate(0, 0, -int(weekday)+int(time.Monday))
	if weekday == time.Sunday {
		start = start.AddDate(0, 0, -7)
	}
	startStr := start.Format("2006-01-02") + " 00:00:00"

	// Calculate the end of the week (Sunday)
	end := start.AddDate(0, 0, 6)
	endStr := end.Format("2006-01-02") + " 23:59:59"

	return startStr, endStr
}

// GetBetweenDates 根据开始日期和结束日期计算出时间段内所有日期
// 参数为日期格式，如：2020-01-01
func GetBetweenDates(sdate, edate string) []string {
	var d []string
	timeFormatTpl := "2006-01-02"
	if len(timeFormatTpl) != len(sdate) {
		timeFormatTpl = timeFormatTpl[0:len(sdate)]
	}
	date, err := time.Parse(timeFormatTpl, sdate)
	if err != nil {
		// 时间解析，异常
		return d
	}
	date2, err := time.Parse(timeFormatTpl, edate)
	if err != nil {
		// 时间解析，异常
		return d
	}
	if date2.Before(date) {
		// 如果结束时间小于开始时间，异常
		return d
	}
	// 输出日期格式固定
	timeFormatTpl = "2006-01-02"
	date2Str := date2.Format(timeFormatTpl)
	d = append(d, date.Format(timeFormatTpl))
	for {
		date = date.AddDate(0, 0, 1)
		dateStr := date.Format(timeFormatTpl)
		d = append(d, dateStr)
		if dateStr == date2Str {
			break
		}
	}
	return d
}

func GetHourBetweenDates(sdate, edate string) []string {
	var d []string
	date, _ := time.Parse("2006-01-02 15:04:05", sdate)
	date2, _ := time.Parse("2006-01-02 15:04:05", edate)
	date2Str := date2.Format("2006-01-02 15:00:00")

	d = append(d, date.Format("2006-01-02 15:00:00"))
	for {
		date = date.Add(time.Hour)
		dateStr := date.Format("2006-01-02 15:00:00")
		d = append(d, dateStr)
		if dateStr == date2Str {
			break
		}
	}
	return d
}

// GetQuarterDay 获得当前季度的初始和结束日期
func GetQuarterDay() (string, string) {
	now := time.Now()
	year, quarter := now.Year(), (now.Month()-1)/3+1
	start := time.Date(year, (quarter-1)*3+1, 1, 0, 0, 0, 0, time.Local).Format("2006-01-02 15:04:05")
	end := time.Date(year, (quarter-1)*3+1+2, daysIn((quarter-1)*3+1+2, year), 23, 59, 59, 0, time.Local).Format("2006-01-02 15:04:05")
	return start, end
}

func daysIn(m time.Month, year int) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// GetTimeByType 根据类型获取开始时间、结束时间及差值 1 天 2 周 3 月 4 年
func GetTimeByType(types int) (index int, begin string, end string) {
	now := time.Now()
	switch types {
	case 1:
		begin = now.Format("2006-01-02 00:00:00")
		end = now.Format("2006-01-02 15:04:05")
		index = now.Hour() + 1
	case 2:
		begin = now.AddDate(0, 0, -7).Format("2006-01-02 00:00:00")
		end = now.Format("2006-01-02 15:04:05")
		index = int(now.Sub(now.AddDate(0, 0, -7)).Hours()/24) + 1
	case 3:
		begin = now.AddDate(0, 0, -30).Format("2006-01-02 00:00:00")
		end = now.Format("2006-01-02 15:04:05")
		index = int(now.Sub(now.AddDate(0, 0, -30)).Hours()/24) + 1
	case 4:
		begin = now.Format("2006-01-01 00:00:00")
		end = now.AddDate(1, 0, 0).Format("2006-01-01 00:00:00")
		index = int(now.Month())
	default:
		begin = now.Format("2006-01-02 00:00:00")
		end = now.Format("2006-01-02 15:04:05")
		index = now.Hour() + 1
	}
	return
}

// GetTime 根据类型和开始时间获取时间段及长度
func GetTime(i int, types int, begin string) (startTime string, endTime string, duration int, unit string) {
	t, _ := time.Parse("2006-01-02 15:04:05", begin)
	switch types {
	case 1:
		startTime = t.Add(time.Duration(i) * time.Hour).Format("2006-01-02 15:04:05")
		endTime = t.Add(time.Duration(i+1) * time.Hour).Format("2006-01-02 15:04:05")
		duration = t.Add(time.Duration(i) * time.Hour).Hour()
		unit = "时"
	case 2, 3:
		startTime = t.AddDate(0, 0, i).Format("2006-01-02 15:04:05")
		endTime = t.AddDate(0, 0, i+1).Format("2006-01-02 15:04:05")
		duration = t.AddDate(0, 0, i).Day()
		unit = "日"
	case 4:
		startTime = t.AddDate(0, i, 0).Format("2006-01-02 15:04:05")
		endTime = t.AddDate(0, i+1, 0).Format("2006-01-02 15:04:05")
		duration = int(t.AddDate(0, i, 0).Month())
		unit = "月"
	default:
		startTime = t.Add(time.Duration(i) * time.Hour).Format("2006-01-02 15:04:05")
		endTime = t.Add(time.Duration(i+1) * time.Hour).Format("2006-01-02 15:04:05")
		duration = t.Add(time.Duration(i) * time.Hour).Hour()
		unit = "时"
	}
	return
}

// CalcDaysFromYearMonth 返回给定年份和月份的天数
func CalcDaysFromYearMonth(year int, month int) int {
	return time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()
}

// GetTimeUnix 转为时间戳->秒数
func GetTimeUnix(t time.Time) int64 {
	return t.Unix()
}

// GetTimeMills 转为时间戳->毫秒数
func GetTimeMills(t time.Time) int64 {
	return t.UnixNano() / 1e6
}

// GetTimeByInt 时间戳转时间
func GetTimeByInt(t1 int64) time.Time {
	return time.Unix(t1, 0)
}

// GetHourDiffer  计算俩个时间差多少小时
func GetHourDiffer(startTime, endTime string) float64 {
	t1, err := time.Parse("2006-01-02 15:04:05", startTime)
	t2, err := time.Parse("2006-01-02 15:04:05", endTime)
	if err == nil && CompareTime(t1, t2) {
		diff := t2.Sub(t1)
		return diff.Hours()
	}
	return 0
}

// GetMinutesDiffer  计算俩个时间差多少分钟
func GetMinutesDiffer(startTime, endTime string) int {
	t1, _ := time.Parse("2006-01-02 15:04:05", startTime)
	t2, _ := time.Parse("2006-01-02 15:04:05", endTime)
	diff := t2.Sub(t1)
	return int(diff.Minutes())
}

// CompareTime 比较两个时间大小
func CompareTime(t1, t2 time.Time) bool {
	return t1.Before(t2)
}

// IsSameDay 是否为同一天
func IsSameDay(t1, t2 int64) bool {
	y1, m1, d1 := time.Unix(t1, 0).Date()
	y2, m2, d2 := time.Unix(t2, 0).Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// IsSameMinute 是否为同一分钟
func IsSameMinute(t1, t2 int64) bool {
	d1 := time.Unix(t1, 0).Format("2006-01-02 15:04")
	d2 := time.Unix(t2, 0).Format("2006-01-02 15:04")
	return d1 == d2
}

// GetCurrentDateString 获取当前日期字符串
func GetCurrentDateString() string {
	return time.Now().Format("2006-01-02")
}

// GetTimeTagGroup 获取当前日期字符串,结果：2024:05
func GetTimeTagGroup() string {
	now := time.Now()
	return fmt.Sprintf("%d:%02d:", now.Year(), now.Month())
}
