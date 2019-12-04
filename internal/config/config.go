package config

import (
	"fmt"
	"github.com/spf13/viper"
	"time"
)

type (
	DBConf struct {
		DSN        string        `mapstructure:"dsn"`
		MaxOpenCon int           `mapstructure:"max_open_con"`
		MaxidleCon int           `mapstructure:"max_idle_con"`
		Lifetime   time.Duration `mapstructure:"lifetime_second"`
	}

	ElasticConf struct {
		Host      string `mapstructure:"host"`
		Port      int    `mapstructure:"port"`
		IndexName string `mapstructure:"index_name"`
	}

	ServerConf struct {
		Bind              string `mapstructure:"bind"`
		ReadTimeout       int64  `mapstructure:"read_timeout"`
		ReadHeaderTimeout int64  `mapstructure:"read_header_timeout"`
		WriteTimeout      int64  `mapstructure:"write_timeout"`
		IdleTimeout       int64  `mapstructure:"idle_timeout"`
		MaxWorker         int    `mapstructure:"max_worker"`
		MaxQueue          int    `mapstructure:"max_queue"`
		PaginationLimit   int    `mapstructure:"pagination_limit"`
		CacheTTL          int    `mapstructure:"cache_ttl"`  // in second
	}

	RedisConf struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		DB       int    `mapstructure:"db"`
		Password string `mapstructure:"password"`
	}

	BrokerConf struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Vhost    string `mapstructure:"vhost"`
	}

	Config struct {
		Broker  BrokerConf  `mapstructure:"broker"`
		DB      DBConf      `mapstructure:"db"`
		Elastic ElasticConf `mapstructure:"elastic"`
		Redis   RedisConf   `mapstructure:"redis"`
		Srv     ServerConf  `mapstructure:"server"`
	}
)

var appConfig Config

func (ec ElasticConf) DSN() string {
	return fmt.Sprintf("http://%s:%d",
		ec.Host,
		ec.Port)
}

func (b BrokerConf) DSN() string {
	var auth = ""
	if len(b.Username) > 0 || len(b.Password) > 0 {
		auth = fmt.Sprintf("%s:%s@", b.Username, b.Password)
	}

	return fmt.Sprintf("amqp://%s%s:%d/%s",
		auth,
		b.Host,
		b.Port,
		b.Vhost)
}

func (r RedisConf) DSN() string {
	var auth string
	if r.Password != "" {
		auth = fmt.Sprintf(":%s@", r.Password)
	}

	return fmt.Sprintf("redis://%s%s:%d/%d?", auth, r.Host, r.Port, r.DB)
}

func GetConfig() Config {
	if appConfig == (Config{}) {
		loadConfig()
	}

	return appConfig
}

func loadConfig() {
	v := viper.New()

	v.SetConfigName("kumparan")                // name of config file (without extension)
	v.AddConfigPath("/etc/kumpar/")            // path to look for the config file in
	v.AddConfigPath("$HOME/.config/damarseta") // call multiple times to add many search paths
	v.AddConfigPath(".")                       // optionally look for config in the working directory
	err := v.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	err = v.Unmarshal(&appConfig) // Find and read the config file
	if err != nil {               // Handle errors reading the config file
		panic(fmt.Sprintf("unable to decode into struct, %v", err))
	}
}

func init() {
	loadConfig()
}
