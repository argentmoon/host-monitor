package monitor

import (
	"context"
	"database/sql"
	"time"

	. "github.com/argentmoon/host-monitor/log"
	"github.com/argentmoon/host-monitor/monitor/report"
	"github.com/argentmoon/host-monitor/monitor/watch"
	_ "github.com/mxk/go-sqlite/sqlite3"
)

var DB_PATH = "monitor.db"

type MonitorMgr struct {
	mtl    []*Monitor
	ctx    context.Context
	cancel context.CancelFunc
}

func NewMonitorMgr() *MonitorMgr {
	return &MonitorMgr{}
}

func (mm *MonitorMgr) Start() {
	mm.ctx, mm.cancel = context.WithCancel(context.Background())
	go mm.run()
}

func (mm *MonitorMgr) Finish() {
	for _, m := range mm.mtl {
		m.finish()
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
	db, err := sql.Open("sqlite3", DB_PATH)
	if err != nil {
		GLog.Panicf("打开DB：%v错误：%v", DB_PATH, err)
	}

	defer db.Close()

	rows, err := db.Query("select name, type, host, method, dingding_report, freq from job where enable != 0")
	if err != nil {
		GLog.Panicf("读取DB：%v错误：%v", DB_PATH, err)
		return
	}

	defer rows.Close()
	for rows.Next() {
		var name, tp, host, method, dingding_report string
		var freq int

		err = rows.Scan(&name, &tp, &host, &method, &dingding_report, &freq)
		if err != nil {
			GLog.Panicf("读取DB：%v错误：%v", DB_PATH, err)
		}

		// 目前没有其它类型，暂时不拓展这块
		m := NewMonitor(
			report.NewDingdingReporter(dingding_report),
			watch.NewWebsiteWatcher(name, host, method, time.Duration(freq)*time.Second),
		)

		GLog.Warnf("加载Monitor:%v", name)

		mm.mtl = append(mm.mtl, m)
	}

	if rows.Err() != nil {
		GLog.Panicf("读取DB：%v错误：%v", DB_PATH, err)
	}
}
