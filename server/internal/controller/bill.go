package controller

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"expense-log/internal/model"
	"expense-log/internal/service"
	"expense-log/pkg/llm"
	"expense-log/pkg/response"
	"expense-log/pkg/utils"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BillController interface {
	GetTrendStats(c *gin.Context)
	GetCategoryStats(c *gin.Context)
	GetDashboardStats(c *gin.Context)
	GetBillDetail(c *gin.Context)
	UploadImageReceipt(c *gin.Context)
	GetBillList(c *gin.Context)
	UpdateRemark(c *gin.Context)
	UpdateBill(c *gin.Context)
	DeleteBill(c *gin.Context)
	CreateBill(c *gin.Context)
}

type billController struct {
	serv      service.BillService
	db        *gorm.DB
	provider  llm.Provider
	quotaServ service.QuotaService
	llmCfg    model.LLMConfig
}

func NewBillController(serv service.BillService, db *gorm.DB, provider llm.Provider, quotaServ service.QuotaService, llmCfg model.LLMConfig) BillController {
	return &billController{serv: serv, db: db, provider: provider, quotaServ: quotaServ, llmCfg: llmCfg}
}

func (ctrl *billController) getUserID(c *gin.Context) (uuid.UUID, bool) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, false
	}
	userID, ok := userIDValue.(uuid.UUID)
	return userID, ok
}

// getLedgerID 解析或回退获取当前账本ID
func (ctrl *billController) getLedgerID(c *gin.Context, userID uuid.UUID) uuid.UUID {
	ledgerHeader := c.GetHeader("X-Ledger-Id")
	if ledgerHeader != "" {
		if id, err := uuid.Parse(ledgerHeader); err == nil {
			return id
		}
	}
	// 兼容回退：前端未传或旧版时拉取个人默认账本
	var ledger model.Ledger
	if err := ctrl.db.Where("owner_id = ? AND type = ?", userID, model.LedgerTypePersonal).First(&ledger).Error; err == nil {
		return ledger.ID
	}
	return uuid.Nil
}

// GetTrendStats 获取最近6个月的支出趋势
func (ctrl *billController) GetTrendStats(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	ledgerID := ctrl.getLedgerID(c, userID)
	// 如果没有任何账本 (offline/legacy mode)，ledgerID 为 uuid.Nil，Service 层 buildLedgerScope 会处理它为 NULL 匹配


	res, err := ctrl.serv.GetTrendStats(userID, ledgerID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	response.Success(c, res)
}

// GetCategoryStats 获取当月支出分类占比
func (ctrl *billController) GetCategoryStats(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	ledgerID := ctrl.getLedgerID(c, userID)


	res, err := ctrl.serv.GetCategoryStats(userID, ledgerID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	response.Success(c, res)
}

// GetDashboardStats 获取总览数据
func (ctrl *billController) GetDashboardStats(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	ledgerID := ctrl.getLedgerID(c, userID)


	res, err := ctrl.serv.GetDashboardStats(userID, ledgerID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	response.Success(c, res)
}

// GetBillDetail 获取单笔账单详情
func (ctrl *billController) GetBillDetail(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}
	billIDStr := c.Param("id")
	billID, err := uuid.Parse(billIDStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "无效的账单ID")
		return
	}
	bill, err := ctrl.serv.GetBillDetail(userID, billID)
	if err != nil {
		response.Fail(c, http.StatusNotFound, 40400, "账单不存在或无权限访问")
		return
	}
	response.Success(c, bill)
}

// billAnalysisResult AI 返回的结构化账单数据
type billAnalysisResult struct {
	Amount          float64 `json:"amount"`
	Merchant        string  `json:"merchant"`
	TransactionNo   string  `json:"transaction_no"`
	TransactionDate string  `json:"transaction_date"` // 格式: 2006-01-02 15:04:05
	Category        string  `json:"category"`
	Remark          string  `json:"remark"`
}

const analyzePrompt = `你是一个专业的账单分析助手。请仔细分析这张支付截图，提取以下信息并以严格的 JSON 格式返回，不要添加任何其他文字说明：

{
  "amount": 金额(数字，不含货币符号),
  "merchant": "商户名称",
  "transaction_no": "交易单号(如果有)",
  "transaction_date": "交易时间(格式: 2006-01-02 15:04:05)",
  "category": "分类",
  "remark": "简短备注"
}

注意：
- 金额必须是纯数字，如 25.80
- 如果某个字段无法识别，返回空字符串 ""
- 日期格式必须严格遵循 2006-01-02 15:04:05
- 只返回 JSON，不要有其他内容
- 分类从以下选项中选择: 餐饮/交通/购物/娱乐/生活缴费/转账/医疗/退款/其他
- 如果截图显示的是退款、退货、退回、已退款等信息，分类必须填"退款"，备注中注明退款原因`

// UploadImageReceipt iOS 快捷指令专用：接收多张上传图片 → 压缩 → 顺序AI分析 → 存账单
func (ctrl *billController) UploadImageReceipt(c *gin.Context) {
	var readers []io.ReadCloser
	fmt.Printf("\n=== 新的请求接入 ===\nContent-Type: %s\nContent-Length: %d\n", c.Request.Header.Get("Content-Type"), c.Request.ContentLength)

	// 1. 尝试解析 multipart 表单（支持多图，可能会把同名字段传多次，或者命名为 fileX）
	form, errForm := c.MultipartForm()
	if errForm == nil && form != nil {
		fmt.Printf("✅ 解析 multipart 表单成功! 找到了 %d 组表单文件\n", len(form.File))
		for key, files := range form.File {
			fmt.Printf("  - 表单键名 [%s] 共有 %d 个文件\n", key, len(files))
			for j, file := range files {
				fmt.Printf("    -> 提取文件 %d: name=%s, size=%d\n", j+1, file.Filename, file.Size)
				if src, err := file.Open(); err == nil {
					readers = append(readers, src)
				} else {
					fmt.Printf("    ❌ 打开文件出错: %v\n", err)
				}
			}
		}
	} else {
		fmt.Printf("⚠️ 解析 multipart 失败或为 nil: %v, 回退 Raw Body 模式\n", errForm)
		// 2. 降级：iOS 快捷指令单图 Raw Body 模式
		if c.Request.Body != nil {
			bodyBytes, errBody := io.ReadAll(c.Request.Body)
			if errBody == nil && len(bodyBytes) > 0 {
				fmt.Printf("✅ 提取到 Raw Body... (大小: %d 字节)\n", len(bodyBytes))
				readers = append(readers, io.NopCloser(bytes.NewReader(bodyBytes)))
			}
		}
	}

	fmt.Printf("--- 最终提取到的文件流数量: %d ---\n", len(readers))

	if len(readers) == 0 {
		c.String(http.StatusBadRequest, "failed: no file or empty body")
		return
	}

	userIDValue, _ := c.Get("userID")
	userID, _ := userIDValue.(uuid.UUID)

	ledgerID := ctrl.getLedgerID(c, userID)
	if ledgerID == uuid.Nil {
		c.String(http.StatusInternalServerError, "failed: cannot determine ledger")
		return
	}

	// 提前检查并扣除额度
	imageCount := len(readers)
	ctxQuota := context.Background()
	_, errQuota := ctrl.quotaServ.CheckAndConsumeQuota(ctxQuota, userID, imageCount, ctrl.llmCfg.DailyQuota)
	if errQuota != nil {
		response.Fail(c, http.StatusTooManyRequests, 42901, fmt.Sprintf("今日 AI 识图额度不足 (每人每日最多 %d 张)，还需要 %d 张", ctrl.llmCfg.DailyQuota, imageCount))
		return
	}

	successCount := 0

	// 3. 顺序处理每一张图片（避免并发触发大模型 API 限流）
	for i, src := range readers {
		func(r io.ReadCloser, index int) {
			defer r.Close()

			// 压缩图片（最大宽度 800px，JPEG 质量 60%）
			compressed, err := utils.CompressImage(r, 800, 60)
			if err != nil {
				fmt.Printf("❌ 第 %d 张图压缩失败: %v\n", index, err)
				return
			}

			// 保存备份
			filePath := fmt.Sprintf("attachments/%d_%s.jpg", time.Now().Unix(), uuid.New().String()[:8])
			utils.WriteFile(filePath, compressed)

			// base64 编码传给千问分析
			b64 := base64.StdEncoding.EncodeToString(compressed)
			imageDataURL := "data:image/jpeg;base64," + b64

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			rawResult, err := ctrl.provider.AnalyzeImage(ctx, imageDataURL, analyzePrompt)
			if err != nil {
				fmt.Printf("❌ 第 %d 张图 AI 分析调用失败: %v\n", index, err)
				return
			}

			fmt.Printf("=== 第 %d 张图 AI 原生返回结果 ===\n%s\n===================\n", index, rawResult)

			// 解析 AI 返回的 JSON
			var analysis billAnalysisResult
			if err := json.Unmarshal([]byte(rawResult), &analysis); err != nil {
				fmt.Printf("⚠️ 第 %d 张图 JSON 直接解析失败: %v, 尝试提取 JSON 块...\n", index, err)
				if jsonStr := extractJSON(rawResult); jsonStr != "" {
					if err2 := json.Unmarshal([]byte(jsonStr), &analysis); err2 != nil {
						fmt.Printf("❌ 第 %d 张图 JSON 块二次解析也失败了: %v\n", index, err2)
					}
				} else {
					fmt.Printf("❌ 第 %d 张图未找到可提取的 JSON 块\n", index)
				}
			}

			fmt.Printf("✅ 第 %d 张图最终解析出的结构体: %+v\n", index, analysis)

			// 构建账单记录并存库
			cst, _ := time.LoadLocation("Asia/Shanghai")
			transDate, _ := time.ParseInLocation("2006-01-02 15:04:05", analysis.TransactionDate, cst)
			if transDate.IsZero() {
				transDate = time.Now()
			}

			// 🔥 智能退款匹配：如果AI识别为退款，自动找到原始订单并标记
			if analysis.Category == "退款" {
				var originalBill model.Bill
				// 按 金额 + 商户名模糊匹配 + 非退款状态 查找原始订单
				matchQuery := ctrl.db.Where("ledger_id = ? AND amount = ? AND category != '退款'", ledgerID, analysis.Amount)
				if analysis.Merchant != "" {
					matchQuery = matchQuery.Where("merchant LIKE ?", "%"+analysis.Merchant+"%")
				}
				// 优先匹配最近的一笔
				err := matchQuery.Order("created_at DESC").First(&originalBill).Error
				if err == nil {
					// 找到匹配的原始订单 → 直接翻转状态
					refundRemark := fmt.Sprintf("[已退款] %s", analysis.Remark)
					ctrl.db.Model(&originalBill).Updates(map[string]interface{}{
						"category": "退款",
						"remark":   refundRemark,
					})
					successCount++
					ctrl.serv.InvalidateLedgerCache(ledgerID)
					fmt.Printf("🔄 第 %d 张图: 退款自动匹配成功! 已将原始订单 %s (¥%.2f %s) 标记为退款\n",
						index, originalBill.ID.String(), originalBill.Amount, originalBill.Merchant)
					return // 不再创建新记录
				}
				fmt.Printf("⚠️ 第 %d 张图: 退款未匹配到原始订单，将作为独立退款记录入库\n", index)
			}

			bill := &model.Bill{
				UserID:          userID,
				LedgerID:        &ledgerID,
				Amount:          analysis.Amount,
				Merchant:        analysis.Merchant,
				TransactionNo:   analysis.TransactionNo,
				TransactionDate: transDate,
				Category:        analysis.Category,
				Remark:          analysis.Remark,
				Source:          model.BillSourceUpload,
				OriginalFile:    filePath,
				RawContent:      rawResult,
			}

			if err := ctrl.db.Create(bill).Error; err == nil {
				successCount++
				ctrl.serv.InvalidateLedgerCache(ledgerID)
			} else {
				fmt.Printf("⚠️ 第 %d 张图账单入库失败(如属单号重复则正常被拦截): %v\n", index, err)
			}
		}(src, i+1)
	}

	// 为了兼容 iOS 快捷指令原有的判断逻辑，依然返回纯文本的 "success"
	c.String(http.StatusOK, "success")
}

// extractJSON 从可能包含额外文字的 AI 响应中提取 JSON 块
func extractJSON(s string) string {
	start := -1
	depth := 0
	for i, ch := range s {
		if ch == '{' {
			if depth == 0 {
				start = i
			}
			depth++
		} else if ch == '}' {
			depth--
			if depth == 0 && start >= 0 {
				return s[start : i+1]
			}
		}
	}
	return ""
}

// GetBillList 获取账单列表（带分页）
func (ctrl *billController) GetBillList(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if s := c.Query("size"); s != "" {
		fmt.Sscanf(s, "%d", &pageSize)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	keyword := c.Query("keyword")
	category := c.Query("category")
	date := c.Query("date")

	ledgerID := ctrl.getLedgerID(c, userID)
	if ledgerID == uuid.Nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "无法确定操作账本")
		return
	}

	bills, total, err := ctrl.serv.GetBillList(userID, ledgerID, page, pageSize, keyword, category, date)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":  bills,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

type updateRemarkRequest struct {
	Remark string `json:"remark"`
}

func (ctrl *billController) UpdateRemark(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	billIDStr := c.Param("id")
	billID, err := uuid.Parse(billIDStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "无效的账单ID")
		return
	}

	var req updateRemarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "数据格式错误")
		return
	}

	if len(req.Remark) > 255 {
		response.Fail(c, http.StatusBadRequest, 40000, "备注过长")
		return
	}

	if err := ctrl.serv.UpdateRemark(userID, billID, req.Remark); err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "保存失败: "+err.Error())
		return
	}

	response.Success(c, "success")
}

type updateBillRequest struct {
	Amount    float64 `json:"amount"`
	Merchant  string  `json:"merchant"`
	Category  string  `json:"category"`
	Remark    string  `json:"remark"`
	CreatedAt string  `json:"created_at"`
}

func (ctrl *billController) UpdateBill(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}
	billIDStr := c.Param("id")
	billID, err := uuid.Parse(billIDStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "无效的账单ID")
		return
	}
	var req updateBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "数据格式错误")
		return
	}

	cst, _ := time.LoadLocation("Asia/Shanghai")
	t, err := time.ParseInLocation("2006-01-02T15:04", req.CreatedAt, cst)
	if err != nil {
		t, err = time.ParseInLocation("2006-01-02 15:04:05", req.CreatedAt, cst)
		if err != nil {
			t = time.Now()
		}
	}

	dto := service.UpdateBillDTO{
		Amount:    req.Amount,
		Merchant:  html.EscapeString(req.Merchant),
		Category:  html.EscapeString(req.Category),
		Remark:    html.EscapeString(req.Remark),
		CreatedAt: t,
	}

	if err := ctrl.serv.UpdateBill(userID, billID, dto); err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "更新失败: "+err.Error())
		return
	}
	ledgerID := ctrl.getLedgerID(c, userID)
	ctrl.serv.InvalidateLedgerCache(ledgerID)
	response.Success(c, "success")
}

func (ctrl *billController) DeleteBill(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}
	billIDStr := c.Param("id")
	billID, err := uuid.Parse(billIDStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "无效的账单ID")
		return
	}
	if err := ctrl.serv.DeleteBill(userID, billID); err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "删除失败: "+err.Error())
		return
	}
	ledgerID := ctrl.getLedgerID(c, userID)
	ctrl.serv.InvalidateLedgerCache(ledgerID)
	response.Success(c, "success")
}

type createBillRequest struct {
	Amount    float64 `json:"amount" binding:"required"`
	Merchant  string  `json:"merchant" binding:"required"`
	Category  string  `json:"category"`
	Remark    string  `json:"remark"`
	CreatedAt string  `json:"created_at"`
}

func (ctrl *billController) CreateBill(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}
	var req createBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "数据格式错误")
		return
	}

	cst, _ := time.LoadLocation("Asia/Shanghai")
	t, err := time.ParseInLocation("2006-01-02T15:04", req.CreatedAt, cst)
	if err != nil {
		t, err = time.ParseInLocation("2006-01-02 15:04:05", req.CreatedAt, cst)
		if err != nil {
			t = time.Now()
		}
	}

	ledgerID := ctrl.getLedgerID(c, userID)

	var ledgerIDPtr *uuid.UUID
	if ledgerID != uuid.Nil {
		ledgerIDPtr = &ledgerID
	}


	bill := &model.Bill{
		UserID:          userID,
		LedgerID:        ledgerIDPtr,
		Amount:          req.Amount,
		Merchant:        html.EscapeString(req.Merchant),
		Category:        html.EscapeString(req.Category),
		Remark:          html.EscapeString(req.Remark),
		Source:          model.BillSourceManual,
		TransactionDate: t,
	}

	if err := ctrl.db.Create(bill).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "录入失败: "+err.Error())
		return
	}
	ctrl.serv.InvalidateLedgerCache(ledgerID)
	response.Success(c, "success")
}
