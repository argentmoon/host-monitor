package monitor

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	. "github.com/argentmoon/host-monitor/log"
	"github.com/argentmoon/host-monitor/monitor/report"
	"github.com/argentmoon/host-monitor/monitor/watch"
)

type Monitor struct {
	// reporter
	r                 report.Reporter
	w                 watch.Watcher
	lastLive          bool
	invalidCheckCount int // 连续无效状态的检测次数，用于多次检测失败时报警用
	ctx               context.Context
	cancel            context.CancelFunc
	failedCount       int64      // 失败次数统计
	successedCount    int64      // 成功次数统计
	statsMutex        sync.Mutex // 统计锁
}

func NewMonitor(r report.Reporter, w watch.Watcher) *Monitor {
	return &Monitor{
		r:        r,
		w:        w,
		lastLive: true,
	}
}

func (m *Monitor) start() {
	m.resetStats()
	GLog.Infof("Monitor.start: name:%v", m.w.Name())
	m.ctx, m.cancel = context.WithCancel(context.Background())
	time.AfterFunc(time.Duration(rand.Int()%30+1)*time.Second, func() {
		go m.run()
	})
}

func (m *Monitor) finish() {
	GLog.Infof("Monitor.finish: name:%v", m.w.Name())
	m.cancel()
}

// Run begin watch and report
func (m *Monitor) run() {
	const reportCount = 3
	timer := time.NewTicker(m.w.FreqInSec())
	for {
		select {
		case <-timer.C:
			live := m.w.IsLive()

			m.statsMutex.Lock()
			lastLive := m.lastLive
			if live {
				m.successedCount++
			} else {
				m.failedCount++
			}
			m.statsMutex.Unlock()

			if live != lastLive {
				// 状态转为有效则进行报告
				if live && m.invalidCheckCount >= reportCount {
					msg := m.getReportMsg(live)
					GLog.Info(msg)
					m.r.Report(msg)
				}

				m.invalidCheckCount = 1
			} else {
				m.invalidCheckCount++

				// 防止溢出
				if m.invalidCheckCount > 99999999 {
					m.invalidCheckCount = reportCount
				}

				// N次检查无效，进行报告
				if reportCount == m.invalidCheckCount && !live {
					msg := m.getReportMsg(live)
					GLog.Info(msg)
					m.r.Report(msg)
				}
			}

			m.statsMutex.Lock()
			m.lastLive = live
			m.statsMutex.Unlock()

			GLog.Debugf("name:%v, live:%v", m.w.Name(), live)

		case <-m.ctx.Done():
			GLog.Infof("name:%v Done", m.w.Name())
			return
		}
	}
}

// 重置统计数
func (m *Monitor) resetStats() {
	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()

	m.successedCount = 0
	m.failedCount = 0
}

// 判断是否需要列入日统计，正常对象不列为统计项目
func (m *Monitor) isNeedToStats() bool {
	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()

	// 主机不能访问，或者失败次数太于0，则需要报告
	return !m.lastLive || m.failedCount > 0
}

// 获得统计信息
func (m *Monitor) getStatsMsg() (msg string) {
	m.statsMutex.Lock()
	defer m.statsMutex.Unlock()

	canUse := "正常访问"
	if !m.lastLive {
		canUse = "无法访问"
	}

	return fmt.Sprintf(
		"主机：%v\n地址：%v\n类型：%v\n成功：%v次\n失败：%v次\n状态：%v\n",
		m.w.Name(),
		m.w.Host(),
		m.w.WatchType(),
		m.successedCount,
		m.failedCount,
		canUse,
	)
}

func (m *Monitor) getReportMsg(live bool) (msg string) {
	canUse := "正常访问"
	if !live {
		canUse = "无法访问"
	}

	timeNow := time.Now()

	return fmt.Sprintf(
		"日期：%v\n时间：%v\n主机：%v\n地址：%v\n类型：%v\n状态：%v\n",
		timeNow.Format("2006-01-02"),
		timeNow.Format("15:04:05"),
		m.w.Name(),
		m.w.Host(),
		m.w.WatchType(),
		canUse,
	)
}
