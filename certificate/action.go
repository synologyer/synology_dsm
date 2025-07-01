package certificate

import (
	"fmt"
	"github.com/synologyer/synology_dsm/core"
	"github.com/synologyer/synology_dsm/openapi"
	"github.com/synologyer/synology_dsm/types"
	"strings"
	"time"
)

// 上传证书
// isExist: 是否存在
func Action(openapiClient *openapi.Client, certBundle *core.CertBundle) (isExist bool, err error) {

	// 1. 获取证书列表
	var certListResp types.CommonResponse[types.CertificateListResponse]
	_, err = openapiClient.R().
		SetQueryParam("api", "SYNO.Core.Certificate.CRT").
		SetQueryParam("version", "1").
		SetQueryParam("method", "list").
		SetResult(&certListResp).
		SetContentType("application/json").
		Get("")
	if err != nil {
		return false, fmt.Errorf("获取证书列表错误: %w", err)
	}
	// 检查证书列表响应
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

	// 2. 上传证书
	var certUpdateResp types.CommonResponse[types.CertificateUpdateResponse]
	form := map[string]string{
		"desc": certBundle.GetNote(),
	}
	if existCertID != "" {
		form["id"] = existCertID
	}

	_, err = openapiClient.R().
		SetQueryParam("api", "SYNO.Core.Certificate").
		SetQueryParam("version", "1").
		SetQueryParam("method", "import").
		SetFormData(form).
		SetFileReader("cert", "cert.pem", strings.NewReader(certBundle.Certificate)).
		SetFileReader("key", "privkey.pem", strings.NewReader(certBundle.PrivateKey)).
		SetFileReader("inter_cert", "chain.pem", strings.NewReader(certBundle.CertificateChain)).
		SetResult(&certUpdateResp).
		SetContentType("application/json").
		SetForceResponseContentType("application/json").
		Post("")
	if err != nil {
		return false, fmt.Errorf("上传证书错误: %w", err)
	}
	// 检查证书上传响应
	if !certUpdateResp.Success {
		return false, fmt.Errorf("上传证书失败")
	}

	return false, nil
}
