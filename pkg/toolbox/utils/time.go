package utils

import (
	"strconv"
	"time"
)

// TimeRange 时间范围
type TimeRange struct {
	StartTime time.Time
	EndTime   time.Time
}

// NewTimeRange 新建时间范围
func NewTimeRange(startTime, endTime time.Time) *TimeRange {
	return &TimeRange{StartTime: startTime, EndTime: endTime}
}

// NewTimeRangeFromUnix 新建时间范围
func NewTimeRangeFromUnix(startTime, endTime int64) *TimeRange {
	data := &TimeRange{}

	if startTime != 0 {
		data.StartTime = time.Unix(startTime, 0)
	}

	if endTime != 0 {
		data.EndTime = time.Unix(endTime, 0)
	}

	return data
}

// NewTimeRangeFromUnixMilli 新建时间范围
func NewTimeRangeFromUnixMilli(startTime, endTime int64) *TimeRange {
	data := &TimeRange{}

	if startTime != 0 {
		data.StartTime = time.UnixMilli(startTime)
	}

	if endTime != 0 {
		data.EndTime = time.UnixMilli(endTime)
	}

	return data
}

// SetLocation 设置时区
func (t *TimeRange) SetLocation(location *time.Location) *TimeRange {
	t.StartTime = t.StartTime.In(location)
	t.EndTime = t.EndTime.In(location)

	return t
}

// HasStartTime 是否有开始时间
func (t *TimeRange) HasStartTime() bool {
	return t != nil && !t.StartTime.IsZero()
}

// HasEndTime 是否有结束时间
func (t *TimeRange) HasEndTime() bool {
	return t != nil && !t.EndTime.IsZero()
}

// SecondsUntilMidnight 计算当前时间到当天午夜24点的剩余秒数
func SecondsUntilMidnight() int64 {
	now := time.Now()

	// 获取当天午夜时间（第二天的00:00:00）
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)

	// 计算时间差并转换为秒
	seconds := midnight.Sub(now).Seconds()

	return int64(seconds)
}

// GetDateInt 获取日期格式 20060102 int类型
func GetDateInt(t time.Time) int32 {
	date, _ := strconv.Atoi(t.Format("20060102"))
	return int32(date)
}
