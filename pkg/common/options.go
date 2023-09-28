package common

type DBOptions struct {
	Driver       string        `yaml:"driver"`
	MysqlOptions *MysqlOptions `yaml:"mysqloptions"`
	ESOptions    *ESOptions    `yaml:"esoptions"`
}

// 1. 定一个 MysqlOptions struct, 字段定一个存储client初始化需要的参数
type MysqlOptions struct {
	Host         string `yaml:"host"`
	Port         string `yaml:"port"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	DBName       string `yaml:"dbName"`
	Charset      string `yaml:"charset"`
	ParseTime    bool   `yaml:"parseTime"`
	MaxIdleConns int    `yaml:"maxIdleConns"`
	MaxOpenConns int    `yaml:"maxOpenConns"`
	MaxLifetime  int    `yaml:"maxLifetime"`
	LogMode      bool   `yaml:"logMode"`
}

func NewMysqlOptions() *MysqlOptions {
	return &MysqlOptions{
		Host: "localhost",
	}
}

type ESOptions struct {
	EndPoint string `yaml:"endpoint"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func NewEsOptions() *ESOptions {
	return &ESOptions{
		EndPoint: "localhost",
		Username: "username",
		Password: "passwd",
	}
}

func NewDefaultOptions() *DBOptions {
	return &DBOptions{
		Driver:       "elasticsearch",
		MysqlOptions: NewMysqlOptions(),
		ESOptions:    NewEsOptions(),
	}
}
