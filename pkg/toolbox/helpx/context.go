package helpx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/hashicorp/go-version"
	"go.opentelemetry.io/otel/trace"

	"kratos-admin/pkg/toolbox/claim"
	"kratos-admin/pkg/toolbox/errorx"
)

const (
	defaultLang   = "en"
	userIdKey     = "helper.userId"
	OSTypeUnknown = 0
	OSTypeAndroid = 1
	OSTypeIOS     = 2
	OSNameUnknown = "unknown"
	OSNameAndroid = "android"
	OSNameIOS     = "ios"

	PlatformTypeUnknown int32 = 0             // 未知端类型
	PlatformAndroid           = "app_android" // 安卓 端运行环境
	PlatformTypeAndroid int32 = 1             // 安卓端类型
	PlatformIos               = "app_ios"     // ios 端运行环境
	PlatformTypeIos     int32 = 2             // ios端类型
	PlatformH5                = "web_h5"      // h5 端运行环境
	PlatformTypeMome    int32 = 1             // mome-app
	PlatformTypeH5      int32 = 3             // h5端类型
)

var (
	countryHeaders = []string{"CF-IPCountry", "Cloudfront-Viewer-Country"}
	ipHeaders      = []string{"X-Forwarded-For", "X-Real-IP", "CF-Connecting-IP"}
)

func SetUserId(ctx context.Context, userId int64) {
	// 使用metadata保存userId，方便读取
	md, ok := metadata.FromServerContext(ctx)
	if !ok {
		panic(errors.New("metadata server middleware must register"))
	}
	md.Set(userIdKey, strconv.FormatInt(userId, 10))
}

func GetUserId(ctx context.Context) int64 {
	md, ok := metadata.FromServerContext(ctx)
	if !ok {
		return 0
	}
	userIdStr := md.Get(userIdKey)
	if len(userIdStr) == 0 {
		return 0
	}
	userId, _ := strconv.ParseInt(userIdStr, 10, 64)
	return userId
}

func GetUserIdByToken(ctx context.Context, secret string) (int64, error) {
	if token := GetAuthToken(ctx); token != "" {
		uc, err := claim.ParseWithClaims(token, secret)
		if err != nil {
			return 0, err
		}
		return uc.UserId, nil
	}
	return 0, errors.New("jwt token is empty")
}

type DeviceInfo struct {
	KCID      string `json:"kcid,omitempty"`
	IDFV      string `json:"idfv,omitempty"`
	IDFA      string `json:"idfa,omitempty"`
	GAID      string `json:"gaid,omitempty"`
	AndroidID string `json:"android_id,omitempty"`
	KSID      string `json:"ksid,omitempty"`
	H5ID      string `json:"h5id,omitempty"`
	DeviceId  string `json:"device_id,omitempty"`
}

func (d *DeviceInfo) String() string {
	if d == nil {
		return ""
	}

	data, _ := json.Marshal(d)

	return string(data)
}

func GetDeviceInfo(ctx context.Context) *DeviceInfo {
	deviceInfo := &DeviceInfo{}
	// 从header中获取deviceId
	if tr, ok := transport.FromServerContext(ctx); ok {
		headers := tr.RequestHeader()
		// ios
		if kcid := headers.Get("X-KCID"); kcid != "" {
			deviceInfo.KCID = kcid
			if deviceInfo.DeviceId == "" {
				deviceInfo.DeviceId = fmt.Sprintf("kcid:%s", kcid)
			}
		}
		if idfv := headers.Get("X-IDFV"); idfv != "" {
			deviceInfo.IDFV = idfv
			if deviceInfo.DeviceId == "" {
				deviceInfo.DeviceId = fmt.Sprintf("idfv:%s", idfv)
			}
		}
		if idfa := headers.Get("X-IDFA"); idfa != "" && idfa != "00000000-0000-0000-0000-000000000000" {
			deviceInfo.IDFA = idfa
			if deviceInfo.DeviceId == "" {
				deviceInfo.DeviceId = fmt.Sprintf("idfa:%s", idfa)
			}
		}
		// android
		if gaid := headers.Get("X-GAID"); gaid != "" && gaid != "00000000-0000-0000-0000-000000000000" {
			deviceInfo.GAID = gaid
			if deviceInfo.DeviceId == "" {
				deviceInfo.DeviceId = fmt.Sprintf("gaid:%s", gaid)
			}
		}
		if androidId := headers.Get("X-Android-ID"); androidId != "" {
			deviceInfo.AndroidID = androidId
			if deviceInfo.DeviceId == "" {
				deviceInfo.DeviceId = fmt.Sprintf("android:%s", androidId)
			}
		}
		if ksid := headers.Get("X-KSID"); ksid != "" {
			deviceInfo.KSID = ksid
			if deviceInfo.DeviceId == "" {
				deviceInfo.DeviceId = fmt.Sprintf("ksid:%s", ksid)
			}
		}
		// h5
		if h5id := headers.Get("X-H5ID"); h5id != "" {
			deviceInfo.H5ID = h5id
			if deviceInfo.DeviceId == "" {
				deviceInfo.DeviceId = fmt.Sprintf("h5id:%s", h5id)
			}
		}
	}
	return deviceInfo
}

func GetDeviceId(ctx context.Context) string {
	deviceInfo := GetDeviceInfo(ctx)
	return deviceInfo.DeviceId
}

// CompareVersion returns -1, 0, or 1 if this version is lower than, equal to, or greater than the other version.
func CompareVersion(version1, version2 string) int {
	v1, err := version.NewVersion(version1)
	if err != nil {
		return -1
	}
	v2, err := version.NewVersion(version2)
	if err != nil {
		return 1
	}

	return v1.Compare(v2)
}

func GetPath(ctx context.Context) string {
	if req, ok := http.RequestFromServerContext(ctx); ok {
		return fmt.Sprintf("%s:%s", req.Method, req.URL.Path)
	} else {
		return ""
	}
}

func GetAcceptLanguage(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		lang := tr.RequestHeader().Get("Accept-Language")
		return lang
	}
	return ""
}

// GetLanguage 获取语言code
func GetLanguage(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		lang := tr.RequestHeader().Get("X-Language")
		if lang == "" {
			return defaultLang
		}
		return lang
	}
	return defaultLang
}

func GetIP(ctx context.Context) string {
	// 尝试从HTTP请求中获取(不用大小写敏感的transport)
	req, ok := http.RequestFromServerContext(ctx)
	if !ok {
		return ""
	}
	// header头获取
	for _, key := range ipHeaders {
		if ips := req.Header.Get(key); ips != "" {
			return strings.Split(ips, ",")[0]
		}
	}
	// 直接获取
	ip, _, _ := net.SplitHostPort(req.RemoteAddr)
	return ip
}

// GetCountry 获取header中的国家
func GetCountry(ctx context.Context) string {
	req, ok := http.RequestFromServerContext(ctx)
	if !ok {
		return ""
	}
	for _, key := range countryHeaders {
		if country := req.Header.Get(key); country != "" {
			return country
		}
	}
	return ""
}

func GetPhoneOS(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		return tr.RequestHeader().Get("X-OS")
	}
	return ""
}

func GetPhoneOSType(ctx context.Context) int32 {
	osName := GetPhoneOS(ctx)
	switch osName {
	case OSNameAndroid:
		return OSTypeAndroid
	case OSNameIOS:
		return OSTypeIOS
	}
	return 0
}

func GetOsNameByType(ctx context.Context, osType int32) string {
	switch osType {
	case OSTypeAndroid:
		return OSNameAndroid
	case OSTypeIOS:
		return OSNameIOS
	}
	return ""
}

func GetXAppInfo(ctx context.Context) (url.Values, error) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		appInfoStr := tr.RequestHeader().Get("X-App-Info")
		appInfo, err := url.ParseQuery(appInfoStr)
		if err != nil {
			return appInfo, errorx.WithStack(err)
		}
		return appInfo, nil
	}
	return nil, errorx.New("no app info")
}

// GetAppVersion 获取app版本号
func GetAppVersion(ctx context.Context) string {
	appInfo, err := GetXAppInfo(ctx)
	if err != nil {
		return ""
	}
	return appInfo.Get("versionName")
}

func GetVersionCode(ctx context.Context) int64 {
	appInfo, err := GetXAppInfo(ctx)
	if err != nil {
		return 0
	}
	codeStr := appInfo.Get("versionCode")
	if versionCode, err := strconv.ParseInt(codeStr, 10, 64); err == nil {
		return versionCode
	}
	return 0
}

func GetNetwork(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		appInfoStr := tr.RequestHeader().Get("X-Device-Info")
		appInfo, err := url.ParseQuery(appInfoStr)
		if err != nil {
			return ""
		}
		return appInfo.Get("network")
	}
	return ""
}

func GetDeviceModel(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		appInfoStr := tr.RequestHeader().Get("X-Device-Info")
		appInfo, err := url.ParseQuery(appInfoStr)
		if err != nil {
			return ""
		}
		return appInfo.Get("device_model")
	}
	return ""
}

func GetPName(ctx context.Context) int32 {
	if tr, ok := transport.FromServerContext(ctx); ok {
		PNameStr := tr.RequestHeader().Get("X-PName")
		if PNameStr == "" {
			return 0
		}
		PNAme, _ := strconv.Atoi(PNameStr)
		//nolint:gosec
		return int32(PNAme)
	}
	return 0
}

// IsLowerVersion 判断当前版本是否低于指定版本
func IsLowerVersion(ctx context.Context, versionName string) bool {
	appVersion := GetAppVersion(ctx)
	if appVersion == "" {
		return true
	}
	if CompareVersion(appVersion, versionName) < 0 {
		return true
	}
	return false
}

// GetRequestHost 获取请求的域名
func GetRequestHost(ctx context.Context) string {
	if hp, ok := http.RequestFromServerContext(ctx); ok {
		return hp.Host
	}
	return ""
}

func GetTimezone(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		return tr.RequestHeader().Get("X-Timezone")
	}
	return ""
}

func GetTraceId(ctx context.Context) string {
	if span := trace.SpanContextFromContext(ctx); span.HasTraceID() {
		return span.TraceID().String()
	}
	return ""
}

func GetUa(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		return tr.RequestHeader().Get("user-agent")
	}
	return ""
}

// GetOsVersion 获取系统版本号
func GetOsVersion(ctx context.Context) string {
	appInfo, err := GetXAppInfo(ctx)
	if err != nil {
		return ""
	}
	return appInfo.Get("osVersion")
}

func GetAuthToken(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		authorization := tr.RequestHeader().Get("Authorization")
		return strings.TrimPrefix(authorization, "Bearer ")
	}
	return ""
}

func GetXTimestamp(ctx context.Context) int64 {
	if tr, ok := transport.FromServerContext(ctx); ok {
		str := tr.RequestHeader().Get("X-Timestamp")
		if xTimestamp, err := strconv.ParseInt(str, 10, 64); err == nil {
			return xTimestamp
		}
	}
	return 0
}

func GetUtm(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		return tr.RequestHeader().Get("X-UTM")
	}
	return ""
}

func GetXReferrer(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		return tr.RequestHeader().Get("X-Referrer")
	}
	return ""
}

func GetInstallReferrer(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		return tr.RequestHeader().Get("X-Install-Referrer")
	}
	return ""
}

func GetHttpBody(ctx context.Context, reset bool) (string, error) {
	if req, ok := http.RequestFromServerContext(ctx); ok {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return "", errorx.WithStack(err)
		}
		if reset {
			req.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		return string(body), nil
	}
	return "", errorx.New("http request not found")
}

// NewContextWithTrace 新 ctx 并继承 trace
func NewContextWithTrace(ctx context.Context) context.Context {
	// 从原始 ctx 中提取 span
	span := trace.SpanFromContext(ctx)

	// 创建一个新的 context（无 cancel），但保留 span
	newCtx := context.Background()
	newCtx = trace.ContextWithSpan(newCtx, span)

	return newCtx
}

// GetAppsflyerId 获取归因设备唯一标识。
// 历史上为 AppsFlyer UID（X-AFID）；停 AF 后由客户端用安装级设备 ID 填充同一字段。
// 无 X-AFID 时按端回退：Android X-KSID，iOS X-KCID（与 GetDeviceInfo 一致）。
func GetAppsflyerId(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		h := tr.RequestHeader()
		if v := h.Get("X-AFID"); v != "" {
			return v
		}
		if v := h.Get("X-KSID"); v != "" {
			return v
		}
		if v := h.Get("X-KCID"); v != "" {
			return v
		}
	}
	return ""
}

// GetHost 获取 host
func GetHost(ctx context.Context) string {
	if req, ok := http.RequestFromServerContext(ctx); ok {
		return req.Host
	}
	return ""
}

// GetSessionId 获取 session id
func GetSessionId(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		return tr.RequestHeader().Get("X-Session-ID")
	}
	return ""
}

// GetPlatform 获取 端运行环境 1:app_ios 2:app_android 3:web_h5 4:web_pc
func GetPlatform(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		return tr.RequestHeader().Get("X-Platform")
	}
	return ""
}

// GetPlatformType 端环境类型(逐渐废弃)
func GetPlatformType(ctx context.Context) int32 {
	platform := GetPlatform(ctx)
	switch platform {
	case PlatformAndroid:
		return PlatformTypeAndroid
	case PlatformIos:
		return PlatformTypeIos
	case PlatformH5:
		return PlatformTypeH5
	}
	return 0
}

// GetPlatformTypeV2 端环境类型 1:mome_app 3:web_h5
func GetPlatformTypeV2(ctx context.Context) int32 {
	platform := GetPlatformTypeFallbackOs(ctx)
	switch platform {
	case PlatformTypeAndroid, PlatformTypeIos:
		return PlatformTypeMome
	case PlatformTypeH5:
		return PlatformTypeH5
	}
	return 0
}

// GetPlatformTypeFallbackOs 未传platform参数, 回退旧版基于OS逻辑
func GetPlatformTypeFallbackOs(ctx context.Context) int32 {
	platformType := GetPlatformType(ctx)
	if platformType != PlatformTypeUnknown {
		return platformType
	}
	return GetPhoneOSType(ctx)
}
