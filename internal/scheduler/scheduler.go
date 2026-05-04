// Package scheduler 提供定时和手动触发能力。
package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Job 表示一个可被调度执行的任务。
type Job func(context.Context) error

// Schedule 定义下一次执行时间的计算规则。
type Schedule interface {
	Next(after time.Time) time.Time
}

// Scheduler 根据调度规则持续执行任务。
type Scheduler struct {
	schedule Schedule
	job      Job
	now      func() time.Time
}

// New 创建固定间隔调度器。
func New(interval time.Duration, job Job) *Scheduler {
	return &Scheduler{
		schedule: intervalSchedule{interval: interval},
		job:      job,
		now:      time.Now,
	}
}

// NewWithSchedule 使用显式调度规则创建调度器。
func NewWithSchedule(schedule Schedule, job Job) *Scheduler {
	return &Scheduler{
		schedule: schedule,
		job:      job,
		now:      time.Now,
	}
}

// NewCron 创建基于 cron 表达式的调度器。
func NewCron(spec string, job Job) (*Scheduler, error) {
	cron, err := ParseCron(spec)
	if err != nil {
		return nil, err
	}

	return NewWithSchedule(cron, job), nil
}

// RunOnce 手动触发一次任务。
func (s *Scheduler) RunOnce(ctx context.Context) error {
	return s.job(ctx)
}

// Start 按调度规则持续执行任务，直到上下文取消。
func (s *Scheduler) Start(ctx context.Context, onError func(error)) {
	for {
		next := s.schedule.Next(s.now())
		wait := time.Until(next)
		if wait < 0 {
			wait = 0
		}

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			if err := s.job(ctx); err != nil && onError != nil {
				onError(err)
			}
		}
	}
}

type intervalSchedule struct {
	interval time.Duration
}

func (s intervalSchedule) Next(after time.Time) time.Time {
	return after.Add(s.interval)
}

// CronSchedule 表示 5 段式 cron 调度规则。
type CronSchedule struct {
	minute cronField
	hour   cronField
	day    cronField
	month  cronField
	week   cronField
}

// ParseCron 解析标准 5 段式 cron 表达式：分 时 日 月 周。
func ParseCron(spec string) (CronSchedule, error) {
	parts := strings.Fields(strings.TrimSpace(spec))
	if len(parts) != 5 {
		return CronSchedule{}, fmt.Errorf("cron 表达式必须包含 5 段")
	}

	minute, err := parseCronField(parts[0], 0, 59)
	if err != nil {
		return CronSchedule{}, fmt.Errorf("解析分钟失败：%w", err)
	}
	hour, err := parseCronField(parts[1], 0, 23)
	if err != nil {
		return CronSchedule{}, fmt.Errorf("解析小时失败：%w", err)
	}
	day, err := parseCronField(parts[2], 1, 31)
	if err != nil {
		return CronSchedule{}, fmt.Errorf("解析日期失败：%w", err)
	}
	month, err := parseCronField(parts[3], 1, 12)
	if err != nil {
		return CronSchedule{}, fmt.Errorf("解析月份失败：%w", err)
	}
	week, err := parseCronField(parts[4], 0, 7)
	if err != nil {
		return CronSchedule{}, fmt.Errorf("解析星期失败：%w", err)
	}
	if week.allowed != nil {
		if _, ok := week.allowed[7]; ok {
			delete(week.allowed, 7)
			week.allowed[0] = struct{}{}
		}
	}

	return CronSchedule{
		minute: minute,
		hour:   hour,
		day:    day,
		month:  month,
		week:   week,
	}, nil
}

// Next 返回给定时间之后最近一次满足 cron 规则的触发时间。
func (s CronSchedule) Next(after time.Time) time.Time {
	current := after.Truncate(time.Minute).Add(time.Minute)
	limit := current.AddDate(5, 0, 0)

	for !current.After(limit) {
		if s.matches(current) {
			return current
		}
		current = current.Add(time.Minute)
	}

	return limit
}

func (s CronSchedule) matches(value time.Time) bool {
	if !s.minute.matches(value.Minute()) || !s.hour.matches(value.Hour()) || !s.month.matches(int(value.Month())) {
		return false
	}

	dayMatch := s.day.matches(value.Day())
	weekMatch := s.week.matches(int(value.Weekday()))
	switch {
	case s.day.all && s.week.all:
		return true
	case s.day.all:
		return weekMatch
	case s.week.all:
		return dayMatch
	default:
		return dayMatch || weekMatch
	}
}

type cronField struct {
	all     bool
	allowed map[int]struct{}
}

func (f cronField) matches(value int) bool {
	if f.all {
		return true
	}
	_, ok := f.allowed[value]
	return ok
}

func parseCronField(spec string, minValue, maxValue int) (cronField, error) {
	if spec == "*" {
		return cronField{all: true}, nil
	}

	field := cronField{allowed: make(map[int]struct{})}
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			return cronField{}, fmt.Errorf("字段不能为空")
		}
		if err := addCronPart(field.allowed, part, minValue, maxValue); err != nil {
			return cronField{}, err
		}
	}

	return field, nil
}

func addCronPart(allowed map[int]struct{}, spec string, minValue, maxValue int) error {
	base := spec
	step := 1
	if strings.Contains(spec, "/") {
		parts := strings.Split(spec, "/")
		if len(parts) != 2 {
			return fmt.Errorf("无效步长表达式：%s", spec)
		}
		base = parts[0]
		parsedStep, err := strconv.Atoi(parts[1])
		if err != nil || parsedStep <= 0 {
			return fmt.Errorf("无效步长：%s", parts[1])
		}
		step = parsedStep
	}

	start := minValue
	end := maxValue
	switch {
	case base == "*" || base == "":
	case strings.Contains(base, "-"):
		rangeParts := strings.Split(base, "-")
		if len(rangeParts) != 2 {
			return fmt.Errorf("无效范围：%s", base)
		}
		parsedStart, err := strconv.Atoi(rangeParts[0])
		if err != nil {
			return fmt.Errorf("无效范围起点：%s", rangeParts[0])
		}
		parsedEnd, err := strconv.Atoi(rangeParts[1])
		if err != nil {
			return fmt.Errorf("无效范围终点：%s", rangeParts[1])
		}
		start = parsedStart
		end = parsedEnd
	default:
		value, err := strconv.Atoi(base)
		if err != nil {
			return fmt.Errorf("无效取值：%s", base)
		}
		start = value
		end = value
	}

	if start < minValue || end > maxValue || start > end {
		return fmt.Errorf("取值超出范围：%s", spec)
	}

	for value := start; value <= end; value += step {
		allowed[value] = struct{}{}
	}

	return nil
}
