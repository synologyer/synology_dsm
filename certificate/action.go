package certificate

import (
	"encoding/json"
	"fmt"
	"github.com/synologyer/synology_dsm/core"
	"github.com/synologyer/synology_dsm/openapi"
	"github.com/synologyer/synology_dsm/types"
	"io"
	"strings"
	"time"
)

// 上传证书
// isExist: 是否存在
func Action(openapiClient *openapi.Client, certBundle *core.CertBundle) (isExist bool, err error) {

	// 1. 获取证书列表
	var certListResp types.CertificateListResponse
	_, err = openapiClient.R().
		SetQueryParam("api", "SYNO.Core.Certificate.CRT").
		SetQueryParam("version", "1").
		SetQueryParam("method", "list").
		SetResult(&certListResp).
		Get("")
	if err != nil {
		return false, fmt.Errorf("获取证书列表错误: %w", err)
	}
	if !certListResp.Success {
		return false, fmt.Errorf("获取证书列表失败")
	}

	var existCertID string
	const customLayout = "Jan 02 15:04:05 2006 MST"
	const renewBefore = 30 * 24 * time.Hour
	for _, certInfo := range certListResp.Data.Certificates {
		if strings.EqualFold(certInfo.Desc, certBundle.GetNote()) {
			if certBundle.IsDNSNamesMatch(certInfo.Subject.SubAltName) {
				validTill, err := time.Parse(customLayout, certInfo.ValidTill)
				if err != nil {
					return false, fmt.Errorf("解析过期时间失败: %w", err)
				}
				// 证书距离到期还有超过 30 天，无需更新
				if validTill.Sub(time.Now()) > renewBefore {
					return true, nil
				}
				// 否则需要更新
				existCertID = certInfo.ID
			}
		}
	}

	// 2. 上传证书（无论是否存在都上传，存在就替换）
	var certUpdateResp types.CommonResponse
	form := map[string]string{
		"desc": certBundle.GetNote(),
	}
	if existCertID != "" {
		form["id"] = existCertID
	}

	resp, err := openapiClient.R().
		SetQueryParam("api", "SYNO.Core.Certificate").
		SetQueryParam("version", "1").
		SetQueryParam("method", "import").
		SetFormData(form).
		SetFileReader("cert", "cert.pem", strings.NewReader(certBundle.Certificate)).
		SetFileReader("key", "privkey.pem", strings.NewReader(certBundle.PrivateKey)).
		SetFileReader("inter_cert", "chain.pem", strings.NewReader(certBundle.CertificateChain)).
		Post("")
	if err != nil {
		return false, fmt.Errorf("上传证书错误: %w", err)
	}

	// 打印原始响应体（调试用）
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return false, fmt.Errorf("读取响应体失败: %w", readErr)
	}

	// 解析 JSON
	if err := json.Unmarshal(bodyBytes, &certUpdateResp); err != nil {
		return false, fmt.Errorf("解析响应体失败: %w", err)
	}

	if !certUpdateResp.Success {
		return false, fmt.Errorf("上传证书失败，DSM返回: %s", string(bodyBytes))
	}

	return false, nil
}
