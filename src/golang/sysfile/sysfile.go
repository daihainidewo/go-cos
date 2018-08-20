// Package sysfile sysfile
// file create by daihao, time is 2018/8/15 13:00
package sysfile

// Sysfile
type Sysfile struct {
	projPath string // 配置文件目录
	rootpath string // 文件目录
}

// SetCosPath SetCosPath
func NewSysfile(rootpath string) (*Sysfile) {
	return &Sysfile{
		rootpath: rootpath,
		projPath: rootpath + "/.gocos",
	}
}

// RootPath RootPath
func (s *Sysfile) RootPath() string {
	return s.rootpath
}

// CosPath CosPath
func (s *Sysfile) ProjPath() string {
	return s.projPath
}

// AddBufferPath AddBufferPath
func (s *Sysfile) AddBufferPath() string {
	return s.projPath + "/.addbuffer"
}

// BackupPath BackupPath
//func (s *Sysfile) BackupPath() string {
//	return s.projPath + "/.backup"
//}

// ConfPath ConfPath
func (s *Sysfile) ConfPath() string {
	return s.projPath + "/conf.toml"
}

// IgnorePath IgnorePath
func (s *Sysfile) IgnorePath() string {
	return s.projPath + "/ignore"
}

// SysfilePath SysfilePath
func (s *Sysfile) SysfilePath() string {
	// TODO
	return s.projPath + "/.sysfile"
}
