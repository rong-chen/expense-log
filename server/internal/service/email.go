package service

import (
	"bytes"
	"crypto/tls"
	"expense-log/internal/model"
	"expense-log/internal/repository"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	gomessage "github.com/emersion/go-message/mail"
	"github.com/google/uuid"
)

type EmailService interface {
	// FetchAllAccounts 遍历所有启用的邮箱账户，拉取新邮件
	FetchAllAccounts() error
	// StartScheduler 启动定时拉取后台任务
	StartScheduler(interval time.Duration)
	// BindAccount 用户绑定邮箱
	BindAccount(userID uuid.UUID, req *model.BindEmailRequest) (*model.EmailAccountResponse, error)
	// GetAccounts 获取用户绑定的邮箱列表
	GetAccounts(userID uuid.UUID) ([]model.EmailAccountResponse, error)
	// DeleteAccount 解绑邮箱
	DeleteAccount(id uuid.UUID, userID uuid.UUID) error
}

type emailService struct {
	repo          repository.EmailRepository
	attachmentDir string
}

func NewEmailService(repo repository.EmailRepository, cfg model.EmailConfig) EmailService {
	attachmentDir := cfg.AttachmentDir
	if attachmentDir == "" {
		attachmentDir = "./attachments"
	}
	if err := os.MkdirAll(attachmentDir, 0755); err != nil {
		log.Printf("[邮件服务] 创建附件目录失败: %v", err)
	}

	return &emailService{
		repo:          repo,
		attachmentDir: attachmentDir,
	}
}

// --- 邮箱账户管理 ---

// BindAccount 用户绑定邮箱
func (s *emailService) BindAccount(userID uuid.UUID, req *model.BindEmailRequest) (*model.EmailAccountResponse, error) {
	port := req.Port
	if port == 0 {
		port = 993
	}
	useTLS := true
	if req.TLS != nil {
		useTLS = *req.TLS
	}
	folder := req.Folder
	if folder == "" {
		folder = "INBOX"
	}

	account := &model.UserEmailAccount{
		UserID:   userID,
		Host:     req.Host,
		Port:     port,
		Username: req.Username,
		Password: req.Password,
		TLS:      useTLS,
		Folder:   folder,
		Enabled:  true,
	}

	// 绑定前先验证能否连接
	if err := s.testConnection(account); err != nil {
		return nil, fmt.Errorf("邮箱连接验证失败: %w", err)
	}

	if err := s.repo.CreateAccount(account); err != nil {
		return nil, err
	}

	return &model.EmailAccountResponse{
		ID:       account.ID,
		Host:     account.Host,
		Username: account.Username,
		Port:     account.Port,
		TLS:      account.TLS,
		Folder:   account.Folder,
		Enabled:  account.Enabled,
	}, nil
}

// GetAccounts 获取用户绑定的邮箱列表
func (s *emailService) GetAccounts(userID uuid.UUID) ([]model.EmailAccountResponse, error) {
	accounts, err := s.repo.GetAccountsByUserID(userID)
	if err != nil {
		return nil, err
	}
	var resp []model.EmailAccountResponse
	for _, a := range accounts {
		resp = append(resp, model.EmailAccountResponse{
			ID:       a.ID,
			Host:     a.Host,
			Username: a.Username,
			Port:     a.Port,
			TLS:      a.TLS,
			Folder:   a.Folder,
			Enabled:  a.Enabled,
		})
	}
	return resp, nil
}

// DeleteAccount 解绑邮箱
func (s *emailService) DeleteAccount(id uuid.UUID, userID uuid.UUID) error {
	return s.repo.DeleteAccount(id, userID)
}

// --- 定时拉取 ---

// StartScheduler 启动定时拉取后台任务
func (s *emailService) StartScheduler(interval time.Duration) {
	go func() {
		log.Printf("[邮件服务] 定时拉取已启动，间隔: %v", interval)

		// 启动时立即执行一次
		if err := s.FetchAllAccounts(); err != nil {
			log.Printf("[邮件服务] 首次拉取失败: %v", err)
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if err := s.FetchAllAccounts(); err != nil {
				log.Printf("[邮件服务] 定时拉取失败: %v", err)
			}
		}
	}()
}

// FetchAllAccounts 遍历所有启用的邮箱账户，逐个拉取
func (s *emailService) FetchAllAccounts() error {
	accounts, err := s.repo.GetAllEnabledAccounts()
	if err != nil {
		return fmt.Errorf("获取邮箱账户列表失败: %w", err)
	}

	if len(accounts) == 0 {
		log.Println("[邮件服务] 暂无绑定的邮箱账户")
		return nil
	}

	log.Printf("[邮件服务] 开始拉取，共 %d 个邮箱账户", len(accounts))

	for _, account := range accounts {
		if err := s.fetchForAccount(&account); err != nil {
			log.Printf("[邮件服务] 拉取失败 [%s]: %v", account.Username, err)
			continue
		}
	}
	return nil
}

// fetchForAccount 拉取单个账户的新邮件
func (s *emailService) fetchForAccount(account *model.UserEmailAccount) error {
	log.Printf("[邮件服务] 拉取邮箱: %s", account.Username)

	// 1. 连接
	client, err := s.connect(account)
	if err != nil {
		return fmt.Errorf("IMAP 连接失败: %w", err)
	}
	defer func() {
		if err := client.Logout().Wait(); err != nil {
			log.Printf("[邮件服务] IMAP 登出失败: %v", err)
		}
		client.Close()
	}()

	// 2. 登录
	if err := client.Login(account.Username, account.Password).Wait(); err != nil {
		return fmt.Errorf("IMAP 登录失败: %w", err)
	}

	// 3. 选择文件夹
	folder := account.Folder
	if folder == "" {
		folder = "INBOX"
	}
	selectData, err := client.Select(folder, nil).Wait()
	if err != nil {
		return fmt.Errorf("选择文件夹 [%s] 失败: %w", folder, err)
	}

	if selectData.NumMessages == 0 {
		log.Printf("[邮件服务] [%s] 邮箱为空", account.Username)
		return nil
	}

	// 4. 搜索未读邮件
	searchCriteria := &imap.SearchCriteria{
		NotFlag: []imap.Flag{imap.FlagSeen},
	}

	searchData, err := client.Search(searchCriteria, nil).Wait()
	if err != nil {
		return fmt.Errorf("搜索未读邮件失败: %w", err)
	}

	uids := searchData.AllUIDs()
	if len(uids) == 0 {
		log.Printf("[邮件服务] [%s] 没有新的未读邮件", account.Username)
		return nil
	}

	log.Printf("[邮件服务] [%s] 发现 %d 封未读邮件", account.Username, len(uids))

	// 5. 获取邮件
	uidSet := imap.UIDSet{}
	for _, uid := range uids {
		uidSet.AddNum(uid)
	}

	fetchOptions := &imap.FetchOptions{
		Envelope:    true,
		BodySection: []*imap.FetchItemBodySection{{}},
	}

	messages, err := client.Fetch(uidSet, fetchOptions).Collect()
	if err != nil {
		return fmt.Errorf("获取邮件失败: %w", err)
	}

	newCount := 0
	for _, msg := range messages {
		if err := s.processMessage(msg, account); err != nil {
			log.Printf("[邮件服务] 处理邮件失败: %v", err)
			continue
		}
		newCount++
	}

	log.Printf("[邮件服务] [%s] 完成，本次新增 %d 封邮件", account.Username, newCount)
	return nil
}

// --- 内部方法 ---

// testConnection 测试邮箱连接是否正常
func (s *emailService) testConnection(account *model.UserEmailAccount) error {
	client, err := s.connect(account)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Login(account.Username, account.Password).Wait(); err != nil {
		return fmt.Errorf("登录失败: %w", err)
	}
	_ = client.Logout().Wait()
	return nil
}

// connect 建立 IMAP 连接
func (s *emailService) connect(account *model.UserEmailAccount) (*imapclient.Client, error) {
	addr := net.JoinHostPort(account.Host, fmt.Sprintf("%d", account.Port))

	if account.TLS {
		return imapclient.DialTLS(addr, &imapclient.Options{
			TLSConfig: &tls.Config{ServerName: account.Host},
		})
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return imapclient.New(conn, nil), nil
}

// processMessage 处理单封邮件
func (s *emailService) processMessage(msg *imapclient.FetchMessageBuffer, account *model.UserEmailAccount) error {
	envelope := msg.Envelope
	if envelope == nil {
		return fmt.Errorf("邮件信封为空")
	}

	messageID := envelope.MessageID
	if messageID == "" {
		messageID = fmt.Sprintf("generated-%s-%s", envelope.Subject, envelope.Date.Format(time.RFC3339))
	}

	// 去重
	exists, err := s.repo.ExistsByMessageID(messageID)
	if err != nil {
		return fmt.Errorf("去重检查失败: %w", err)
	}
	if exists {
		return nil
	}

	fromAddr := ""
	if len(envelope.From) > 0 {
		fromAddr = envelope.From[0].Addr()
	}

	// 解析附件
	hasAttachment := false
	var attachments []model.EmailAttachment

	bodySection := &imap.FetchItemBodySection{}
	bodyData := msg.FindBodySection(bodySection)
	if bodyData != nil {
		parsedAttachments, err := s.extractAttachments(bodyData)
		if err != nil {
			log.Printf("[邮件服务] 解析附件警告 (MessageID=%s): %v", messageID, err)
		} else {
			attachments = parsedAttachments
			hasAttachment = len(attachments) > 0
		}
	}

	// 保存邮件记录（关联用户和账户）
	record := &model.EmailRecord{
		UserID:        account.UserID,
		AccountID:     account.ID,
		MessageID:     messageID,
		Subject:       envelope.Subject,
		FromAddress:   fromAddr,
		Date:          envelope.Date,
		HasAttachment: hasAttachment,
		Processed:     false,
	}

	if err := s.repo.CreateRecord(record); err != nil {
		return fmt.Errorf("保存邮件记录失败: %w", err)
	}

	for i := range attachments {
		attachments[i].EmailID = record.ID
		if err := s.repo.CreateAttachment(&attachments[i]); err != nil {
			log.Printf("[邮件服务] 保存附件记录失败: %v", err)
		}
	}

	log.Printf("[邮件服务] 新邮件: [%s] 来自 %s, %d 个附件", envelope.Subject, fromAddr, len(attachments))
	return nil
}

// extractAttachments 从邮件体中提取并保存附件
func (s *emailService) extractAttachments(body []byte) ([]model.EmailAttachment, error) {
	var attachments []model.EmailAttachment

	reader := bytes.NewReader(body)
	mr, err := gomessage.CreateReader(reader)
	if err != nil {
		return nil, fmt.Errorf("解析邮件 MIME 失败: %w", err)
	}
	defer mr.Close()

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("[邮件服务] 读取 MIME part 警告: %v", err)
			continue
		}

		switch header := part.Header.(type) {
		case *gomessage.AttachmentHeader:
			filename, _ := header.Filename()
			if filename == "" {
				filename = "unknown_attachment"
			}
			if decoded, decErr := decodeFilename(filename); decErr == nil {
				filename = decoded
			}

			contentType, _, _ := header.ContentType()

			ext := filepath.Ext(filename)
			storedName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
			storedPath := filepath.Join(s.attachmentDir, storedName)

			data, err := io.ReadAll(part.Body)
			if err != nil {
				log.Printf("[邮件服务] 读取附件 [%s] 失败: %v", filename, err)
				continue
			}

			if err := os.WriteFile(storedPath, data, 0644); err != nil {
				log.Printf("[邮件服务] 保存附件 [%s] 失败: %v", filename, err)
				continue
			}

			attachments = append(attachments, model.EmailAttachment{
				Filename: filename,
				FilePath: storedPath,
				MimeType: contentType,
				Size:     int64(len(data)),
			})

			log.Printf("[邮件服务] 附件已保存: %s (%s, %d bytes)", filename, contentType, len(data))
		}
	}

	return attachments, nil
}

func decodeFilename(encoded string) (string, error) {
	dec := new(mime.WordDecoder)
	return dec.DecodeHeader(encoded)
}
