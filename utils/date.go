package utils

import (
	"fmt"
	"sync"
	"time"
)

var (
	timePool   sync.Pool
	bufferPool sync.Pool
)

func init() {
	timePool.New = func() interface{} {
		return new(time.Time)
	}
	bufferPool.New = func() interface{} {
		return make([]byte, 0, 64)
	}
}

// getTimeFromPool 从对象池获取时间对象
func getTimeFromPool() *time.Time {
	return timePool.Get().(*time.Time)
}

// putTimeToPool 将时间对象放回对象池
func putTimeToPool(t *time.Time) {
	timePool.Put(t)
}

// getBufferFromPool 从对象池获取缓冲区
func getBufferFromPool() []byte {
	return bufferPool.Get().([]byte)[:0]
}

// putBufferToPool 将缓冲区放回对象池
func putBufferToPool(b []byte) {
	bufferPool.Put(b)
}

// GetWeekDay 获取本周的开始时间和结束时间
func GetWeekDay() (start, end string) {
	now := time.Now()
	weekday := now.Weekday()

	startTime := getTimeFromPool()
	endTime := getTimeFromPool()
	defer putTimeToPool(startTime)
	defer putTimeToPool(endTime)

	*startTime = now.AddDate(0, 0, -int(weekday)+int(time.Monday))
	if weekday == time.Sunday {
		*startTime = startTime.AddDate(0, 0, -7)
	}
	*endTime = startTime.AddDate(0, 0, 6)

	startBuf := getBufferFromPool()
	endBuf := getBufferFromPool()
	defer putBufferToPool(startBuf)
	defer putBufferToPool(endBuf)

	startBuf = startTime.AppendFormat(startBuf, "2006-01-02 00:00:00")
	endBuf = endTime.AppendFormat(endBuf, "2006-01-02 23:59:59")

	return string(startBuf), string(endBuf)
}

// GetBetweenDates 根据开始日期和结束日期计算出时间段内所有日期
func GetBetweenDates(sdate, edate string) []string {
	start, err := time.Parse("2006-01-02", sdate)
	if err != nil {
		return nil
	}
	end, err := time.Parse("2006-01-02", edate)
	if err != nil {
		return nil
	}
	if end.Before(start) {
		return nil
	}

	days := int(end.Sub(start).Hours()/24) + 1
	dates := make([]string, 0, days)
	current := getTimeFromPool()
	defer putTimeToPool(current)
	*current = start

	buf := getBufferFromPool()
	defer putBufferToPool(buf)

	for i := 0; i < days; i++ {
		buf = current.AppendFormat(buf[:0], "2006-01-02")
		dates = append(dates, string(buf))
		*current = current.AddDate(0, 0, 1)
	}

	return dates
}

// GetHourBetweenDates 获取两个日期之间的所有小时
func GetHourBetweenDates(sdate, edate string) []string {
	start, err := time.Parse("2006-01-02 15:04:05", sdate)
	if err != nil {
		return nil
	}
	end, err := time.Parse("2006-01-02 15:04:05", edate)
	if err != nil {
		return nil
	}

	hours := int(end.Sub(start).Hours()) + 1
	dates := make([]string, 0, hours)
	current := getTimeFromPool()
	defer putTimeToPool(current)
	*current = start

	buf := getBufferFromPool()
	defer putBufferToPool(buf)

	for i := 0; i < hours; i++ {
		buf = current.AppendFormat(buf[:0], "2006-01-02 15:00:00")
		dates = append(dates, string(buf))
		*current = current.Add(time.Hour)
	}

	return dates
}

// GetQuarterDay 获得当前季度的初始和结束日期
func GetQuarterDay() (start, end string) {
	now := time.Now()
	currentQuarter := (now.Month()-1)/3 + 1

	startTime := getTimeFromPool()
	endTime := getTimeFromPool()
	defer putTimeToPool(startTime)
	defer putTimeToPool(endTime)

	*startTime = time.Date(now.Year(), (currentQuarter-1)*3+1, 1, 0, 0, 0, 0, time.Local)
	*endTime = startTime.AddDate(0, 3, -1)

	startBuf := getBufferFromPool()
	endBuf := getBufferFromPool()
	defer putBufferToPool(startBuf)
	defer putBufferToPool(endBuf)

	startBuf = startTime.AppendFormat(startBuf, "2006-01-02 15:04:05")
	endBuf = endTime.AppendFormat(endBuf, "2006-01-02 15:04:05")

	return string(startBuf), string(endBuf)
}

// GetTimeByType 根据类型获取开始时间、结束时间及差值 1 天 2 周 3 月 4 年
func GetTimeByType(types int) (index int, begin, end string) {
	now := time.Now()
	beginTime := getTimeFromPool()
	endTime := getTimeFromPool()
	defer putTimeToPool(beginTime)
	defer putTimeToPool(endTime)

	switch types {
	case 1:
		*beginTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		*endTime = now
		index = now.Hour() + 1
	case 2:
		*beginTime = now.AddDate(0, 0, -7)
		*endTime = now
		index = int(now.Sub(*beginTime).Hours()/24) + 1
	case 3:
		*beginTime = now.AddDate(0, 0, -30)
		*endTime = now
		index = int(now.Sub(*beginTime).Hours()/24) + 1
	case 4:
		*beginTime = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.Local)
		*endTime = beginTime.AddDate(1, 0, 0)
		index = int(now.Month())
	default:
		*beginTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		*endTime = now
		index = now.Hour() + 1
	}

	beginBuf := getBufferFromPool()
	endBuf := getBufferFromPool()
	defer putBufferToPool(beginBuf)
	defer putBufferToPool(endBuf)

	beginBuf = beginTime.AppendFormat(beginBuf, "2006-01-02 15:04:05")
	endBuf = endTime.AppendFormat(endBuf, "2006-01-02 15:04:05")

	return index, string(beginBuf), string(endBuf)
}

// GetTime 根据类型和开始时间获取时间段及长度
func GetTime(i, types int, begin string) (startTime, endTime string, duration int, unit string) {
	t, err := time.Parse("2006-01-02 15:04:05", begin)
	if err != nil {
		return "", "", 0, ""
	}

	startT := getTimeFromPool()
	endT := getTimeFromPool()
	defer putTimeToPool(startT)
	defer putTimeToPool(endT)

	switch types {
	case 1:
		*startT = t.Add(time.Duration(i) * time.Hour)
		*endT = startT.Add(time.Hour)
		duration = startT.Hour()
		unit = "时"
	case 2, 3:
		*startT = t.AddDate(0, 0, i)
		*endT = startT.AddDate(0, 0, 1)
		duration = startT.Day()
		unit = "日"
	case 4:
		*startT = t.AddDate(0, i, 0)
		*endT = startT.AddDate(0, 1, 0)
		duration = int(startT.Month())
		unit = "月"
	default:
		*startT = t.Add(time.Duration(i) * time.Hour)
		*endT = startT.Add(time.Hour)
		duration = startT.Hour()
		unit = "时"
	}

	startBuf := getBufferFromPool()
	endBuf := getBufferFromPool()
	defer putBufferToPool(startBuf)
	defer putBufferToPool(endBuf)

	startBuf = startT.AppendFormat(startBuf, "2006-01-02 15:04:05")
	endBuf = endT.AppendFormat(endBuf, "2006-01-02 15:04:05")

	return string(startBuf), string(endBuf), duration, unit
}

// CalcDaysFromYearMonth 返回给定年份和月份的天数
func CalcDaysFromYearMonth(year, month int) int {
	t := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC)
	return t.Day()
}

// GetTimeUnix 转为时间戳（秒数）
func GetTimeUnix(t time.Time) int64 {
	return t.Unix()
}

// GetTimeMills 转为时间戳（毫秒数）
func GetTimeMills(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

// GetTimeByInt 时间戳转时间
func GetTimeByInt(t int64) time.Time {
	return time.Unix(t, 0)
}

// GetHourDiffer 计算两个时间差多少小时
func GetHourDiffer(startTime, endTime string) float64 {
	t1, err := time.Parse("2006-01-02 15:04:05", startTime)
	if err != nil {
		return 0
	}
	t2, err := time.Parse("2006-01-02 15:04:05", endTime)
	if err != nil {
		return 0
	}
	if t2.After(t1) {
		return t2.Sub(t1).Hours()
	}
	return 0
}

// GetMinutesDiffer 计算两个时间差多少分钟
func GetMinutesDiffer(startTime, endTime string) int {
	t1, err := time.Parse("2006-01-02 15:04:05", startTime)
	if err != nil {
		return 0
	}
	t2, err := time.Parse("2006-01-02 15:04:05", endTime)
	if err != nil {
		return 0
	}
	return int(t2.Sub(t1).Minutes())
}

// CompareTime 比较两个时间大小
func CompareTime(t1, t2 time.Time) bool {
	return t1.Before(t2)
}

// IsSameDay 是否为同一天
func IsSameDay(t1, t2 int64) bool {
	time1 := time.Unix(t1, 0)
	time2 := time.Unix(t2, 0)
	return time1.Year() == time2.Year() && time1.Month() == time2.Month() && time1.Day() == time2.Day()
}

// IsSameMinute 是否为同一分钟
func IsSameMinute(t1, t2 int64) bool {
	time1 := time.Unix(t1, 0)
	time2 := time.Unix(t2, 0)
	return time1.Format("2006-01-02 15:04") == time2.Format("2006-01-02 15:04")
}

// GetCurrentDateString 获取当前日期字符串
func GetCurrentDateString() string {
	buf := getBufferFromPool()
	defer putBufferToPool(buf)
	return string(time.Now().AppendFormat(buf, "2006-01-02"))
}

// GetTimeTagGroup 获取当前日期字符串,结果：2024:05
func GetTimeTagGroup() string {
	now := time.Now()
	buf := getBufferFromPool()
	defer putBufferToPool(buf)
	buf = fmt.Appendf(buf, "%d:%02d:", now.Year(), now.Month())
	return string(buf)
}

// Time 获得当前时间戳单位s
func Time() int64 {
	return time.Now().Unix()
}

// TimeByTime 获得时间戳
func TimeByTime(t time.Time) int64 {
	return t.Unix()
}

// TimeLong 获得当前时间戳到毫秒
func TimeLong() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// TimeToStr 将时间戳转字符串
func TimeToStr(t int64) string {
	buf := getBufferFromPool()
	defer putBufferToPool(buf)
	return string(time.Unix(t, 0).AppendFormat(buf, "2006-01-02 15:04:05"))
}

// TimeLongToStr 将时间戳转字符串
func TimeLongToStr(t int64) string {
	buf := getBufferFromPool()
	defer putBufferToPool(buf)
	return string(time.Unix(t/1000, 0).AppendFormat(buf, "2006-01-02 15:04:05"))
}

// StrToTime 将字符串转时间戳
func StrToTime(date string) (int64, error) {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", date, time.Local)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// StrToTimeFormat 将字符串转时间戳按格式
func StrToTimeFormat(date, format string) (int64, error) {
	t, err := time.ParseInLocation(format, date, time.Local)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// StrToTimeFormatByLocal 将字符串转时间戳按格式和时区
func StrToTimeFormatByLocal(date, format, localName string) (int64, error) {
	loc, err := time.LoadLocation(localName)
	if err != nil {
		return 0, err
	}
	t, err := time.ParseInLocation(format, date, loc)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// Now 获得当前时间
func Now() string {
	buf := getBufferFromPool()
	defer putBufferToPool(buf)
	return string(time.Now().AppendFormat(buf, "2006-01-02 15:04:05"))
}

// TimeToFormat 将时间戳转字符串并格式化
func TimeToFormat(t int64, format string) string {
	buf := getBufferFromPool()
	defer putBufferToPool(buf)
	return string(time.Unix(t, 0).AppendFormat(buf, format))
}

// TimeFormat 将当前时间转字符串
func TimeFormat(format string) string {
	buf := getBufferFromPool()
	defer putBufferToPool(buf)
	return string(time.Now().AppendFormat(buf, format))
}

// UTCTime 获得UTC时间字符串
func UTCTime() string {
	buf := getBufferFromPool()
	defer putBufferToPool(buf)
	return string(time.Now().UTC().AppendFormat(buf, time.RFC3339))
}

// ZeroTime 获得凌晨零点时间戳
func ZeroTime() int64 {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
}

// ZeroTimeByLocal 获得指定时区凌晨零点时间戳
func ZeroTimeByLocal(localName string) (int64, error) {
	loc, err := time.LoadLocation(localName)
	if err != nil {
		return 0, err
	}
	now := time.Now().In(loc)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).Unix(), nil
}

// ZeroTimeByTimeByLocal 获得时间戳对应的凌晨0点时间戳（指定时区）
func ZeroTimeByTimeByLocal(ti int64, localName string) (int64, error) {
	loc, err := time.LoadLocation(localName)
	if err != nil {
		return 0, err
	}
	t := time.Unix(ti, 0).In(loc)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc).Unix(), nil
}

// ZeroTimeByTime 获得时间戳对应的凌晨0点时间戳
func ZeroTimeByTime(ti int64) int64 {
	t := time.Unix(ti, 0)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).Unix()
}

// EndTime 获得今天结束时间戳
func EndTime() int64 {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location()).Unix()
}

// EndTimeByTime 获得时间戳对应的23点59分59秒时间戳
func EndTimeByTime(ti int64) int64 {
	t := time.Unix(ti, 0)
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location()).Unix()
}

// StartAndEndDayTime 获得时间戳对应的起始时间和结束时间
func StartAndEndDayTime(ti int64) [2]int64 {
	return [2]int64{ZeroTimeByTime(ti), EndTimeByTime(ti)}
}

// GetYear 获得当前时间的年份
func GetYear() int {
	return time.Now().Year()
}

// GetYearByTimeLong 获得指定时间戳的年份
func GetYearByTimeLong(t int64) int {
	return time.Unix(t/1000, 0).Year()
}

// GetYearByTimeInt 获得指定时间戳的年份
func GetYearByTimeInt(t int64) int {
	return time.Unix(t, 0).Year()
}

// GetMonthFirstDay 获得当前月份的第一天日期
func GetMonthFirstDay() int64 {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Unix()
}

// GetMonthFirstDayByMonth 获得指定月份的第一天日期
func GetMonthFirstDayByMonth(year int, month time.Month) int64 {
	return time.Date(year, month, 1, 0, 0, 0, 0, time.Local).Unix()
}

// GetMonthLastDay 获得当前月份的最后一天日期
func GetMonthLastDay() int64 {
	now := time.Now()
	lastDay := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 999999999, now.Location())
	return lastDay.Unix()
}

// GetMonthLastDayByMonth 获得指定月份的最后一天日期
func GetMonthLastDayByMonth(year int, month time.Month) int64 {
	lastDay := time.Date(year, month+1, 0, 23, 59, 59, 999999999, time.Local)
	return lastDay.Unix()
}

// GetMonth 获得当前时间的月份
func GetMonth() time.Month {
	return time.Now().Month()
}

// GetMonthByTimeLong 获得指定时间戳的月份
func GetMonthByTimeLong(t int64) time.Month {
	return time.Unix(t/1000, 0).Month()
}

// GetMonthByTimeInt 获得指定时间戳的月份
func GetMonthByTimeInt(t int64) time.Month {
	return time.Unix(t, 0).Month()
}

// GetDay 获得当前时间的天数
func GetDay() int {
	return time.Now().Day()
}

// GetDayByTimeLong 获得指定时间戳的天数
func GetDayByTimeLong(t int64) int {
	return time.Unix(t/1000, 0).Day()
}

// GetDayByTimeInt 获得指定时间戳的天数
func GetDayByTimeInt(t int64) int {
	return time.Unix(t, 0).Day()
}

// GetTimeByDay 获得当月指定日期的时间戳
func GetTimeByDay(day int) int64 {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), day, now.Hour(), now.Minute(), 0, 0, now.Location()).Unix()
}

// GetHour 获得当前时间的小时数
func GetHour() int {
	return time.Now().Hour()
}

// GetHourByTimeInt 获得指定时间戳的小时数
func GetHourByTimeInt(t int64) int {
	return time.Unix(t, 0).Hour()
}

// GetHourByTimeLong 获得指定时间戳的小时数
func GetHourByTimeLong(t int64) int {
	return time.Unix(t/1000, 0).Hour()
}

// GetMinute 获得当前时间的分钟数
func GetMinute() int {
	return time.Now().Minute()
}

// GetMinuteByTimeInt 获得指定时间戳的分钟数
func GetMinuteByTimeInt(t int64) int {
	return time.Unix(t, 0).Minute()
}

// GetMinuteByTimeLong 获得指定时间戳的分钟数
func GetMinuteByTimeLong(t int64) int {
	return time.Unix(t/1000, 0).Minute()
}

// GetWeek 获得当前时间的星期几（0-6，0表示星期日）
func GetWeek() time.Weekday {
	return time.Now().Weekday()
}

// GetWeekByTimeLong 获得指定时间戳的星期几
func GetWeekByTimeLong(t int64) time.Weekday {
	return time.Unix(t/1000, 0).Weekday()
}

// GetWeekByTimeInt 获得指定时间戳的星期几
func GetWeekByTimeInt(t int64) time.Weekday {
	return time.Unix(t, 0).Weekday()
}

// GetSecond 获得当前时间的秒数
func GetSecond() int {
	return time.Now().Second()
}

// GetSecondByTimeInt 获得指定时间戳的秒数
func GetSecondByTimeInt(t int64) int {
	return time.Unix(t, 0).Second()
}

// GetSecondByTimeLong 获得指定时间戳的秒数
func GetSecondByTimeLong(t int64) int {
	return time.Unix(t/1000, 0).Second()
}

// GetWeekStr 获得周几字符串
func GetWeekStr() string {
	return GetWeekStrByWeekInt(int(time.Now().Weekday()))
}

// GetMondayTime 获得当前时间戳下的周一的时间戳
func GetMondayTime(ti int64) int64 {
	t := time.Unix(ti, 0)
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	return time.Date(t.Year(), t.Month(), t.Day()-int(weekday)+1, 0, 0, 0, 0, t.Location()).Unix()
}

// StartAndEndWeekTime 获得指定时间戳的周一到周日的时间戳
func StartAndEndWeekTime(ti int64) [2]int64 {
	mondayTime := GetMondayTime(ti)
	return [2]int64{
		mondayTime,
		mondayTime + 7*24*3600 - 1,
	}
}

// GetWeekStrByWeekInt 获得周几字符串
func GetWeekStrByWeekInt(w int) string {
	days := [7]string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}
	if w < 0 || w > 6 {
		return "未知"
	}
	return days[w]
}

// GetBeforeTime 获得指定日期的前一天日期
func GetBeforeTime(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local).AddDate(0, 0, -1)
}

// GetBeforeYear 获得指定日期的前一天的年
func GetBeforeYear(year int, month time.Month, day int) int {
	return GetBeforeTime(year, month, day).Year()
}

// GetBeforeMonth 获得指定日期的前一天的月
func GetBeforeMonth(year int, month time.Month, day int) time.Month {
	return GetBeforeTime(year, month, day).Month()
}

// StartAndEndMonthTime 获得指定时间戳的那个月的起始时间和结束时间
func StartAndEndMonthTime(ti int64) [2]int64 {
	t := time.Unix(ti, 0)
	startOfMonth := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, -1)
	return [2]int64{
		startOfMonth.Unix(),
		endOfMonth.Unix() + 24*3600 - 1,
	}
}

// GetPreviousYearMonthBy 获得指定日期的前一个年月
func GetPreviousYearMonthBy(year int, month time.Month) (int, time.Month) {
	t := time.Date(year, month, 1, 0, 0, 0, 0, time.Local).AddDate(0, -1, 0)
	return t.Year(), t.Month()
}

// GetPreviousYearMonth 获得当前日期的上一个年月
func GetPreviousYearMonth() (int, time.Month) {
	now := time.Now()
	return GetPreviousYearMonthBy(now.Year(), now.Month())
}

// StartAndEndYearTime 获得指定时间戳的那年的起始时间和结束时间
func StartAndEndYearTime(ti int64) [2]int64 {
	t := time.Unix(ti, 0)
	startOfYear := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
	endOfYear := startOfYear.AddDate(1, 0, -1)
	return [2]int64{
		startOfYear.Unix(),
		endOfYear.Unix() + 24*3600 - 1,
	}
}

// GetBeforeDay 获得指定日期的前一天的日
func GetBeforeDay(year int, month time.Month, day int) int {
	return GetBeforeTime(year, month, day).Day()
}

// GetBeforeWeek 获得指定日期前一天的星期几
func GetBeforeWeek(year int, month time.Month, day int) time.Weekday {
	return GetBeforeTime(year, month, day).Weekday()
}

// GetYesterday 获取昨天日
func GetYesterday() int {
	return time.Now().AddDate(0, 0, -1).Day()
}

// GetYesterdayInt 获取昨日时间戳
func GetYesterdayInt() int64 {
	return time.Now().AddDate(0, 0, -1).Unix()
}

// GetStartQuarter 获取每个季度的起始月份
func GetStartQuarter() time.Month {
	month := time.Now().Month()
	return ((month-1)/3)*3 + 1
}

// GetStartQuarterByMonth 获取指定月份所属季度的起始月份
func GetStartQuarterByMonth(month time.Month) time.Month {
	return ((month-1)/3)*3 + 1
}

// GetQuarterByMonth 获取指定月份所属的季度
func GetQuarterByMonth(month time.Month) int {
	return int((month-1)/3) + 1
}

// GetCurrentQuarter 获取当前季度
func GetCurrentQuarter() int {
	return GetQuarterByMonth(time.Now().Month())
}

// IsLeapYear 判断是否为闰年
func IsLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// GetDaysInYear 获取指定年份的天数
func GetDaysInYear(year int) int {
	if IsLeapYear(year) {
		return 366
	}
	return 365
}

// GetDaysInMonth 获取指定年月的天数
func GetDaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// GetFirstDateOfMonth 获取指定年月的第一天
func GetFirstDateOfMonth(year int, month time.Month) time.Time {
	return time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
}

// GetLastDateOfMonth 获取指定年月的最后一天
func GetLastDateOfMonth(year int, month time.Month) time.Time {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local)
}

// GetFirstDateOfWeek 获取指定日期所在周的周一
func GetFirstDateOfWeek(t time.Time) time.Time {
	offset := int(time.Monday - t.Weekday())
	if offset > 0 {
		offset = -6
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).AddDate(0, 0, offset)
}

// GetLastDateOfWeek 获取指定日期所在周的周日
func GetLastDateOfWeek(t time.Time) time.Time {
	return GetFirstDateOfWeek(t).AddDate(0, 0, 6)
}

// GetWeekOfYear 获取指定日期是所在年份的第几周
func GetWeekOfYear(t time.Time) int {
	_, week := t.ISOWeek()
	return week
}

// GetDayOfYear 获取指定日期是所在年份的第几天
func GetDayOfYear(t time.Time) int {
	return t.YearDay()
}

// AddYears 在指定日期上添加年份
func AddYears(t time.Time, years int) time.Time {
	return t.AddDate(years, 0, 0)
}

// AddMonths 在指定日期上添加月份
func AddMonths(t time.Time, months int) time.Time {
	return t.AddDate(0, months, 0)
}

// AddDays 在指定日期上添加天数
func AddDays(t time.Time, days int) time.Time {
	return t.AddDate(0, 0, days)
}

// AddHours 在指定时间上添加小时数
func AddHours(t time.Time, hours int) time.Time {
	return t.Add(time.Duration(hours) * time.Hour)
}

// AddMinutes 在指定时间上添加分钟数
func AddMinutes(t time.Time, minutes int) time.Time {
	return t.Add(time.Duration(minutes) * time.Minute)
}

// AddSeconds 在指定时间上添加秒数
func AddSeconds(t time.Time, seconds int) time.Time {
	return t.Add(time.Duration(seconds) * time.Second)
}

// FormatDuration 格式化时间间隔
func FormatDuration(d time.Duration) string {
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// ParseDuration 解析时间间隔字符串
func ParseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

// GetAgeByBirthday 根据生日计算年龄
func GetAgeByBirthday(birthday time.Time) int {
	now := time.Now()
	years := now.Year() - birthday.Year()
	if now.YearDay() < birthday.YearDay() {
		years--
	}
	return years
}

// IsBetween 判断时间是否在两个时间之间
func IsBetween(t, start, end time.Time) bool {
	return (t.After(start) || t.Equal(start)) && (t.Before(end) || t.Equal(end))
}

// GetTimeZone 获取当前时区
func GetTimeZone() *time.Location {
	return time.Local
}

// SetTimeZone 设置时区
func SetTimeZone(name string) error {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return err
	}
	time.Local = loc
	return nil
}

// GetTimeWithTimeZone 获取指定时区的时间
func GetTimeWithTimeZone(t time.Time, zoneName string) (time.Time, error) {
	loc, err := time.LoadLocation(zoneName)
	if err != nil {
		return time.Time{}, err
	}
	return t.In(loc), nil
}

// GetUTCTime 获取UTC时间
func GetUTCTime() time.Time {
	return time.Now().UTC()
}

// ConvertToUTC 将指定时间转换为UTC时间
func ConvertToUTC(t time.Time) time.Time {
	return t.UTC()
}

// GetTimeFromUnixMilli 从毫秒级时间戳获取时间
func GetTimeFromUnixMilli(millis int64) time.Time {
	return time.Unix(0, millis*int64(time.Millisecond))
}

// GetUnixMilliFromTime 获取时间的毫秒级时间戳
func GetUnixMilliFromTime(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

// GetMicrosecondsFromTime 获取时间的微秒数
func GetMicrosecondsFromTime(t time.Time) int64 {
	return t.UnixNano() / int64(time.Microsecond)
}

// GetNanosecondsFromTime 获取时间的纳秒数
func GetNanosecondsFromTime(t time.Time) int64 {
	return t.UnixNano()
}

// IsWeekend 判断给定时间是否为周末
func IsWeekend(t time.Time) bool {
	return t.Weekday() == time.Saturday || t.Weekday() == time.Sunday
}

// GetNextWeekday 获取下一个工作日
func GetNextWeekday(t time.Time) time.Time {
	for {
		t = t.AddDate(0, 0, 1)
		if !IsWeekend(t) {
			return t
		}
	}
}

// GetPreviousWeekday 获取上一个工作日
func GetPreviousWeekday(t time.Time) time.Time {
	for {
		t = t.AddDate(0, 0, -1)
		if !IsWeekend(t) {
			return t
		}
	}
}

// GetFirstDayOfNextMonth 获取下个月的第一天
func GetFirstDayOfNextMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
}

// GetLastDayOfPreviousMonth 获取上个月的最后一天
func GetLastDayOfPreviousMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()).AddDate(0, 0, -1)
}

// IsExpired 判断给定时间是否已过期
func IsExpired(t time.Time) bool {
	return time.Now().After(t)
}

// GetRemainingDuration 获取距离给定时间的剩余时间
func GetRemainingDuration(t time.Time) time.Duration {
	return t.Sub(time.Now())
}

// FormatTimeRange 格式化时间范围
func FormatTimeRange(start, end time.Time) string {
	if start.Location() != end.Location() {
		end = end.In(start.Location())
	}

	startFormat := "2006-01-02 15:04:05"
	endFormat := "15:04:05"

	if start.Year() != end.Year() {
		endFormat = "2006-01-02 " + endFormat
	} else if start.Month() != end.Month() || start.Day() != end.Day() {
		endFormat = "01-02 " + endFormat
	}

	return start.Format(startFormat) + " - " + end.Format(endFormat)
}

// GetQuarterRange 获取指定年份和季度的时间范围
func GetQuarterRange(year int, quarter int) (time.Time, time.Time) {
	if quarter < 1 || quarter > 4 {
		return time.Time{}, time.Time{}
	}

	firstMonthOfQuarter := time.Month((quarter-1)*3 + 1)
	start := time.Date(year, firstMonthOfQuarter, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 3, -1)
	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 999999999, end.Location())

	return start, end
}
