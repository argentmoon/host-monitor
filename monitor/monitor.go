package monitor

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	. "github.com/argentmoon/host-monitor/log"
	"github.com/argentmoon/host-monitor/monitor/report"
	"github.com/argentmoon/host-monitor/monitor/watch"
)

type Monitor struct {
	// reporter
	r        report.Reporter
	w        watch.Watcher
	lastLive bool
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewMonitor(r report.Reporter, w watch.Watcher) *Monitor {
	return &Monitor{
		r:        r,
		w:        w,
		lastLive: true,
	}
}

func (m *Monitor) start() {
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
	timer := time.NewTicker(m.w.FreqInSec())
	for {
		select {
		case <-timer.C:
			live := m.w.IsLive()
			if live != m.lastLive {
				m.r.Report(m.getReportMsg(m.w, live))
			}

			m.lastLive = live

			GLog.Infof("name:%v, live:%v", m.w.Name(), live)

		case <-m.ctx.Done():
			GLog.Infof("name:%v Done", m.w.Name())
			return
		}
	}
}

func (m *Monitor) getReportMsg(w watch.Watcher, live bool) (msg string) {
	canUse := "正常"
	if !live {
		canUse = "宕机"
	}

	return fmt.Sprintf("主机：%v\n地址：%v\n可用性：%v\n", w.Name(), w.Host(), canUse)
}
