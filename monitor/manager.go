package monitor

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	. "github.com/argentmoon/host-monitor/log"
	"github.com/argentmoon/host-monitor/monitor/report"
	"github.com/argentmoon/host-monitor/monitor/watch"
	"github.com/go-ini/ini"
	_ "github.com/mxk/go-sqlite/sqlite3"
)

var dbPath = "monitor.db"
var statsDingDing = ""
var statsReportTime = -1

func init() {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		GLog.Fatal("未找到config.ini")
	}

	statsDingDing = cfg.Section("").Key("stats_dingding").String()
	statsReportTimeStr := cfg.Section("").Key("stats_report_time").String()
	if len(statsReportTimeStr) >= 0 {
		var err error
		statsReportTime, err = strconv.Atoi(statsReportTimeStr)
		if err != nil || statsReportTime < 0 || statsReportTime >= 24 {
			GLog.Fatalf("未能启动安全日报任务，statsReportTime格式错误：%v", statsReportTimeStr)
		}
	}
}

// MonitorMgr 监测器管理
type MonitorMgr struct {
	mtl    []*Monitor
	ctx    context.Context
	cancel context.CancelFunc
}

// NewMonitorMgr
func NewMonitorMgr() *MonitorMgr {
	return &MonitorMgr{}
}

func (mm *MonitorMgr) Start() {
	mm.ctx, mm.cancel = context.WithCancel(context.Background())
	go mm.run()
	mm.statsJobBegin()
}

func (mm *MonitorMgr) Finish() {
	for _, m := range mm.mtl {
		m.finish()
		m.resetStats()
	}
}

func (mm *MonitorMgr) run() {
	for _, m := range mm.mtl {
		m.start()
	}
}

func (mm *MonitorMgr) CleanAll() {
	mm.Finish()
	mm.mtl = []*Monitor{}
}

func (mm *MonitorMgr) Load() {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		GLog.Panicf("打开DB：%v错误：%v", dbPath, err)
	}

	defer db.Close()

	rows, err := db.Query("select name, type, host, method, dingding_report, freq from job where enable != 0")
	if err != nil {
		GLog.Panicf("读取DB：%v错误：%v", dbPath, err)
		return
	}

	defer rows.Close()
	for rows.Next() {
		var name, tp, host, method, dingdingReport sql.NullString
		var freq sql.NullInt64

		err = rows.Scan(&name, &tp, &host, &method, &dingdingReport, &freq)
		if err != nil {
			GLog.Panicf("读取DB：%v错误：%v", dbPath, err)
		}

		var m *Monitor
		switch strings.ToLower(tp.String) {
		case "web":
			m = NewMonitor(
				report.NewDingdingReporter(dingdingReport.String),
				watch.NewWebsiteWatcher(name.String, host.String, method.String, time.Duration(freq.Int64)*time.Second),
			)

			break
		case "port":
			m = NewMonitor(
				report.NewDingdingReporter(dingdingReport.String),
				watch.NewPortWatcher(name.String, host.String, time.Duration(freq.Int64)*time.Second),
			)

			break
		case "ping":
			m = NewMonitor(
				report.NewDingdingReporter(dingdingReport.String),
				watch.NewPingWatcher(name.String, host.String, time.Duration(freq.Int64)*time.Second),
			)

			break
		default:
			GLog.Warn("遇到错误的监测类型：", name.String, tp.String, host.String, method.String)
			continue
		}

		mm.mtl = append(mm.mtl, m)
		GLog.Info("加载Monitor:", name.String, tp.String, host.String, method.String)
	}

	if rows.Err() != nil {
		GLog.Panicf("读取DB：%v错误：%v", dbPath, err)
	}
}

// 安全统计
func (mm *MonitorMgr) statsJob() {
	if len(mm.mtl) <= 0 {
		return
	}

	// 获取统计信息
	statsMsgs := []string{"监测日报："}
	for _, m := range mm.mtl {
		statsMsgs = append(statsMsgs, m.getStatsMsg())

		// 重置
		m.resetStats()
	}

	msg := strings.Join(statsMsgs, "\n\n")
	GLog.Println(msg)

	// 发送信息
	postMsg := fmt.Sprintf("{ \"msgtype\": \"text\", \"text\": {\"content\": \"%s\"}}", msg)
	_, err := http.Post(statsDingDing, "application/json; charset=utf-8", strings.NewReader(postMsg))
	if err != nil {
		GLog.Error(err)
	}

	// 重新计算下一天的时间，避免误差放大
	now := time.Now()
	nexttime := now
	if now.Hour() >= statsReportTime {
		// 由于误差，超出时点，先加24小时，再校正时点
		nextday := now.Add(time.Hour * 24)
		nexttime = time.Date(
			nextday.Year(),
			nextday.Month(),
			nextday.Day(),
			statsReportTime,
			0,
			0,
			0,
			time.Local,
		)
	} else {
		// 由于误差，未超出时点，先校正时点，再加24小时
		nexttime = time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			statsReportTime,
			0,
			0,
			0,
			time.Local,
		).Add(time.Hour * 24)
	}

	GLog.Println("安全日报任务下一次任务开始时间：", nexttime)

	duration := nexttime.Sub(now)

	// 计划任务
	time.AfterFunc(duration, func() { //非阻塞
		mm.statsJob()
	})
}

func (mm *MonitorMgr) statsJobBegin() {
	if statsReportTime == -1 || len(statsDingDing) <= 0 {
		return
	}

	now := time.Now()
	nexttime := now
	if now.Hour() >= statsReportTime {
		// 第二天才开始
		nextday := now.Add(time.Hour * 24)
		nexttime = time.Date(
			nextday.Year(),
			nextday.Month(),
			nextday.Day(),
			statsReportTime,
			0,
			0,
			0,
			time.Local,
		)
	} else {
		// 今天开始
		nexttime = time.Date(
			now.Year(),
			now.Month(),
			now.Day(),
			statsReportTime,
			0,
			0,
			0,
			time.Local,
		)
	}

	duration := nexttime.Sub(now)

	// 避免临近的误差
	if duration <= 0 {
		duration = 1
	}

	GLog.Println("启动安全日报任务，下一次任务开始时间：", nexttime)

	time.AfterFunc(duration, func() { //非阻塞
		// 第一次执行
		mm.statsJob()
	})
}
