package lakala

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/SuperH-0630/gopay"
	"github.com/SuperH-0630/gopay/pkg/util"
	"github.com/SuperH-0630/gopay/pkg/xhttp"
)

// Client lakala
type Client struct {
	ctx            context.Context   // 上下文
	PartnerCode    string            // partner_code:商户编码，由4~6位大写字母或数字构成
	credentialCode string            // credential_code:系统为商户分配的开发校验码，请妥善保管，不要在公开场合泄露
	bodySize       int               // http response body size(MB), default is 10MB
	IsProd         bool              // 是否生产环境
	DebugSwitch    gopay.DebugSwitch // 调试开关，是否打印日志
}

// NewClient 初始化lakala户端
// partnerCode: 商户编码，由4~6位大写字母或数字构成
// credentialCode: 系统为商户分配的开发校验码，请妥善保管，不要在公开场合泄露
// isProd: 是否生产环境
func NewClient(partnerCode, credentialCode string, isProd bool) (client *Client, err error) {
	if partnerCode == util.NULL || credentialCode == util.NULL {
		return nil, gopay.MissLakalaInitParamErr
	}
	client = &Client{
		ctx:            context.Background(),
		PartnerCode:    partnerCode,
		credentialCode: credentialCode,
		IsProd:         isProd,
		DebugSwitch:    gopay.DebugOff,
	}
	return client, nil
}

// SetBodySize 设置http response body size(MB)
func (c *Client) SetBodySize(sizeMB int) {
	if sizeMB > 0 {
		c.bodySize = sizeMB
	}
}

// 公共参数处理 Query Params
func (c *Client) pubParamsHandle() (param string, err error) {
	bm := make(gopay.BodyMap)
	bm.Set("time", time.Now().UnixMilli())
	bm.Set("nonce_str", util.RandomString(20))
	sign, err := c.getRsaSign(bm)
	if err != nil {
		return "", fmt.Errorf("GetRsaSign Error: %w", err)
	}
	bm.Set("sign", sign)
	param = bm.EncodeURLParams()
	return
}

// 验证签名
func VerifySign(notifyReq *NotifyRequest, partnerCode string, credentialCode string) (err error) {
	validStr := fmt.Sprintf("%v&%v&%v&%v", partnerCode, notifyReq.Time, notifyReq.NonceStr, credentialCode)
	h := sha256.New()
	h.Write([]byte(validStr))
	validSign := strings.ToLower(hex.EncodeToString(h.Sum(nil)))
	if notifyReq.Sign != validSign {
		return fmt.Errorf("签名验证失败")
	}
	return
}

// getRsaSign 获取签名字符串
func (c *Client) getRsaSign(bm gopay.BodyMap) (sign string, err error) {
	var (
		partnerCode    = c.PartnerCode
		ts             = bm.Get("time")
		nonceStr       = bm.Get("nonce_str")
		credentialCode = c.credentialCode
	)
	if ts == "" || nonceStr == "" {
		return "", fmt.Errorf("签名缺少必要的参数")
	}
	validStr := fmt.Sprintf("%v&%v&%v&%v", partnerCode, ts, nonceStr, credentialCode)
	h := sha256.New()
	h.Write([]byte(validStr))
	sign = strings.ToLower(hex.EncodeToString(h.Sum(nil)))
	return
}

// PUT 发起请求
func (c *Client) doPut(ctx context.Context, path string, bm gopay.BodyMap) (bs []byte, err error) {
	httpClient := xhttp.NewClient().Type(xhttp.TypeJSON)
	if c.bodySize > 0 {
		httpClient.SetBodySize(c.bodySize)
	}
	httpClient.Header.Add("Content-Type", "application/json")
	httpClient.Header.Add("Accept", "application/json")
	var url = baseUrlProd + path
	param, err := c.pubParamsHandle()
	if err != nil {
		return nil, err
	}
	uri := url + "?" + param
	res, bs, err := httpClient.Put(uri).SendBodyMap(bm).EndBytes(ctx)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Request Error, StatusCode = %d", res.StatusCode)
	}
	return bs, nil
}

// PUT 发起请求
func (c *Client) doPost(ctx context.Context, path string, bm gopay.BodyMap) (bs []byte, err error) {
	httpClient := xhttp.NewClient().Type(xhttp.TypeJSON)
	if c.bodySize > 0 {
		httpClient.SetBodySize(c.bodySize)
	}
	httpClient.Header.Add("Content-Type", "application/json")
	httpClient.Header.Add("Accept", "application/json")
	var url = baseUrlProd + path
	param, err := c.pubParamsHandle()
	if err != nil {
		return nil, err
	}
	uri := url + "?" + param
	res, bs, err := httpClient.Post(uri).SendBodyMap(bm).EndBytes(ctx)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Request Error, StatusCode = %d", res.StatusCode)
	}
	return bs, nil
}

// GET 发起请求
func (c *Client) doGet(ctx context.Context, path, queryParams string) (bs []byte, err error) {
	httpClient := xhttp.NewClient().Type(xhttp.TypeJSON)
	if c.bodySize > 0 {
		httpClient.SetBodySize(c.bodySize)
	}
	httpClient.Header.Add("Content-Type", "application/json")
	httpClient.Header.Add("Accept", "application/json")

	var url = baseUrlProd + path
	param, err := c.pubParamsHandle()
	if err != nil {
		return nil, err
	}
	if queryParams != "" {
		param = param + "&" + queryParams
	}
	uri := url + "?" + param
	res, bs, err := httpClient.Get(uri).EndBytes(ctx)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Request Error, StatusCode = %d", res.StatusCode)
	}
	return bs, nil

}
