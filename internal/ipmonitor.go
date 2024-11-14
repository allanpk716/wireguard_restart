package internal

import (
	"context"
	"github.com/WQGroup/logger"
	"net"
	"time"
)

type IPMonitor struct {
	config      *Config
	ctx         context.Context
	cancel      context.CancelFunc
	done        chan struct{}
	currentIPv4 net.IP
	currentIPv6 net.IP
}

func NewIPMonitor(config *Config) *IPMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &IPMonitor{
		config: config,
		ctx:    ctx,
		cancel: cancel,
		done:   make(chan struct{}),
	}
}

func (m *IPMonitor) Start() {
	ticker := time.NewTicker(time.Duration(m.config.Interval) * time.Minute)
	defer ticker.Stop()

	// 先检测一次 IP
	m.checkIP()

	for {
		select {
		case <-m.ctx.Done():
			close(m.done)
			return
		case <-ticker.C:
			m.checkIP()
		}
	}
}

func (m *IPMonitor) checkIP() {

	logger.Infoln("--------------------")
	logger.Infoln("Checking IP...")
	// 检测提供的域名是否能够解析到 IP，并且根据配置的 IP 版本，检测是否有变化
	ips, err := net.LookupIP(m.config.Domain)
	if err != nil {
		logger.Errorln("Error looking up IP:", err)
		return
	}

	// 打印解析到的 IP
	if m.config.IPVersion == "ipv4" || m.config.IPVersion == "both" {
		for _, ip := range ips {
			if ip.To4() != nil {
				logger.Infoln("IPv4:", ip)
			}
		}
	} else if m.config.IPVersion == "ipv6" {
		for _, ip := range ips {
			if ip.To16() != nil {
				logger.Infoln("IPv6:", ip)
			}
		}
	} else {
		for _, ip := range ips {
			logger.Infoln("IP:", ip)
		}
	}

	ipv4Changed, ipv6Changed := false, false
	for _, ip := range ips {
		if m.config.IPVersion == "ipv4" || m.config.IPVersion == "both" {
			ipv4 := ip.To4()
			if ipv4 != nil && !ipv4.Equal(m.currentIPv4) {
				logger.Infoln("IPv4 from:", m.currentIPv4, "to:", ipv4)
				m.currentIPv4 = ipv4
				ipv4Changed = true
			}
		}
		if m.config.IPVersion == "ipv6" || m.config.IPVersion == "both" {
			ipv6 := ip.To16()
			if ipv6 != nil && !ipv6.Equal(m.currentIPv6) {
				logger.Infoln("IPv6 from:", m.currentIPv6, "to:", ipv6)
				m.currentIPv6 = ipv6
				ipv6Changed = true
			}
		}
	}
	// 如果有变化，执行重启 WireGuard 的命令
	if ipv4Changed || ipv6Changed {
		logger.Infoln("IP changed, restarting WireGuard...")
		err = RestartWireGuardTunnel(m.config.TunnelName)
		if err != nil {
			logger.Errorln("Error restarting WireGuard:", err)
			// 那么就需要清空当前的 IP，下次再检测
			m.currentIPv4 = nil
			m.currentIPv6 = nil
			logger.Errorln("WireGuard 重启失败，清空当前 IP，下次再检测")
			return
		}
		logger.Infoln("WireGuard 重启成功")
	}
}

func (m *IPMonitor) Done() <-chan struct{} {
	return m.done
}
