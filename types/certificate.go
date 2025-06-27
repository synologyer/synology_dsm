package types

// 获取证书列表 响应参数

type CertificateListResponse struct {
	Data struct {
		Certificates []struct {
			Desc    string `json:"desc"` // 证书描述
			ID      string `json:"id"`   // 证书ID
			Subject struct {
				CommonName string   `json:"common_name"`  // 主域名
				SubAltName []string `json:"sub_alt_name"` //
			} `json:"subject"`
		} `json:"certificates"`
	} `json:"data"`
	Success bool `json:"success"`
}
