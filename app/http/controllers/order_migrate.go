package controllers

import (
	"fmt"
	"h5/pkg/model"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderMigrate struct {
}

const (
	migrateBatchSize = 2000
	backupTable       = "car_shop_dadi_order_backup_2025"
	sourceTable       = "car_shop_dadi_order"
)

// 迁移进度(内存存储,单实例够用)
type migrateProgress struct {
	mu          sync.Mutex
	Running     bool      `json:"running"`
	TotalTarget int64     `json:"total_target"`
	Migrated    int64     `json:"migrated"`
	StartTime   time.Time `json:"start_time"`
	LastError   string    `json:"last_error,omitempty"`
	Finished    bool      `json:"finished"`
}

var progress = &migrateProgress{}

func (p *migrateProgress) snapshot() migrateProgress {
	p.mu.Lock()
	defer p.mu.Unlock()
	return *p
}

// Start 启动订单归档迁移
// GET /order/migrate-start?cutoff_date=2025-01-01
func (a *OrderMigrate) Start(c *gin.Context) {
	cutoffDate := c.DefaultQuery("cutoff_date", "2025-01-01")

	t, err := time.ParseInLocation("2006-01-02", cutoffDate, time.Local)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "日期格式错误,应为 YYYY-MM-DD"})
		return
	}
	cutoffTimestamp := t.Unix()

	progress.mu.Lock()
	if progress.Running {
		progress.mu.Unlock()
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "迁移任务已在运行中"})
		return
	}
	progress.Running = true
	progress.Migrated = 0
	progress.LastError = ""
	progress.Finished = false
	progress.StartTime = time.Now()
	progress.mu.Unlock()

	tx := model.RDB[model.MASTER].Db

	var total int64
	if err := tx.Table(sourceTable).
		Where("c_time < ?", cutoffTimestamp).
		Count(&total).Error; err != nil {
		progress.mu.Lock()
		progress.Running = false
		progress.LastError = err.Error()
		progress.mu.Unlock()
		c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "统计待迁移数据失败: " + err.Error()})
		return
	}

	progress.mu.Lock()
	progress.TotalTarget = total
	progress.mu.Unlock()

	// 异步执行,236万行预计耗时较久,不能阻塞在请求里
	go a.runMigration(tx, cutoffTimestamp)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "迁移任务已启动,请通过 /order/migrate-progress 查询进度",
		"data": gin.H{"total_target": total},
	})
}

// Progress 查询迁移进度
// GET /order/migrate-progress
func (a *OrderMigrate) Progress(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": progress.snapshot()})
}

func (a *OrderMigrate) runMigration(tx *gorm.DB, cutoffTimestamp int64) {
	defer func() {
		progress.mu.Lock()
		progress.Running = false
		progress.Finished = true
		progress.mu.Unlock()
	}()

	tx.Exec("SET SESSION unique_checks = 0")
	tx.Exec("SET SESSION foreign_key_checks = 0")
	defer func() {
		tx.Exec("SET SESSION unique_checks = 1")
		tx.Exec("SET SESSION foreign_key_checks = 1")
	}()

	for {
		ids, err := a.fetchBatchIDs(tx, cutoffTimestamp, migrateBatchSize)
		if err != nil {
			a.recordError(fmt.Errorf("查询待迁移数据失败: %w", err))
			return
		}

		if len(ids) == 0 {
			break
		}

		if err := a.migrateBatch(tx, ids); err != nil {
			a.recordError(fmt.Errorf("迁移批次失败: %w", err))
			return
		}

		progress.mu.Lock()
		progress.Migrated += int64(len(ids))
		progress.mu.Unlock()

		time.Sleep(300 * time.Millisecond)
	}
}

func (a *OrderMigrate) recordError(err error) {
	progress.mu.Lock()
	progress.LastError = err.Error()
	progress.mu.Unlock()
}

// fetchBatchIDs 查询一批待迁移的订单 ID
func (a *OrderMigrate) fetchBatchIDs(tx *gorm.DB, cutoffTimestamp int64, limit int) ([]int64, error) {
	var ids []int64
	err := tx.Table(sourceTable).
		Select("id").
		Where("c_time < ?", cutoffTimestamp).
		Order("id ASC").
		Limit(limit).
		Pluck("id", &ids).Error
	return ids, err
}

// migrateBatch 在事务内完成一批数据的 INSERT + DELETE
func (a *OrderMigrate) migrateBatch(tx *gorm.DB, ids []int64) error {
	idStrs := make([]string, len(ids))
	for i, id := range ids {
		idStrs[i] = strconv.FormatInt(id, 10)
	}
	idsClause := strings.Join(idStrs, ",")

	return tx.Transaction(func(txn *gorm.DB) error {
		insertSQL := fmt.Sprintf(`
			INSERT INTO %s 
			SELECT * FROM %s 
			WHERE id IN (%s)
		`, backupTable, sourceTable, idsClause)

		if err := txn.Exec(insertSQL).Error; err != nil {
			return fmt.Errorf("插入备份表失败: %w", err)
		}

		deleteSQL := fmt.Sprintf(`
			DELETE FROM %s 
			WHERE id IN (%s)
		`, sourceTable, idsClause)

		if err := txn.Exec(deleteSQL).Error; err != nil {
			return fmt.Errorf("删除原表数据失败: %w", err)
		}

		return nil
	})
}