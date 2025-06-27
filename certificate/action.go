package certificate

import (
	"fmt"
	"strings"

	"github.com/whc95800/synology_dsm/core"
	"github.com/whc95800/synology_dsm/openapi"
	"github.com/whc95800/synology_dsm/types"
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
		err = fmt.Errorf("获取证书列表错误: %w", err)
		return
	}

	var existCertID string
	for _, certInfo := range certListResp.Data.Certificates {
		if strings.EqualFold(certInfo.Desc, certBundle.GetNote()) &&
			stringSlicesEqual(certInfo.Subject.SubAltName, certBundle.DNSNames) {
			existCertID = certInfo.ID
			isExist = true // 标记已存在（但仍继续上传）
			break
		}
	}

	// 2. 上传证书（无论是否存在都上传，存在就替换）
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
		SetFileReader("cert", "certificate.pem", strings.NewReader(certBundle.Certificate)).
		SetFileReader("key", "private_key.pem", strings.NewReader(certBundle.PrivateKey)).
		Post("")
	if err != nil {
		return isExist, fmt.Errorf("上传证书错误: %w", err)
	}

	return isExist, nil
}

// 比较两个 []string 是否内容相同（忽略顺序）
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[string]bool)
	for _, v := range a {
		aMap[v] = true
	}

	for _, v := range b {
		if !aMap[v] {
			return false
		}
	}

	return true
}
