package cmd

import (
	"bytes"
	"errors"
	"log"
	"net"
	"os/exec"
	"strings"
)

type NetInfo struct {
	Name string
	Mac  string
	Ip   string
	Mask string
}

// GetLocalIp 获取本地Ip, 有外网时使用
func GetLocalIp() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Printf("net.Dial error: %v", err)
		return ""
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// getLocalIface 获取默认网卡名
func getLocalIface() (string, error) {
	cmd := exec.Command("sh", "-c", "netstat -rn | grep -E 'default|UG[ \t]'")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}

	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 && (fields[0] == "default" || fields[0] == "0.0.0.0") {
			return fields[len(fields)-1], nil
		}
	}
	return "", errors.New("defalut iface not found")
}

func GatherNetInfo() (*NetInfo, error) {
	ifaceName, err := getLocalIface()
	if err != nil {
		return nil, err
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatalf("net.Interfaces error: %v", err)
	}
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			log.Fatalf("iface.Addrs error: %v", err)
		}
		for _, addr := range addrs {
			var (
				ip    net.IP
				ipNet *net.IPNet
			)
			switch v := addr.(type) {
			case *net.IPNet: // ipv4
				ip = v.IP
				ipNet = v
			case *net.IPAddr: // ipv6
				ip = v.IP
				ipNet = &net.IPNet{IP: v.IP, Mask: v.IP.DefaultMask()}
			}
			if ip == nil || ipNet == nil {
				continue
			}
			if ip.To4() == nil { // ipv6哒咩
				continue
			}
			if iface.Name == ifaceName {
				info := &NetInfo{
					Name: iface.Name,
					Mac:  iface.HardwareAddr.String(),
					Ip:   ip.String(),
					Mask: net.IP(ipNet.Mask).String(),
				}
				return info, nil
			}
		}
	}
	return nil, errors.New("fail to get net info")
}

func GetGwAndBc(ipStr, maskStr string) (string, string) {
	// 解析ip
	ip := net.ParseIP(ipStr).To4()
	if ip == nil {
		log.Fatal("parse ip failed")
	}

	// mask
	mask := net.IPMask(net.ParseIP(maskStr).To4())
	if mask == nil {
		log.Fatal("parse mask failed")
	}

	network := ip.Mask(mask)
	gateway := net.IP(make([]byte, len(network)))
	copy(gateway, network)
	gateway[len(gateway)-1]++

	broadcast := make(net.IP, len(ip))
	for i := 0; i < len(ip); i++ {
		broadcast[i] = ip[i] | ^mask[i]
	}

	return gateway.String(), broadcast.String()
}
