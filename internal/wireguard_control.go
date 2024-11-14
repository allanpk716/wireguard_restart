package internal

import (
	"fmt"
	"github.com/WQGroup/logger"
	"golang.org/x/text/encoding/simplifiedchinese"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// 将 GBK 编码的字节转换为 UTF-8 字符串
func convertGBK(bytes []byte) string {
	decoder := simplifiedchinese.GBK.NewDecoder()
	result, err := decoder.Bytes(bytes)
	if err != nil {
		return fmt.Sprintf("解码错误: %v, 原始输出: %s", err, string(bytes))
	}
	return string(result)
}

// 执行命令并处理输出
func executeCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return convertGBK(output), err
}

func RestartWireGuardTunnel(tunnelName string) error {

	// 隧道名称
	logger.Infoln("隧道名称:", tunnelName)

	programFiles := os.Getenv("ProgramFiles")
	configPath1 := filepath.Join(programFiles, "WireGuard", "Data", "Configurations", tunnelName+configExtension1)
	configPath2 := filepath.Join(programFiles, "WireGuard", "Data", "Configurations", tunnelName+configExtension2)
	// 检查配置文件是否存在
	var configPath string
	if _, err := os.Stat(configPath1); err != nil {
		if _, err := os.Stat(configPath2); err != nil {
			return fmt.Errorf("配置文件不存在: %s 或 %s, 错误: %w", configPath1, configPath2, err)
		} else {
			configPath = configPath2
		}
	} else {
		configPath = configPath1
	}
	logger.Infoln("配置文件路径:", configPath)

	// 卸载服务（使用隧道名称）
	logger.Infoln("卸载隧道服务:", tunnelName)
	output, err := executeCommand(wireGuardExe, "/uninstalltunnelservice", tunnelName)
	if err != nil {
		logger.Errorln("卸载命令输出:", string(output))
		// 继续执行，因为服务可能已经不存在
	}
	logger.Infoln("等待 2s, 等待服务完全停止...")
	// 等待服务完全停止
	time.Sleep(2 * time.Second)

	// 查询服务状态
	logger.Infoln("查询服务状态:", wireGuardService+tunnelName)
	statusCmd := exec.Command("sc.exe", "query", wireGuardService+tunnelName)
	outputBytes, err := statusCmd.CombinedOutput()
	if err != nil {
		logger.Infoln("服务停止成功")
	} else {
		logger.Errorln("服务停止 Err")
	}

	// 重新安装服务（使用配置文件路径）
	logger.Infoln("安装隧道服务，使用配置文件:", configPath)
	output, err = executeCommand(wireGuardExe, "/installtunnelservice", configPath)
	if err != nil {
		return fmt.Errorf("安装隧道服务失败: %s, 错误: %w", string(output), err)
	}

	logger.Infoln("等待 2s, 等待服务启动...")
	time.Sleep(2 * time.Second)

	// 查询服务状态
	logger.Infoln("查询服务状态:", wireGuardService+tunnelName)
	statusCmd = exec.Command("sc.exe", "query", wireGuardService+tunnelName)
	outputBytes, err = statusCmd.CombinedOutput()
	if err != nil {
		logger.Errorln("没有查询到服务 Err:", string(outputBytes))
		return err
	} else {
		logger.Infoln("查询服务状态:", string(outputBytes))
		if strings.Contains(string(outputBytes), "RUNNING") == true {
			logger.Infoln("服务启动成功")
		} else {
			logger.Errorln("服务启动失败")
			return fmt.Errorf("服务启动失败")
		}
	}

	return nil
}

const (
	configExtension1 = ".conf.dpapi"
	configExtension2 = ".conf"
	wireGuardExe     = "wireguard.exe"
	wireGuardService = "WireGuardTunnel$"
)
