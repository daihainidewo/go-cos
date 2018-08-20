// Package conf conf
// file create by daihao, time is 2018/8/10 10:42
package conf

type Conf struct {
	SecretId  string `toml:"secret_id,omitempty"`
	SecretKey string `toml:"secret_key,omitempty"`
	AppId     string `toml:"app_id,omitempty"`
	Host      string `toml:"host,omitempty"`
}

type BackupConf struct {
	ProjectRootPath string `json:"project_root_path"`
	LastUpdateTime  int64  `json:"last_update_time"`
}
