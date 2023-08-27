package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Conf 定义全局的变量
var Conf = new(SrvConfig)

// viper.GetXxx()读取的方式
// Viper使用的是 `mapstructure`

// SrvConfig 服务配置
type SrvConfig struct {
	Name    string `mapstructure:"name"`
	Mode    string `mapstructure:"mode"`
	Version string `mapstructure:"version"`
	// snowflake
	StartTime string `mapstructure:"start_time"`
	MachineID int64  `mapstructure:"machine_id"`

	IP       string `mapstructure:"ip"`
	RpcPort  int    `mapstructure:"rpcPort"`
	HttpPort int    `mapstructure:"httpPort"`

	*LogConfig      `mapstructure:"log"`
	*MySQLConfig    `mapstructure:"mysql"`
	*RedisConfig    `mapstructure:"redis"`
	*ConsulConfig   `mapstructure:"consul"`
	*RocketMqConfig `mapstructure:"rocketmq"`

	*GoodsService     `mapstructure:"goods_service"`
	*RepertoryService `mapstructure:"repertory_service"`
}

type GoodsService struct {
	Name string `mapstructure:"name"`
}

type RepertoryService struct {
	Name string `mapstructure:"name"`
}

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DB           string `mapstructure:"dbname"`
	Port         int    `mapstructure:"port"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Password     string `mapstructure:"password"`
	Port         int    `mapstructure:"port"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

type RocketMqConfig struct {
	Addr    string `mapstructure:"addr"`
	GroupId string `mapstructure:"group_id"`
	Topic   struct {
		PayTimeOut        string `mapstructure:"pay_timeout"`
		RepertoryRollback string `mapstructure:"repertory_rollback"`
	} `mapstructure:"topic"`
}

type ConsulConfig struct {
	Addr string `mapstructure:"addr"`
}

// Init 整个服务配置文件初始化的方法
func Init(filePath string) (err error) {

	viper.SetConfigFile(filePath)

	err = viper.ReadInConfig() // 读取配置信息
	if err != nil {
		// 读取配置信息失败
		fmt.Printf("viper.ReadInConfig failed, err:%v\n", err)
		return
	}

	// 把读取到的配置信息反序列化到 Conf 变量中
	if err := viper.Unmarshal(Conf); err != nil {
		fmt.Printf("viper.Unmarshal failed, err:%v\n", err)
	}

	viper.WatchConfig() // 监听配置文件
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("配置文件修改了...")
		if err := viper.Unmarshal(Conf); err != nil {
			fmt.Printf("viper.Unmarshal failed, err:%v\n", err)
		}
	})

	return
}
