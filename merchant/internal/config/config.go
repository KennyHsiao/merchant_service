package config

import (
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	Host           string
	Server         string
	Domain         string
	FrontEndDomain string
	rest.RestConf
	service.ServiceConf
	RpcService     zrpc.RpcClientConf
	TransactionRpc zrpc.RpcClientConf
	Mysql          struct {
		Host       string
		Port       int
		DBName     string
		UserName   string
		Password   string
		DebugLevel string
	}
	RedisCache struct {
		RedisSentinelNode string
		RedisMasterName   string
		RedisDB           int
	}
	ApiKey struct {
		PublicKey string
		PayKey    string
		ProxyKey  string
	}
	ResourceHost string
	Target       string
}
