package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	tauth "github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	xproxy "golang.org/x/net/proxy"
)

type mtStorage struct {
	data []byte
}

func (s *mtStorage) LoadSession(_ context.Context) ([]byte, error) {
	if len(s.data) == 0 {
		return nil, session.ErrNotFound
	}
	return s.data, nil
}

func (s *mtStorage) StoreSession(_ context.Context, data []byte) error {
	s.data = append([]byte(nil), data...)
	return nil
}

func MTProtoSendCode(apiID int, apiHash, proxyURL, phone string) error {
	apiHash = strings.TrimSpace(apiHash)
	proxyURL = strings.TrimSpace(proxyURL)
	phone = strings.TrimSpace(phone)
	if apiID <= 0 || apiHash == "" || phone == "" {
		return fmt.Errorf("api_id、api_hash、phone 不能为空")
	}

	st := &mtStorage{}
	client, err := newMTProtoClient(apiID, apiHash, proxyURL, st)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	var codeHash string
	if err := client.Run(ctx, func(ctx context.Context) error {
		authClient := tauth.NewClient(client.API(), rand.Reader, apiID, apiHash)
		sentCode, err := authClient.SendCode(ctx, phone, tauth.SendCodeOptions{})
		if err != nil {
			return err
		}
		s, ok := sentCode.(*tg.AuthSentCode)
		if !ok {
			return fmt.Errorf("发送验证码失败：返回类型异常")
		}
		codeHash = s.PhoneCodeHash
		return nil
	}); err != nil {
		return err
	}

	state, _ := getOrCreateAuthState()
	state.Phone = phone
	state.PhoneCodeHash = codeHash
	state.TempSession = base64.StdEncoding.EncodeToString(st.data)
	state.NeedPassword = false
	return database.DB().Save(&state).Error
}

func MTProtoSignIn(code string) (needPassword bool, err error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return false, fmt.Errorf("验证码不能为空")
	}
	state, err := getOrCreateAuthState()
	if err != nil {
		return false, err
	}
	if state.Phone == "" || state.PhoneCodeHash == "" {
		return false, fmt.Errorf("请先发送验证码")
	}

	cfg, err := getSystemConfig()
	if err != nil {
		return false, err
	}
	if cfg.TgAPIID <= 0 || strings.TrimSpace(cfg.TgAPIHash) == "" {
		return false, fmt.Errorf("请先在系统配置填写 tg_api_id / tg_api_hash")
	}

	st := &mtStorage{}
	if state.TempSession != "" {
		if buf, decErr := base64.StdEncoding.DecodeString(state.TempSession); decErr == nil {
			st.data = buf
		}
	}
	client, err := newMTProtoClient(cfg.TgAPIID, strings.TrimSpace(cfg.TgAPIHash), strings.TrimSpace(cfg.TgProxyURL), st)
	if err != nil {
		return false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	runErr := client.Run(ctx, func(ctx context.Context) error {
		authClient := tauth.NewClient(client.API(), rand.Reader, cfg.TgAPIID, strings.TrimSpace(cfg.TgAPIHash))
		_, signErr := authClient.SignIn(ctx, state.Phone, code, state.PhoneCodeHash)
		return signErr
	})

	state.TempSession = base64.StdEncoding.EncodeToString(st.data)
	if runErr != nil {
		if errorsIsPasswordNeeded(runErr) {
			state.NeedPassword = true
			_ = database.DB().Save(&state).Error
			return true, nil
		}
		_ = database.DB().Save(&state).Error
		return false, runErr
	}

	cfg.TgSession = base64.StdEncoding.EncodeToString(st.data)
	state.NeedPassword = false
	state.PhoneCodeHash = ""
	if err := database.DB().Save(&cfg).Error; err != nil {
		return false, err
	}
	if err := database.DB().Save(&state).Error; err != nil {
		return false, err
	}
	return false, nil
}

func MTProtoCheckPassword(password string) error {
	password = strings.TrimSpace(password)
	if password == "" {
		return fmt.Errorf("2FA 密码不能为空")
	}
	state, err := getOrCreateAuthState()
	if err != nil {
		return err
	}
	cfg, err := getSystemConfig()
	if err != nil {
		return err
	}
	if cfg.TgAPIID <= 0 || strings.TrimSpace(cfg.TgAPIHash) == "" {
		return fmt.Errorf("请先在系统配置填写 tg_api_id / tg_api_hash")
	}
	if state.TempSession == "" {
		return fmt.Errorf("请先验证码登录")
	}
	tempBytes, err := base64.StdEncoding.DecodeString(state.TempSession)
	if err != nil {
		return fmt.Errorf("临时会话无效")
	}

	st := &mtStorage{data: tempBytes}
	client, err := newMTProtoClient(cfg.TgAPIID, strings.TrimSpace(cfg.TgAPIHash), strings.TrimSpace(cfg.TgProxyURL), st)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	if err := client.Run(ctx, func(ctx context.Context) error {
		authClient := tauth.NewClient(client.API(), rand.Reader, cfg.TgAPIID, strings.TrimSpace(cfg.TgAPIHash))
		_, passErr := authClient.Password(ctx, password)
		return passErr
	}); err != nil {
		return err
	}

	cfg.TgSession = base64.StdEncoding.EncodeToString(st.data)
	state.NeedPassword = false
	state.PhoneCodeHash = ""
	if err := database.DB().Save(&cfg).Error; err != nil {
		return err
	}
	return database.DB().Save(&state).Error
}

func MTProtoSessionStatus() (map[string]interface{}, error) {
	cfg, err := getSystemConfig()
	if err != nil {
		return nil, err
	}
	state, _ := getOrCreateAuthState()
	return map[string]interface{}{
		"has_api":       cfg.TgAPIID > 0 && strings.TrimSpace(cfg.TgAPIHash) != "",
		"has_session":   strings.TrimSpace(cfg.TgSession) != "",
		"need_password": state.NeedPassword,
		"phone":         state.Phone,
	}, nil
}

type MTProtoConfig struct {
	APIID   int
	APIHash string
	ProxyURL string
}

func GetMTProtoConfigFromDB() (MTProtoConfig, error) {
	cfg, err := getSystemConfig()
	if err != nil {
		return MTProtoConfig{}, err
	}
	return MTProtoConfig{
		APIID:   cfg.TgAPIID,
		APIHash: strings.TrimSpace(cfg.TgAPIHash),
		ProxyURL: strings.TrimSpace(cfg.TgProxyURL),
	}, nil
}

func newMTProtoClient(apiID int, apiHash, proxyURL string, st *mtStorage) (*telegram.Client, error) {
	opts := telegram.Options{SessionStorage: st}
	proxyURL = strings.TrimSpace(proxyURL)
	if proxyURL == "" {
		return telegram.NewClient(apiID, apiHash, opts), nil
	}

	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("tg_proxy_url 格式错误")
	}
	switch strings.ToLower(parsed.Scheme) {
	case "socks5", "socks5h":
		dialer, err := xproxy.FromURL(parsed, &net.Dialer{Timeout: 15 * time.Second})
		if err != nil {
			return nil, fmt.Errorf("tg_proxy_url 无效: %w", err)
		}
		ctxDialer, ok := dialer.(xproxy.ContextDialer)
		if !ok {
			return nil, fmt.Errorf("socks5 代理不支持上下文拨号")
		}
		opts.Resolver = dcs.Plain(dcs.PlainOptions{Dial: ctxDialer.DialContext})
		return telegram.NewClient(apiID, apiHash, opts), nil
	case "http", "https":
		return nil, fmt.Errorf("MTProto 仅支持 socks5 代理，请将 tg_proxy_url 设置为 socks5://")
	default:
		return nil, fmt.Errorf("tg_proxy_url 仅支持 socks5://（MTProto）")
	}
}

func getSystemConfig() (model.SystemConfig, error) {
	var cfg model.SystemConfig
	if err := database.DB().Order("id ASC").First(&cfg).Error; err != nil {
		return cfg, err
	}
	return cfg, nil
}

func getOrCreateAuthState() (model.TelegramAuthState, error) {
	var state model.TelegramAuthState
	db := database.DB()
	if err := db.Order("id ASC").First(&state).Error; err == nil {
		return state, nil
	}
	state = model.TelegramAuthState{}
	if err := db.Create(&state).Error; err != nil {
		return state, err
	}
	return state, nil
}

func errorsIsPasswordNeeded(err error) bool {
	if err == nil {
		return false
	}
	if err == tauth.ErrPasswordAuthNeeded {
		return true
	}
	return strings.Contains(strings.ToUpper(err.Error()), "SESSION_PASSWORD_NEEDED")
}

