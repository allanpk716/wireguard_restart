package internal

import (
	"context"
	"net"
	"time"

	"github.com/WQGroup/logger"
	gpkg "github.com/allanpk716/go-protocol-detector/pkg"
)

type IPMonitor struct {
	config      *Config
	ctx         context.Context
	cancel      context.CancelFunc
	done        chan struct{}
	currentIPv4 net.IP
	currentIPv6 net.IP

	detector *gpkg.Detector
}

func NewIPMonitor(config *Config) *IPMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &IPMonitor{
		config:   config,
		ctx:      ctx,
		cancel:   cancel,
		done:     make(chan struct{}),
		detector: gpkg.NewDetector(3 * time.Second),
	}
}

// 复合判断是否需要重启 WireGuard
func (m *IPMonitor) jugde() {

	// 因为某些情况下，WireGuard 服务器 IP 没有变化，但是依然是链接故障了（比如 WireGuard 显示最后一次握手时间超过 N 分钟）
	if m.checkIP() == true {

		logger.Infoln("IP changed, restarting WireGuard...")

		m.restartWireGuard()
		return
	}
	// 优先检测 IP 地址是否变化，如果没有变化，再检测内部服务端口是否正常
	if m.config.CheckInternalServicePort != "" && m.config.CheckInternalServiceHost != "" {
		logger.Infoln("Checking internal service port...")
		logger.Infoln("Host:", m.config.CheckInternalServiceHost, "Port:", m.config.CheckInternalServicePort)
		// 检测内部服务端口是否正常
		if m.checkInternalServicePort() == true {
			logger.Infoln("Internal service port is not working, restarting WireGuard...")
			m.restartWireGuard()
			return
		}
	}
}

// Start 开始监控 IP 地址是否变化
func (m *IPMonitor) Start() {
	ticker := time.NewTicker(time.Duration(m.config.Interval) * time.Minute)
	defer ticker.Stop()

	// 先检测一次
	m.jugde()

	for {
		select {
		case <-m.ctx.Done():
			close(m.done)
			return
		case <-ticker.C:
			m.jugde()
		}
	}
}

// 检测内部服务端口是否正常
func (m *IPMonitor) checkInternalServicePort() bool {

	// 检测内部服务端口是否正常
	err := m.detector.CommonPortCheck(m.config.CheckInternalServiceHost, m.config.CheckInternalServicePort)
	if err != nil {
		logger.Errorln("Error checking internal service port:", err)
		return true
	}

	return false
}

// 检测提供的域名是否能够解析到 IP，并且根据配置的 IP 版本，检测是否有变化
func (m *IPMonitor) checkIP() bool {

	logger.Infoln("--------------------")
	logger.Infoln("Checking IP...")
	// 检测提供的域名是否能够解析到 IP，并且根据配置的 IP 版本，检测是否有变化
	ips, err := net.LookupIP(m.config.Domain)
	if err != nil {
		logger.Errorln("Error looking up IP:", err)
		return false
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
		return true
	}

	return false
}

func (m *IPMonitor) restartWireGuard() {

	err := RestartWireGuardTunnel(m.config.TunnelName)
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

func (m *IPMonitor) Done() <-chan struct{} {
	return m.done
}
