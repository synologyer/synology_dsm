package types

// 登录
// https://kb.synology.cn/zh-cn/DG/DSM_Login_Web_API_Guide/2#x_anchor_iddbcc293edb
type AuthResponse struct {
	Data struct {
		Sid       string `json:"sid"`       // 设备唯一标识
		Synotoken string `json:"synotoken"` // Token
	} `json:"data"` // 数据
	Success bool `json:"success"` // 是否成功
}
