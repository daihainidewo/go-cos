# go-cos
利用golang写一个类似于git的操作，将文件上传到腾讯云对象存储上

## 初始化项目
`cos init`，初始化项目，cos会生成默认文件，包括：
* conf.toml，配置文件，采用toml格式，主要填写secret_id，secret_key，app_id，host等系统必须参数
* ignore，填写忽略列表，文件夹以'/'结尾，过滤不需要上传的文件

## 添加需要上传的文件
`cos add`，将上次上传之后修改后的文件进行标记，在push之前需先执行add

## 上传文件
`cos push`，将add的文件上传至host指定的目标