package repository

import (
	"context"
	"time"

	"license-manager/internal/models"
	pkgcontext "license-manager/pkg/context"
	"license-manager/pkg/i18n"

	"gorm.io/gorm"
)

type dashboardRepository struct {
	db *gorm.DB
}

// NewDashboardRepository 创建仪表盘数据访问层实例
func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{
		db: db,
	}
}

// GetAuthorizationTrendData 获取授权趋势数据
// 使用固定次数的范围查询，避免按日期循环访问数据库。
func (r *dashboardRepository) GetAuthorizationTrendData(ctx context.Context, startDate, endDate time.Time) ([]models.TrendData, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)
	rangeStart := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	rangeEndExclusive := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, endDate.Location()).AddDate(0, 0, 1)

	var initialCount int64
	if err := r.db.WithContext(ctx).Model(&models.AuthorizationCode{}).
		Where("created_at < ?", rangeStart).
		Count(&initialCount).Error; err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	var createdTimes []time.Time
	if err := r.db.WithContext(ctx).Model(&models.AuthorizationCode{}).
		Where("created_at >= ? AND created_at < ?", rangeStart, rangeEndExclusive).
		Pluck("created_at", &createdTimes).Error; err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	var expiredTimes []time.Time
	if err := r.db.WithContext(ctx).Model(&models.AuthorizationCode{}).
		Where("end_date >= ? AND end_date < ?", rangeStart, rangeEndExclusive).
		Pluck("end_date", &expiredTimes).Error; err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	return buildAuthorizationTrendData(rangeStart, rangeEndExclusive, initialCount, createdTimes, expiredTimes), nil
}

func buildAuthorizationTrendData(startDate, endDateExclusive time.Time, initialCount int64, createdTimes, expiredTimes []time.Time) []models.TrendData {
	newByDate := make(map[string]int64)
	expiredByDate := make(map[string]int64)
	loc := startDate.Location()
	for _, createdAt := range createdTimes {
		newByDate[createdAt.In(loc).Format("2006-01-02")]++
	}
	for _, expiredAt := range expiredTimes {
		expiredByDate[expiredAt.In(loc).Format("2006-01-02")]++
	}

	trendData := make([]models.TrendData, 0)
	totalCount := initialCount
	for date := startDate; date.Before(endDateExclusive); date = date.AddDate(0, 0, 1) {
		key := date.Format("2006-01-02")
		totalCount += newByDate[key]
		trendData = append(trendData, models.TrendData{
			Date:                  key,
			TotalAuthorizations:   totalCount,
			NewAuthorizations:     newByDate[key],
			ExpiredAuthorizations: expiredByDate[key],
		})
	}
	return trendData
}

// GetRecentAuthorizations 获取最近授权列表
func (r *dashboardRepository) GetRecentAuthorizations(ctx context.Context, req *models.DashboardRecentAuthorizationsRequest) (*models.DashboardRecentAuthorizationsResponse, error) {
	lang := pkgcontext.GetLanguageFromContext(ctx)

	// 设置默认值
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// 构建查询
	query := r.db.WithContext(ctx).Table("authorization_codes ac").
		Select(`ac.id, ac.code, ac.customer_id, COALESCE(c.customer_name, '') as customer_name,
		        COALESCE(ac.description, '') as description, ac.start_date, ac.end_date, ac.max_activations,
		        COALESCE(l.active_count, 0) as current_activations, 
		        ac.created_at, ac.updated_at, ac.is_locked`).
		Joins("LEFT JOIN customers c ON ac.customer_id = c.id").
		Joins(`LEFT JOIN (
			SELECT authorization_code_id, COUNT(*) as active_count 
			FROM licenses 
			WHERE status = 'active' AND deleted_at IS NULL 
			GROUP BY authorization_code_id
		) l ON ac.id = l.authorization_code_id`).
		Where("1=1").
		Order("ac.created_at DESC")

	// 添加筛选条件
	if req.CustomerID != "" {
		query = query.Where("ac.customer_id = ?", req.CustomerID)
	}

	// 状态筛选（虚字段）
	now := time.Now()
	switch req.Status {
	case "normal":
		query = query.Where("ac.is_locked = ? AND ac.end_date > ?", false, now)
	case "locked":
		query = query.Where("ac.is_locked = ?", true)
	case "expired":
		query = query.Where("ac.end_date <= ?", now)
	}

	// 查询总数 - 使用简化查询避免JOIN复杂性
	var total int64
	countQuery := r.db.WithContext(ctx).Model(&models.AuthorizationCode{}).Where("1=1")

	// 添加相同的筛选条件
	if req.CustomerID != "" {
		countQuery = countQuery.Where("customer_id = ?", req.CustomerID)
	}

	// 状态筛选
	switch req.Status {
	case "normal":
		countQuery = countQuery.Where("is_locked = ? AND end_date > ?", false, now)
	case "locked":
		countQuery = countQuery.Where("is_locked = ?", true)
	case "expired":
		countQuery = countQuery.Where("end_date <= ?", now)
	}

	err := countQuery.Count(&total).Error
	if err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	// 查询列表数据
	type queryResult struct {
		ID                 string    `gorm:"column:id"`
		Code               string    `gorm:"column:code"`
		CustomerID         *string   `gorm:"column:customer_id"`
		CustomerName       string    `gorm:"column:customer_name"`
		Description        string    `gorm:"column:description"`
		StartDate          time.Time `gorm:"column:start_date"`
		EndDate            time.Time `gorm:"column:end_date"`
		MaxActivations     int       `gorm:"column:max_activations"`
		CurrentActivations int       `gorm:"column:current_activations"`
		CreatedAt          time.Time `gorm:"column:created_at"`
		UpdatedAt          time.Time `gorm:"column:updated_at"`
		IsLocked           bool      `gorm:"column:is_locked"`
	}

	var results []queryResult
	err = query.Limit(limit).Find(&results).Error
	if err != nil {
		return nil, i18n.NewI18nError("900004", lang, err.Error())
	}

	// 转换为响应格式
	authorizations := make([]models.RecentAuthorization, 0, len(results))
	for _, result := range results {
		// 计算状态
		status := "normal"
		statusDisplay := i18n.GetEnumMessage("authorization_code_status", "normal", lang)

		if result.IsLocked {
			status = "locked"
			statusDisplay = i18n.GetEnumMessage("authorization_code_status", "locked", lang)
		} else if result.EndDate.Before(now) {
			status = "expired"
			statusDisplay = i18n.GetEnumMessage("authorization_code_status", "expired", lang)
		}

		authorizations = append(authorizations, models.RecentAuthorization{
			ID:                 result.ID,
			Code:               result.Code,
			CustomerID:         result.CustomerID,
			CustomerName:       result.CustomerName,
			Description:        result.Description,
			Status:             status,
			StatusDisplay:      statusDisplay,
			StartDate:          result.StartDate,
			EndDate:            result.EndDate,
			MaxActivations:     result.MaxActivations,
			CurrentActivations: result.CurrentActivations,
			CreatedAt:          result.CreatedAt,
			UpdatedAt:          result.UpdatedAt,
		})
	}

	return &models.DashboardRecentAuthorizationsResponse{
		List:  authorizations,
		Total: total,
	}, nil
}
