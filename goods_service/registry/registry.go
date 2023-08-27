package registry

import "github.com/hashicorp/consul/api"

// Register 自定义一个注册中心的抽象
type Register interface {
	// RegisterService 注册
	RegisterService(serviceName string, ip string, port int, tags []string) error
	// ListService 服务发现
	ListService(serviceName string) (map[string]*api.AgentService, error)
	// Deregister 注销
	Deregister(serviceID string) error
}
