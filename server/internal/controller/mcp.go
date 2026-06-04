package controller

import (
	"context"
	"encoding/json"
	"expense-log/internal/model"
	"expense-log/pkg/llm"
	"expense-log/pkg/response"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// SSEClient 代表一个活跃的 MCP 客户端连接通道
type SSEClient struct {
	ID     string
	Send   chan []byte
	UserID uuid.UUID // 该 Key 绑定的系统用户ID
}

type MCPController interface {
	// API Key 申请与管理接口 (管理员 / 普通用户)
	ApplyMCPKey(c *gin.Context)
	ListMyMCPKeys(c *gin.Context)
	AdminListMCPKeys(c *gin.Context)
	AdminApproveMCPKey(c *gin.Context)
	AdminRejectMCPKey(c *gin.Context)
	AdminRevokeMCPKey(c *gin.Context)

	// MCP 标准协议核心端点 (SSE 传输通道)
	HandleSSE(c *gin.Context)
	HandleMessage(c *gin.Context)
}

type mcpController struct {
	db          *gorm.DB
	rdb         *redis.Client
	domain      string
	llmProvider llm.Provider
	clients     map[string]*SSEClient
	mu          sync.RWMutex
}

func NewMCPController(db *gorm.DB, rdb *redis.Client, domain string, llmProvider llm.Provider) MCPController {
	return &mcpController{
		db:          db,
		rdb:         rdb,
		domain:      domain,
		llmProvider: llmProvider,
		clients:     make(map[string]*SSEClient),
	}
}

func (ctrl *mcpController) getUserID(c *gin.Context) (uuid.UUID, bool) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, false
	}
	userID, ok := userIDValue.(uuid.UUID)
	return userID, ok
}

// ==========================================
// 1. API Key 管理系统接口
// ==========================================

type ApplyMCPKeyRequest struct {
	Applicant string `json:"applicant" binding:"required"`
	Purpose   string `json:"purpose" binding:"required"`
}

// ApplyMCPKey 普通用户在线申请 API Key
func (ctrl *mcpController) ApplyMCPKey(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	var req ApplyMCPKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, "参数验证错误")
		return
	}

	key := &model.MCPKey{
		UserID:    userID,
		Applicant: req.Applicant,
		Purpose:   req.Purpose,
		Status:    "pending",
	}

	if err := ctrl.db.Create(key).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "提交申请失败")
		return
	}

	response.Success(c, key)
}

// ListMyMCPKeys 用户查看自己申请的 API Key
func (ctrl *mcpController) ListMyMCPKeys(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	var keys []model.MCPKey
	if err := ctrl.db.Where("user_id = ?", userID).Order("created_at desc").Find(&keys).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "查询失败")
		return
	}

	response.Success(c, keys)
}

// AdminListMCPKeys 管理员获取全部申请列表
func (ctrl *mcpController) AdminListMCPKeys(c *gin.Context) {
	var keys []model.MCPKey
	if err := ctrl.db.Order("created_at desc").Find(&keys).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "查询失败")
		return
	}
	response.Success(c, keys)
}

// AdminApproveMCPKey 管理员同意申请，分发高强度哈希 Key 令牌
func (ctrl *mcpController) AdminApproveMCPKey(c *gin.Context) {
	adminID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	idStr := c.Param("id")
	keyID, err := uuid.Parse(idStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, "无效的 ID 格式")
		return
	}

	var mcpKey model.MCPKey
	if err := ctrl.db.First(&mcpKey, "id = ?", keyID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, 40400, "未找到申请记录")
		return
	}

	if mcpKey.Status != "pending" {
		response.Fail(c, http.StatusBadRequest, 40002, "当前申请单不是待审批状态")
		return
	}

	now := time.Now()
	// 生成随机高强度令牌
	token := "mcp_" + uuid.New().String()

	mcpKey.Key = token
	mcpKey.Status = "approved"
	mcpKey.ApprovedBy = &adminID
	mcpKey.ApprovedAt = &now

	if err := ctrl.db.Save(&mcpKey).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "批准操作失败")
		return
	}

	response.Success(c, gin.H{
		"message": "批准成功，已分发密钥",
		"key":     token,
	})
}

// AdminRejectMCPKey 管理员驳回申请
func (ctrl *mcpController) AdminRejectMCPKey(c *gin.Context) {
	adminID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	idStr := c.Param("id")
	keyID, err := uuid.Parse(idStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, "无效的 ID 格式")
		return
	}

	var mcpKey model.MCPKey
	if err := ctrl.db.First(&mcpKey, "id = ?", keyID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, 40400, "未找到申请记录")
		return
	}

	mcpKey.Status = "rejected"
	mcpKey.ApprovedBy = &adminID
	now := time.Now()
	mcpKey.ApprovedAt = &now

	if err := ctrl.db.Save(&mcpKey).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "操作失败")
		return
	}

	response.Success(c, "已成功驳回该申请")
}

// AdminRevokeMCPKey 管理员撤销/吊销已分发的密钥
func (ctrl *mcpController) AdminRevokeMCPKey(c *gin.Context) {
	adminID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	idStr := c.Param("id")
	keyID, err := uuid.Parse(idStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, "无效的 ID 格式")
		return
	}

	var mcpKey model.MCPKey
	if err := ctrl.db.First(&mcpKey, "id = ?", keyID).Error; err != nil {
		response.Fail(c, http.StatusNotFound, 40400, "未找到记录")
		return
	}

	mcpKey.Status = "revoked"
	mcpKey.ApprovedBy = &adminID
	now := time.Now()
	mcpKey.ApprovedAt = &now

	if err := ctrl.db.Save(&mcpKey).Error; err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "操作失败")
		return
	}

	response.Success(c, "已成功吊销该 API Key")
}

// ==========================================
// 2. MCP (Model Context Protocol) 核心 SSE 与消息端点
// ==========================================

// JSON-RPC 2.0 基础请求与响应定义
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// getMCPAPIKey 统一从 Header (Authorization 或 X-API-Key/X-MCP-Key) 与 Query 中获取 API 密钥
func getMCPAPIKey(c *gin.Context) string {
	// 1. 优先获取 Authorization header (Bearer 模式或直接字符串)
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
		return authHeader
	}

	// 2. 获取自定义 of API Key Header
	if apiKey := c.GetHeader("X-API-Key"); apiKey != "" {
		return apiKey
	}
	if apiKey := c.GetHeader("X-MCP-Key"); apiKey != "" {
		return apiKey
	}

	// 3. 最后回退到 URL query 参数
	return c.Query("api_key")
}

// HandleSSE 处理 MCP SSE 协议通道的初始连接
func (ctrl *mcpController) HandleSSE(c *gin.Context) {
	// 1. 支持 Header 与 Query 参数进行 API Key 安全校验
	apiKey := getMCPAPIKey(c)
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing API Key"})
		return
	}

	var mcpKey model.MCPKey
	if err := ctrl.db.First(&mcpKey, "key = ?", apiKey).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API Key"})
		return
	}

	if mcpKey.Status != "approved" {
		c.JSON(http.StatusForbidden, gin.H{"error": "API Key is not approved or has been revoked"})
		return
	}

	// 2. 建立 SSE 通道
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	clientID := uuid.New().String()
	clientChan := make(chan []byte, 100)

	client := &SSEClient{
		ID:     clientID,
		Send:   clientChan,
		UserID: mcpKey.UserID,
	}

	// 3. 注册客户端连接
	ctrl.mu.Lock()
	ctrl.clients[clientID] = client
	ctrl.mu.Unlock()

	defer func() {
		ctrl.mu.Lock()
		delete(ctrl.clients, clientID)
		ctrl.mu.Unlock()
		close(clientChan)
	}()

	// 4. 发送初始化 endpoint 事件给客户端，指示未来的消息投递地址
	// 按照 MCP SSE 传输规范：event: endpoint, data: 投递 URL
	// 为了安全性考虑，不再在 GET 传递的 URL 中携带敏感的 api_key，后续的 POST 交互由 clientID 进行会话匹配与鉴权
	scheme := "https"
	if c.Request.TLS == nil && c.Request.Header.Get("X-Forwarded-Proto") != "https" {
		scheme = "http"
	}
	messageURL := fmt.Sprintf("%s://%s/api/v1/mcp/message?client_id=%s", scheme, c.Request.Host, clientID)
	c.SSEvent("endpoint", messageURL)
	c.Writer.Flush()

	// 监听连接断开与发送队列
	c.Stream(func(w io.Writer) bool {
		select {
		case msg, ok := <-clientChan:
			if !ok {
				return false
			}
			// 发送消息包：event: message, data: jsonrpc
			c.SSEvent("message", string(msg))
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}

// HandleMessage 处理客户端通过 POST 发送的 JSON-RPC 消息请求
func (ctrl *mcpController) HandleMessage(c *gin.Context) {
	clientID := c.Query("client_id")
	if clientID == "" {
		// ==========================================
		// 支持无状态同步 POST 请求 (MCP over direct POST)
		// ==========================================
		// 1. 校验 API Key 凭证 (支持 Header 与 Query 参数)
		apiKey := getMCPAPIKey(c)

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing client_id or api_key"})
			return
		}

		var mcpKey model.MCPKey
		if err := ctrl.db.First(&mcpKey, "key = ?", apiKey).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API Key"})
			return
		}

		if mcpKey.Status != "approved" {
			c.JSON(http.StatusForbidden, gin.H{"error": "API Key is not approved or has been revoked"})
			return
		}

		var req JSONRPCRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"jsonrpc": "2.0",
				"id":      nil,
				"error":   gin.H{"code": -32700, "message": "Parse error"},
			})
			return
		}

		// 2. 同步执行并直接在 Response Body 中返回标准 JSON-RPC 2.0 结果
		resp := ctrl.handleRPC(mcpKey.UserID, req)
		c.JSON(http.StatusOK, resp)
		return
	}

	// ==========================================
	// 异步 SSE 投递流程
	// ==========================================
	// 1. 查找活跃的 SSE 会话以保证多租户环境安全
	ctrl.mu.RLock()
	client, exists := ctrl.clients[clientID]
	ctrl.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "SSE Client session expired or invalid"})
		return
	}

	var req JSONRPCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON-RPC format"})
		return
	}

	// 2. 根据方法处理 JSON-RPC 请求并回传响应给对应的 SSE 客户端
	go ctrl.processRPC(client, req)

	// 消息送达后直接返回 HTTP 202 Accepted
	c.Status(http.StatusAccepted)
}

// handleRPC 核心方法：同步处理具体的 JSON-RPC 2.0 请求，并保证租户隔离
func (ctrl *mcpController) handleRPC(userID uuid.UUID, req JSONRPCRequest) JSONRPCResponse {
	var resp JSONRPCResponse
	resp.JSONRPC = "2.0"
	resp.ID = req.ID

	switch req.Method {
	case "initialize":
		resp.Result = gin.H{
			"protocolVersion": "2024-11-05",
			"capabilities": gin.H{
				"tools": gin.H{}, // 声明我们支持 Tools
			},
			"serverInfo": gin.H{
				"name":    "expense-log-mcp",
				"version": "1.0.0",
			},
		}

	case "notifications/initialized":
		resp.Result = "ok" // 同步请求可直接返回 ok

	case "tools/list":
		// 返回我们能支持的记账核心工具
		resp.Result = gin.H{
			"tools": []gin.H{
				{
					"name":        "list_bills",
					"description": "获取用户的记账账单明细列表 (可支持分类与搜索)",
					"inputSchema": gin.H{
						"type": "object",
						"properties": gin.H{
							"category": gin.H{"type": "string", "description": "分类筛选(例如：餐饮、交通、购物、娱乐)"},
							"search":   gin.H{"type": "string", "description": "商户名或备注搜索词"},
							"limit":    gin.H{"type": "integer", "description": "返回条数限制，默认10"},
						},
					},
				},
				{
					"name":        "create_bill",
					"description": "快捷写入并记录一笔新的消费账单",
					"inputSchema": gin.H{
						"type": "object",
						"properties": gin.H{
							"amount":           gin.H{"type": "number", "description": "消费金额 (必填)"},
							"merchant":         gin.H{"type": "string", "description": "商户或消费方名称 (必填)"},
							"category":         gin.H{"type": "string", "description": "账单分类，默认'其他'"},
							"remark":           gin.H{"type": "string", "description": "备注说明"},
							"transaction_date": gin.H{"type": "string", "description": "交易日期，格式 YYYY-MM-DD，默认今天"},
						},
						"required": []string{"amount", "merchant"},
					},
				},
				{
					"name":        "get_expense_stats",
					"description": "获取用户指定或当前（年、月、日）整体收支统计及财务汇总",
					"inputSchema": gin.H{
						"type": "object",
						"properties": gin.H{
							"year":  gin.H{"type": "integer", "description": "指定统计年份，例如 2026"},
							"month": gin.H{"type": "integer", "description": "指定统计月份，1-12，例如 6"},
							"day":   gin.H{"type": "integer", "description": "指定统计具体日期，1-31，例如 2"},
						},
					},
				},
				{
					"name":        "semantic_search_bills",
					"description": "通过高维向量 (AI 智能语义) 检索用户的账单明细 (如搜：'买大件电子产品了没', '最近聚餐支出情况')",
					"inputSchema": gin.H{
						"type": "object",
						"properties": gin.H{
							"query": gin.H{"type": "string", "description": "智能语义查询词 (必填)"},
							"limit": gin.H{"type": "integer", "description": "返回条数，默认5"},
						},
						"required": []string{"query"},
					},
				},
				{
					"name":        "get_current_time",
					"description": "获取服务器当前的精确年月日、时间和星期，供大模型明确当前时间以做相对时间分析(例如：今天、上个月等)",
					"inputSchema": gin.H{
						"type": "object",
						"properties": gin.H{},
					},
				},
			},
		}

	case "tools/call":
		// 执行大模型选定的工具
		var callParams struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &callParams); err != nil {
			resp.Error = gin.H{"code": -32602, "message": "Invalid params"}
			break
		}

		result, err := ctrl.executeTool(userID, callParams.Name, callParams.Arguments)
		if err != nil {
			resp.Error = gin.H{"code": -32603, "message": err.Error()}
		} else {
			resp.Result = result
		}

	default:
		resp.Error = gin.H{"code": -32601, "message": "Method not found"}
	}

	return resp
}

// 异步处理具体的 JSON-RPC 消息并通过 SSE 返回
func (ctrl *mcpController) processRPC(client *SSEClient, req JSONRPCRequest) {
	resp := ctrl.handleRPC(client.UserID, req)
	if req.Method == "notifications/initialized" {
		return // 仅仅是通知，不需要回传响应
	}

	// 序列化回执，发送回 SSE 信道
	respBytes, err := json.Marshal(resp)
	if err == nil {
		// 为了保活和防掉线，尝试推送到通道中
		select {
		case client.Send <- respBytes:
		default:
			// 连接可能已异常，忽略或打日志
		}
	}
}

// 核心执行工具：多租户沙箱隔离的数据处理
func (ctrl *mcpController) executeTool(userID uuid.UUID, toolName string, args map[string]interface{}) (interface{}, error) {
	switch toolName {
	case "get_current_time":
		now := time.Now()
		weekdayCn := map[time.Weekday]string{
			time.Sunday:    "星期日",
			time.Monday:    "星期一",
			time.Tuesday:   "星期二",
			time.Wednesday: "星期三",
			time.Thursday:  "星期四",
			time.Friday:    "星期五",
			time.Saturday:  "星期六",
		}
		timeText := fmt.Sprintf("当前服务器时间为: %s (%s)", now.Format("2006-01-02 15:04:05"), weekdayCn[now.Weekday()])
		return gin.H{
			"content": []gin.H{
				{"type": "text", "text": timeText},
			},
		}, nil

	case "list_bills":
		// 根据租户进行安全隔离查询
		tx := ctrl.db.Where("user_id = ?", userID)

		if cat, exists := args["category"].(string); exists && cat != "" {
			tx = tx.Where("category = ?", cat)
		}
		if search, exists := args["search"].(string); exists && search != "" {
			tx = tx.Where("merchant LIKE ? OR remark LIKE ?", "%"+search+"%", "%"+search+"%")
		}

		limit := 10
		if limVal, exists := args["limit"].(float64); exists {
			limit = int(limVal)
		}
		tx = tx.Order("transaction_date desc").Limit(limit)

		var bills []model.Bill
		if err := tx.Find(&bills).Error; err != nil {
			return nil, fmt.Errorf("查询账单失败: %w", err)
		}

		// 格式化输出为大模型易读的文本
		var text string
		if len(bills) == 0 {
			text = "没有找到符合条件的账单明细。"
		} else {
			for idx, b := range bills {
				text += fmt.Sprintf("[%d] 交易时间: %s | 商户: %s | 金额: ¥%.2f | 分类: %s | 备注: %s\n",
					idx+1, b.TransactionDate.Format("2006-01-02"), b.Merchant, b.Amount, b.Category, b.Remark)
			}
		}

		return gin.H{
			"content": []gin.H{
				{"type": "text", "text": text},
			},
		}, nil

	case "create_bill":
		// 记账录单
		amount, ok := args["amount"].(float64)
		if !ok || amount <= 0 {
			return nil, fmt.Errorf("消费金额必须是大于 0 的数字")
		}

		merchant, ok := args["merchant"].(string)
		if !ok || merchant == "" {
			return nil, fmt.Errorf("商户名称不能为空")
		}

		category := "其他"
		if cat, ok := args["category"].(string); ok && cat != "" {
			category = cat
		}

		remark := ""
		if rem, ok := args["remark"].(string); ok {
			remark = rem
		}

		txDate := time.Now()
		if dateStr, ok := args["transaction_date"].(string); ok && dateStr != "" {
			if t, err := time.Parse("2006-01-02", dateStr); err == nil {
				txDate = t
			}
		}

		// 获取个人默认账本，避免成为孤立账单
		var ledger model.Ledger
		var ledgerIDPtr *uuid.UUID
		if err := ctrl.db.Where("owner_id = ? AND type = ?", userID, model.LedgerTypePersonal).First(&ledger).Error; err == nil {
			ledgerIDPtr = &ledger.ID
		}

		bill := &model.Bill{
			UserID:          userID,
			LedgerID:        ledgerIDPtr,
			Amount:          amount,
			Merchant:        merchant,
			Category:        category,
			Remark:          remark,
			TransactionDate: txDate,
			Source:          model.BillSourceManual,
		}

		// 🌟 自动计算新创建账单的语义特征向量 (用于后续的 AI 智能检索)
		desc := fmt.Sprintf("账单明细: 商户是 %s, 消费金额 %.2f 元, 分类属于 %s, 备注为: %s, 交易时间是 %s", 
			bill.Merchant, bill.Amount, bill.Category, bill.Remark, bill.TransactionDate.Format("2006-01-02"))
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		emb, err := ctrl.llmProvider.GetEmbedding(ctx, desc)
		cancel()
		if err == nil {
			bill.Embedding = model.Vector(emb)
		}

		if err := ctrl.db.Create(bill).Error; err != nil {
			return nil, fmt.Errorf("账单录入数据库失败: %w", err)
		}

		// 立即清除 Redis 统计缓存，确保前端能同步看到新记账数据
		if ctrl.rdb != nil && ledgerIDPtr != nil {
			ctrl.rdb.Del(context.Background(),
				fmt.Sprintf("ledger:%s:stats:trend", ledgerIDPtr.String()),
				fmt.Sprintf("ledger:%s:stats:category", ledgerIDPtr.String()),
				fmt.Sprintf("ledger:%s:stats:dashboard", ledgerIDPtr.String()),
			)
		}

		successText := fmt.Sprintf("🎉 记账录单成功！已安全存入您的账本中：\n账单号: %s\n交易商户: %s\n交易金额: ¥%.2f\n所属分类: %s\n交易日期: %s\n备注信息: %s",
			bill.ID.String()[:8], bill.Merchant, bill.Amount, bill.Category, bill.TransactionDate.Format("2006-01-02"), bill.Remark)

		return gin.H{
			"content": []gin.H{
				{"type": "text", "text": successText},
			},
		}, nil

	case "get_expense_stats":
		// 获取收支汇总
		now := time.Now()
		year := now.Year()
		month := int(now.Month())
		day := 0

		// 从 args 提取参数并解析
		if yVal, exists := args["year"]; exists {
			switch v := yVal.(type) {
			case float64:
				year = int(v)
			case string:
				fmt.Sscanf(v, "%d", &year)
			}
		}
		if mVal, exists := args["month"]; exists {
			switch v := mVal.(type) {
			case float64:
				month = int(v)
			case string:
				fmt.Sscanf(v, "%d", &month)
			}
		}
		if dVal, exists := args["day"]; exists {
			switch v := dVal.(type) {
			case float64:
				day = int(v)
			case string:
				fmt.Sscanf(v, "%d", &day)
			}
		}

		var startDate, endDate time.Time
		var periodText string

		if day > 0 {
			// 精确到某一天
			startDate = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
			endDate = startDate.AddDate(0, 0, 1).Add(-time.Second)
			periodText = fmt.Sprintf("%d年%02d月%02d日", year, month, day)
		} else if args["month"] != nil || args["year"] != nil {
			if args["month"] == nil {
				// 仅按年统计
				startDate = time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
				endDate = startDate.AddDate(1, 0, 0).Add(-time.Second)
				periodText = fmt.Sprintf("%d年年度", year)
			} else {
				// 按月统计
				startDate = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
				endDate = startDate.AddDate(0, 1, 0).Add(-time.Second)
				periodText = fmt.Sprintf("%d年%02d月", year, month)
			}
		} else {
			// 默认当前月
			startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
			endDate = startDate.AddDate(0, 1, 0).Add(-time.Second)
			periodText = fmt.Sprintf("本月（%d年%02d月）", now.Year(), now.Month())
		}

		var totalExpense float64
		var totalIncome float64

		// 获取个人默认账本，并将其作为统计的数据隔离边界
		var ledger model.Ledger
		var ledgerIDPtr *uuid.UUID
		if err := ctrl.db.Where("owner_id = ? AND type = ?", userID, model.LedgerTypePersonal).First(&ledger).Error; err == nil {
			ledgerIDPtr = &ledger.ID
		}

		// 餐饮、购物、娱乐等通常代表支出，退款或特定收入代表收入
		// 这里简单归类：Category == "退款" 或者是 "收入" 算收入，其余都计入支出
		var bills []model.Bill
		query := ctrl.db.Model(&model.Bill{})
		if ledgerIDPtr != nil {
			query = query.Where("(ledger_id = ? OR (ledger_id IS NULL AND user_id = ?))", *ledgerIDPtr, userID)
		} else {
			query = query.Where("user_id = ?", userID)
		}

		if err := query.Where("transaction_date BETWEEN ? AND ?", startDate, endDate).Find(&bills).Error; err != nil {
			return nil, fmt.Errorf("拉取统计数据失败: %w", err)
		}

		for _, b := range bills {
			if b.Category == "退款" || b.Category == "收入" {
				totalIncome += b.Amount
			} else {
				totalExpense += b.Amount
			}
		}

		statsText := fmt.Sprintf("📊 %s财务汇总报告：\n- 累计支出: ¥%.2f\n- 累计收入: ¥%.2f\n- 累计记账数: %d 笔\n*注：该数据展示指定时段内记账明细的自动计算。*",
			periodText, totalExpense, totalIncome, len(bills))

		return gin.H{
			"content": []gin.H{
				{"type": "text", "text": statsText},
			},
		}, nil

	case "semantic_search_bills":
		// 🌟 语义检索：基于高维特征向量做余弦相似度检索，带有高强度租户防越权隔离
		query, ok := args["query"].(string)
		if !ok || query == "" {
			return nil, fmt.Errorf("智能语义查询词不能为空")
		}

		limit := 5
		if limVal, exists := args["limit"]; exists {
			switch v := limVal.(type) {
			case float64:
				limit = int(v)
			case int:
				limit = v
			case string:
				fmt.Sscanf(v, "%d", &limit)
			}
		}

		// 1. 调用通义千问 Embedding 接口生成 1536 维查询词特征向量
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		queryVector, err := ctrl.llmProvider.GetEmbedding(ctx, query)
		cancel()
		if err != nil {
			return nil, fmt.Errorf("大模型语义特征提取失败: %w", err)
		}

		// 2. 将向量格式化为 PostgreSQL pgvector 数组格式
		vectorStr, err := model.Vector(queryVector).Value()
		if err != nil {
			return nil, fmt.Errorf("格式化特征数据错误: %w", err)
		}

		// 3. 执行 Postgres 原生余弦相似度极速检索 (带有 user_id 安全隔离)
		var bills []model.Bill
		err = ctrl.db.Where("user_id = ?", userID).
			Order(gorm.Expr("embedding <=> ?", vectorStr)).
			Limit(limit).
			Find(&bills).Error
		if err != nil {
			return nil, fmt.Errorf("特征空间搜索失败: %w", err)
		}

		// 4. 组装文字呈现结果发送给外部 Agent
		var text string
		if len(bills) == 0 {
			text = "未找到任何语义相似或高度关联的账单明细记录。"
		} else {
			text = fmt.Sprintf("🔍 针对语义词 '%s' 最匹配的前 %d 条账单记录：\n", query, len(bills))
			for idx, b := range bills {
				text += fmt.Sprintf("[%d] 匹配度优 | 时间: %s | 商户: %s | 金额: ¥%.2f | 分类: %s | 备注: %s\n",
					idx+1, b.TransactionDate.Format("2006-01-02"), b.Merchant, b.Amount, b.Category, b.Remark)
			}
		}

		return gin.H{
			"content": []gin.H{
				{"type": "text", "text": text},
			},
		}, nil
	}

	return nil, fmt.Errorf("unsupported tool: %s", toolName)
}
