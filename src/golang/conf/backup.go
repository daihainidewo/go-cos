// Package conf conf
// file create by daihao, time is 2018/8/10 10:42
package conf

type BackupConf struct {
	ProjectRootPath string `json:"project_root_path"`
	LastUpdateTime  int64  `json:"last_update_time"`
}
