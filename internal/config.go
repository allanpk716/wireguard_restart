package internal

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Domain                   string `yaml:"domain"`                      // WireGuard 服务器的域名
	TunnelName               string `yaml:"tunnel_name"`                 // WireGuard 隧道的名称
	Interval                 int    `yaml:"interval"`                    // 检查 IP 地址是否变化的时间间隔，单位为分钟
	IPVersion                string `yaml:"ip_version"`                  // IP 地址的版本，可以是 "ipv4" 或 "ipv6" 或 "both"
	CheckInternalServiceHost string `yaml:"check_internal_service_host"` // 用于检测，连入 WireGuard 内网后，内网的一个端口服务是否活着，因为某些情况下，WireGuard 服务器 IP 没有变化，但是依然是链接故障了（比如 WireGuard 显示最后一次握手时间超过 N 分钟）
	CheckInternalServicePort string `yaml:"check_internal_service_port"` // 用于检测，连入 WireGuard 内网后，内网的一个端口服务是否活着，因为某些情况下，WireGuard 服务器 IP 没有变化，但是依然是链接故障了（比如 WireGuard 显示最后一次握手时间超过 N 分钟）
}

func loadConfig(file string) (*Config, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
