package handler

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/oauth"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
	"github.com/tidwall/gjson"
)

const (
	dingtalkOAuthCookiePath        = "/api/v1/auth/oauth/dingtalk"
	dingtalkOAuthStateCookieName   = "dingtalk_oauth_state"
	dingtalkOAuthRedirectCookie    = "dingtalk_oauth_redirect"
	dingtalkOAuthCookieMaxAgeSec   = 10 * 60 // 10 minutes
	dingtalkOAuthDefaultRedirectTo = "/dashboard"
	dingtalkOAuthDefaultFrontendCB = "/auth/dingtalk/callback"

	dingtalkOAuthMaxRedirectLen      = 2048
	dingtalkOAuthMaxFragmentValueLen = 512
	dingtalkOAuthMaxSubjectLen       = 64
)

// dingtalkTokenRequest 钉钉 Token 请求体（JSON 格式）
type dingtalkTokenRequest struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	Code         string `json:"code"`
	GrantType    string `json:"grantType"`
}

// dingtalkTokenResponse 钉钉 Token 响应
type dingtalkTokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
	ExpireIn     int64  `json:"expireIn"`
	CorpID       string `json:"corpId,omitempty"`
}

type dingtalkTokenExchangeError struct {
	StatusCode          int
	ProviderError       string
	ProviderDescription string
	Body                string
}

func (e *dingtalkTokenExchangeError) Error() string {
	if e == nil {
		return ""
	}
	parts := []string{fmt.Sprintf("token exchange status=%d", e.StatusCode)}
	if strings.TrimSpace(e.ProviderError) != "" {
		parts = append(parts, "error="+strings.TrimSpace(e.ProviderError))
	}
	if strings.TrimSpace(e.ProviderDescription) != "" {
		parts = append(parts, "error_description="+strings.TrimSpace(e.ProviderDescription))
	}
	return strings.Join(parts, " ")
}

// DingTalkOAuthStart 启动钉钉 OAuth 登录流程
// GET /api/v1/auth/oauth/dingtalk/start?redirect=/dashboard
func (h *AuthHandler) DingTalkOAuthStart(c *gin.Context) {
	cfg, err := h.getDingTalkOAuthConfig(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	state, err := oauth.GenerateState()
	if err != nil {
		response.ErrorFrom(c, infraerrors.InternalServer("OAUTH_STATE_GEN_FAILED", "failed to generate oauth state").WithCause(err))
		return
	}

	redirectTo := sanitizeDingTalkFrontendRedirectPath(c.Query("redirect"))
	if redirectTo == "" {
		redirectTo = dingtalkOAuthDefaultRedirectTo
	}

	secureCookie := isDingTalkRequestHTTPS(c)
	setDingTalkCookie(c, dingtalkOAuthStateCookieName, encodeDingTalkCookieValue(state), dingtalkOAuthCookieMaxAgeSec, secureCookie)
	setDingTalkCookie(c, dingtalkOAuthRedirectCookie, encodeDingTalkCookieValue(redirectTo), dingtalkOAuthCookieMaxAgeSec, secureCookie)

	redirectURI := strings.TrimSpace(cfg.RedirectURL)
	if redirectURI == "" {
		response.ErrorFrom(c, infraerrors.InternalServer("OAUTH_CONFIG_INVALID", "oauth redirect url not configured"))
		return
	}

	authURL, err := buildDingTalkAuthorizeURL(cfg, state, redirectURI)
	if err != nil {
		response.ErrorFrom(c, infraerrors.InternalServer("OAUTH_BUILD_URL_FAILED", "failed to build oauth authorization url").WithCause(err))
		return
	}

	c.Redirect(http.StatusFound, authURL)
}

// DingTalkOAuthCallback 处理钉钉 OAuth 回调
// GET /api/v1/auth/oauth/dingtalk/callback?authCode=...&state=...
func (h *AuthHandler) DingTalkOAuthCallback(c *gin.Context) {
	cfg, cfgErr := h.getDingTalkOAuthConfig(c.Request.Context())
	if cfgErr != nil {
		response.ErrorFrom(c, cfgErr)
		return
	}

	frontendCallback := strings.TrimSpace(cfg.FrontendRedirectURL)
	if frontendCallback == "" {
		frontendCallback = dingtalkOAuthDefaultFrontendCB
	}

	if providerErr := strings.TrimSpace(c.Query("error")); providerErr != "" {
		redirectDingTalkOAuthError(c, frontendCallback, "provider_error", providerErr, c.Query("error_description"))
		return
	}

	// 钉钉使用 authCode 而非 code
	code := strings.TrimSpace(c.Query("authCode"))
	state := strings.TrimSpace(c.Query("state"))
	if code == "" || state == "" {
		redirectDingTalkOAuthError(c, frontendCallback, "missing_params", "missing authCode/state", "")
		return
	}

	secureCookie := isDingTalkRequestHTTPS(c)
	defer func() {
		clearDingTalkCookie(c, dingtalkOAuthStateCookieName, secureCookie)
		clearDingTalkCookie(c, dingtalkOAuthRedirectCookie, secureCookie)
	}()

	expectedState, err := readDingTalkCookieDecoded(c, dingtalkOAuthStateCookieName)
	if err != nil || expectedState == "" || state != expectedState {
		redirectDingTalkOAuthError(c, frontendCallback, "invalid_state", "invalid oauth state", "")
		return
	}

	redirectTo, _ := readDingTalkCookieDecoded(c, dingtalkOAuthRedirectCookie)
	redirectTo = sanitizeDingTalkFrontendRedirectPath(redirectTo)
	if redirectTo == "" {
		redirectTo = dingtalkOAuthDefaultRedirectTo
	}

	tokenResp, err := dingtalkExchangeCode(c.Request.Context(), cfg, code)
	if err != nil {
		description := ""
		var exchangeErr *dingtalkTokenExchangeError
		if errors.As(err, &exchangeErr) && exchangeErr != nil {
			log.Printf(
				"[DingTalk OAuth] token exchange failed: status=%d provider_error=%q provider_description=%q body=%s",
				exchangeErr.StatusCode,
				exchangeErr.ProviderError,
				exchangeErr.ProviderDescription,
				truncateDingTalkLogValue(exchangeErr.Body, 2048),
			)
			description = exchangeErr.Error()
		} else {
			log.Printf("[DingTalk OAuth] token exchange failed: %v", err)
			description = err.Error()
		}
		redirectDingTalkOAuthError(c, frontendCallback, "token_exchange_failed", "failed to exchange oauth code", singleLineDingTalk(description))
		return
	}

	username, subject, err := dingtalkFetchUserInfo(c.Request.Context(), cfg, tokenResp)
	if err != nil {
		log.Printf("[DingTalk OAuth] userinfo fetch failed: %v", err)
		redirectDingTalkOAuthError(c, frontendCallback, "userinfo_failed", "failed to fetch user info", "")
		return
	}

	// 使用合成邮箱避免账号冲突
	email := dingtalkSyntheticEmail(subject)

	jwtToken, _, err := h.authService.LoginOrRegisterOAuth(c.Request.Context(), email, username)
	if err != nil {
		redirectDingTalkOAuthError(c, frontendCallback, "login_failed", infraerrors.Reason(err), infraerrors.Message(err))
		return
	}

	fragment := url.Values{}
	fragment.Set("access_token", jwtToken)
	fragment.Set("token_type", "Bearer")
	fragment.Set("redirect", redirectTo)
	redirectDingTalkWithFragment(c, frontendCallback, fragment)
}

func (h *AuthHandler) getDingTalkOAuthConfig(ctx context.Context) (config.DingTalkOAuthConfig, error) {
	if h != nil && h.settingSvc != nil {
		return h.settingSvc.GetDingTalkOAuthConfig(ctx)
	}
	if h == nil || h.cfg == nil {
		return config.DingTalkOAuthConfig{}, infraerrors.ServiceUnavailable("CONFIG_NOT_READY", "config not loaded")
	}
	if !h.cfg.DingTalk.Enabled {
		return config.DingTalkOAuthConfig{}, infraerrors.NotFound("OAUTH_DISABLED", "dingtalk oauth login is disabled")
	}
	return h.cfg.DingTalk, nil
}

// dingtalkExchangeCode 使用授权码交换访问令牌
// 注意：钉钉使用 JSON body 而非 form-urlencoded
func dingtalkExchangeCode(
	ctx context.Context,
	cfg config.DingTalkOAuthConfig,
	code string,
) (*dingtalkTokenResponse, error) {
	client := req.C().SetTimeout(30 * time.Second)

	reqBody := dingtalkTokenRequest{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Code:         code,
		GrantType:    "authorization_code",
	}

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetBody(reqBody).
		Post(cfg.TokenURL)

	if err != nil {
		return nil, fmt.Errorf("request token: %w", err)
	}

	body := strings.TrimSpace(resp.String())
	if !resp.IsSuccessState() {
		providerErr, providerDesc := parseDingTalkOAuthError(body)
		return nil, &dingtalkTokenExchangeError{
			StatusCode:          resp.StatusCode,
			ProviderError:       providerErr,
			ProviderDescription: providerDesc,
			Body:                body,
		}
	}

	tokenResp, ok := parseDingTalkTokenResponse(body)
	if !ok || strings.TrimSpace(tokenResp.AccessToken) == "" {
		return nil, &dingtalkTokenExchangeError{
			StatusCode: resp.StatusCode,
			Body:       body,
		}
	}
	return tokenResp, nil
}

// dingtalkFetchUserInfo 获取钉钉用户信息
// 钉钉使用自定义 header: x-acs-dingtalk-access-token
// 流程：先获取 unionId，再通过通讯录 API 获取 userId
func dingtalkFetchUserInfo(
	ctx context.Context,
	cfg config.DingTalkOAuthConfig,
	token *dingtalkTokenResponse,
) (username string, subject string, err error) {
	client := req.C().SetTimeout(30 * time.Second)

	// Step 1: 获取用户基本信息（unionId）
	resp, err := client.R().
		SetContext(ctx).
		SetHeader("x-acs-dingtalk-access-token", token.AccessToken).
		SetHeader("Accept", "application/json").
		Get(cfg.UserInfoURL)

	if err != nil {
		return "", "", fmt.Errorf("request userinfo: %w", err)
	}
	if !resp.IsSuccessState() {
		return "", "", fmt.Errorf("userinfo status=%d body=%s", resp.StatusCode, truncateDingTalkLogValue(resp.String(), 512))
	}

	body := resp.String()
	log.Printf("[DingTalk OAuth] userinfo response: %s", truncateDingTalkLogValue(body, 1024))

	unionId := strings.TrimSpace(firstNonEmptyDingTalk(
		getGJSONDingTalk(body, "unionId"),
		getGJSONDingTalk(body, "openId"),
	))
	if unionId == "" {
		return "", "", errors.New("userinfo missing unionId/openId")
	}

	// Step 2: 获取企业内部应用 access_token
	appToken, err := dingtalkGetAppAccessToken(ctx, cfg)
	if err != nil {
		log.Printf("[DingTalk OAuth] get app access token failed: %v, fallback to unionId", err)
		// 获取应用 token 失败，降级使用 unionId
		return unionId, unionId, nil
	}

	// Step 3: 通过 unionId 获取 userId
	userId, err := dingtalkGetUserIdByUnionId(ctx, appToken, unionId)
	if err != nil {
		log.Printf("[DingTalk OAuth] get userId by unionId failed: %v, fallback to unionId", err)
		// 获取 userId 失败，降级使用 unionId
		return unionId, unionId, nil
	}

	log.Printf("[DingTalk OAuth] resolved userId=%s from unionId=%s", userId, unionId)
	return userId, userId, nil
}

// dingtalkGetAppAccessToken 获取钉钉企业内部应用的 access_token
func dingtalkGetAppAccessToken(ctx context.Context, cfg config.DingTalkOAuthConfig) (string, error) {
	client := req.C().SetTimeout(15 * time.Second)

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetBody(map[string]string{
			"appKey":    cfg.ClientID,
			"appSecret": cfg.ClientSecret,
		}).
		Post("https://api.dingtalk.com/v1.0/oauth2/accessToken")

	if err != nil {
		return "", fmt.Errorf("request app access token: %w", err)
	}
	if !resp.IsSuccessState() {
		return "", fmt.Errorf("app token status=%d body=%s", resp.StatusCode, truncateDingTalkLogValue(resp.String(), 512))
	}

	accessToken := strings.TrimSpace(getGJSONDingTalk(resp.String(), "accessToken"))
	if accessToken == "" {
		return "", errors.New("empty app access token in response")
	}
	return accessToken, nil
}

// dingtalkGetUserIdByUnionId 通过 unionId 获取企业内部 userId
func dingtalkGetUserIdByUnionId(ctx context.Context, appAccessToken, unionId string) (string, error) {
	client := req.C().SetTimeout(15 * time.Second)

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetQueryParam("access_token", appAccessToken).
		SetBody(map[string]string{
			"unionid": unionId,
		}).
		Post("https://oapi.dingtalk.com/topapi/user/getbyunionid")

	if err != nil {
		return "", fmt.Errorf("request getbyunionid: %w", err)
	}

	body := resp.String()
	errCode := getGJSONDingTalk(body, "errcode")
	if errCode != "0" {
		errMsg := getGJSONDingTalk(body, "errmsg")
		return "", fmt.Errorf("getbyunionid errcode=%s errmsg=%s", errCode, errMsg)
	}

	userId := strings.TrimSpace(getGJSONDingTalk(body, "result.userid"))
	if userId == "" {
		return "", errors.New("empty userid in getbyunionid response")
	}
	return userId, nil
}

func dingtalkParseUserInfo(body string, cfg config.DingTalkOAuthConfig) (username string, subject string, err error) {
	// 提取用户唯一标识：优先 userId，然后 openId，最后 unionId
	// userId 是企业内部用户标识（如果 API 返回的话）
	subject = firstNonEmptyDingTalk(
		getGJSONDingTalk(body, cfg.UserInfoIDPath),
		getGJSONDingTalk(body, "userid"),
		getGJSONDingTalk(body, "userId"),
		getGJSONDingTalk(body, "openId"),
		getGJSONDingTalk(body, "unionId"),
	)

	subject = strings.TrimSpace(subject)
	if subject == "" {
		return "", "", errors.New("userinfo missing id field (userId/openId/unionId)")
	}
	if !isSafeDingTalkSubject(subject) {
		return "", "", errors.New("userinfo returned invalid id field")
	}

	// 用户名直接使用 userId（subject）
	username = subject

	return username, subject, nil
}

func buildDingTalkAuthorizeURL(cfg config.DingTalkOAuthConfig, state string, redirectURI string) (string, error) {
	u, err := url.Parse(cfg.AuthorizeURL)
	if err != nil {
		return "", fmt.Errorf("parse authorize_url: %w", err)
	}

	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", cfg.ClientID)
	q.Set("redirect_uri", redirectURI)
	if strings.TrimSpace(cfg.Scopes) != "" {
		q.Set("scope", cfg.Scopes)
	}
	q.Set("state", state)
	q.Set("prompt", "consent")

	u.RawQuery = q.Encode()
	return u.String(), nil
}

func redirectDingTalkOAuthError(c *gin.Context, frontendCallback string, code string, message string, description string) {
	fragment := url.Values{}
	fragment.Set("error", truncateDingTalkFragmentValue(code))
	if strings.TrimSpace(message) != "" {
		fragment.Set("error_message", truncateDingTalkFragmentValue(message))
	}
	if strings.TrimSpace(description) != "" {
		fragment.Set("error_description", truncateDingTalkFragmentValue(description))
	}
	redirectDingTalkWithFragment(c, frontendCallback, fragment)
}

func redirectDingTalkWithFragment(c *gin.Context, frontendCallback string, fragment url.Values) {
	u, err := url.Parse(frontendCallback)
	if err != nil {
		c.Redirect(http.StatusFound, dingtalkOAuthDefaultRedirectTo)
		return
	}
	if u.Scheme != "" && !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		c.Redirect(http.StatusFound, dingtalkOAuthDefaultRedirectTo)
		return
	}
	u.Fragment = fragment.Encode()
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	c.Redirect(http.StatusFound, u.String())
}

func firstNonEmptyDingTalk(values ...string) string {
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}

func parseDingTalkOAuthError(body string) (providerErr string, providerDesc string) {
	body = strings.TrimSpace(body)
	if body == "" {
		return "", ""
	}

	// 钉钉错误格式通常为 {"code": "xxx", "message": "xxx"}
	providerErr = firstNonEmptyDingTalk(
		getGJSONDingTalk(body, "code"),
		getGJSONDingTalk(body, "error"),
	)
	providerDesc = firstNonEmptyDingTalk(
		getGJSONDingTalk(body, "message"),
		getGJSONDingTalk(body, "error_description"),
	)

	return providerErr, providerDesc
}

func parseDingTalkTokenResponse(body string) (*dingtalkTokenResponse, bool) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, false
	}

	accessToken := strings.TrimSpace(getGJSONDingTalk(body, "accessToken"))
	if accessToken == "" {
		return nil, false
	}

	refreshToken := strings.TrimSpace(getGJSONDingTalk(body, "refreshToken"))
	expireIn := gjson.Get(body, "expireIn").Int()
	corpID := strings.TrimSpace(getGJSONDingTalk(body, "corpId"))

	return &dingtalkTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpireIn:     expireIn,
		CorpID:       corpID,
	}, true
}

func getGJSONDingTalk(body string, path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	res := gjson.Get(body, path)
	if !res.Exists() {
		return ""
	}
	return res.String()
}

func truncateDingTalkLogValue(value string, maxLen int) string {
	value = strings.TrimSpace(value)
	if value == "" || maxLen <= 0 {
		return ""
	}
	if len(value) <= maxLen {
		return value
	}
	value = value[:maxLen]
	for !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}
	return value
}

func singleLineDingTalk(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return strings.Join(strings.Fields(value), " ")
}

func sanitizeDingTalkFrontendRedirectPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if len(path) > dingtalkOAuthMaxRedirectLen {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		return ""
	}
	if strings.HasPrefix(path, "//") {
		return ""
	}
	if strings.Contains(path, "://") {
		return ""
	}
	if strings.ContainsAny(path, "\r\n") {
		return ""
	}
	return path
}

func isDingTalkRequestHTTPS(c *gin.Context) bool {
	if c.Request.TLS != nil {
		return true
	}
	proto := strings.ToLower(strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")))
	return proto == "https"
}

func encodeDingTalkCookieValue(value string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(value))
}

func decodeDingTalkCookieValue(value string) (string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func readDingTalkCookieDecoded(c *gin.Context, name string) (string, error) {
	ck, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	return decodeDingTalkCookieValue(ck.Value)
}

func setDingTalkCookie(c *gin.Context, name string, value string, maxAgeSec int, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     dingtalkOAuthCookiePath,
		MaxAge:   maxAgeSec,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearDingTalkCookie(c *gin.Context, name string, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     dingtalkOAuthCookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func truncateDingTalkFragmentValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) > dingtalkOAuthMaxFragmentValueLen {
		value = value[:dingtalkOAuthMaxFragmentValueLen]
		for !utf8.ValidString(value) {
			value = value[:len(value)-1]
		}
	}
	return value
}

func isSafeDingTalkSubject(subject string) bool {
	subject = strings.TrimSpace(subject)
	if subject == "" || len(subject) > dingtalkOAuthMaxSubjectLen {
		return false
	}
	for _, r := range subject {
		switch {
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r == '_' || r == '-':
		default:
			return false
		}
	}
	return true
}

func dingtalkSyntheticEmail(subject string) string {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return ""
	}
	return "dingtalk-" + subject + service.DingTalkSyntheticEmailDomain
}
