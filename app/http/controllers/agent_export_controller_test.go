package controllers

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func TestParseMonthRange(t *testing.T) {
	start, end, err := parseMonthRange("2026-01", "2026-03")
	if err != nil {
		t.Fatalf("parseMonthRange returned error: %v", err)
	}
	if got := start.Format("2006-01-02"); got != "2026-01-01" {
		t.Fatalf("unexpected start date: %s", got)
	}
	if got := end.Format("2006-01-02"); got != "2026-04-01" {
		t.Fatalf("unexpected exclusive end date: %s", got)
	}

	if _, _, err := parseMonthRange("2026-03", "2026-01"); err == nil {
		t.Fatal("expected reversed month range to fail")
	}
	if _, _, err := parseMonthRange("2026/01", ""); err == nil {
		t.Fatal("expected invalid month format to fail")
	}
}

func TestWriteAgentStatisticsWorkbookCreatesMonthlySheets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	data := []diyAgentMonthlyStat{
		{Month: "2026-01", WorkNum: "A001", Name: "张三", TotalNum: 3},
		{Month: "2026-01", WorkNum: "A002", Name: "李四", TotalNum: 2},
		{Month: "2026-02", WorkNum: "A001", Name: "张三", TotalNum: 1},
	}

	if err := writeAgentStatisticsWorkbook(ctx, data); err != nil {
		t.Fatalf("writeAgentStatisticsWorkbook returned error: %v", err)
	}

	f, err := excelize.OpenReader(bytes.NewReader(recorder.Body.Bytes()))
	if err != nil {
		t.Fatalf("open generated workbook: %v", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) != 2 || sheets[0] != "2026-01" || sheets[1] != "2026-02" {
		t.Fatalf("unexpected sheets: %#v", sheets)
	}
	rows, err := f.GetRows("2026-01")
	if err != nil {
		t.Fatalf("read January sheet: %v", err)
	}
	if len(rows) != 3 || rows[1][1] != "A001" || rows[2][1] != "A002" {
		t.Fatalf("unexpected January rows: %#v", rows)
	}
	if len(rows[0]) != 8 || rows[0][2] != "代理人姓名" || rows[0][7] != "上传图片数" {
		t.Fatalf("unexpected headers: %#v", rows[0])
	}
}

func TestMergeAgentMonthlyStatisticsIncludesAgentsWithoutCustomers(t *testing.T) {
	agents := []diyAgent{
		{WorkNum: "A001", Name: "张三"},
		{WorkNum: "A002", Name: "李四"},
	}
	statistics := []diyAgentMonthlyStat{
		{Month: "2026-01", WorkNum: "A001", TotalNum: 3, OrderNum: 1},
	}

	result := mergeAgentMonthlyStatistics(agents, statistics, []string{"2026-01"})
	if len(result) != 2 {
		t.Fatalf("expected two agents, got %#v", result)
	}
	if result[1].WorkNum != "A002" || result[1].TotalNum != 0 || result[1].OrderNum != 0 {
		t.Fatalf("agent without customers was not exported with zero values: %#v", result[1])
	}
}

func TestAgentStatisticMonthsIncludesEmptyRequestedMonths(t *testing.T) {
	start, end, err := parseMonthRange("2026-01", "2026-03")
	if err != nil {
		t.Fatal(err)
	}
	months := agentStatisticMonths(nil, start, end)
	if len(months) != 3 || months[0] != "2026-01" || months[2] != "2026-03" {
		t.Fatalf("unexpected months: %#v", months)
	}
}
