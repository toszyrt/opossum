package cmd

import (
	"bytes"
	"fmt"
	"log"
	"net"

	"github.com/spf13/cobra"
)

const (
	MacOfH610m = "C8-7F-54-5D-08-53"
	Wol_Port   = 9
)

func init() {
	var cmd = &cobra.Command{
		Use:   "wol",
		Short: "wake on lan",
		Run:   Safely(Wol),
	}
	cmd.PersistentFlags().String("mac", MacOfH610m, "mac of board nic") // 主板网卡mac
	cmd.PersistentFlags().String("ip", "", "ip of board nic")
	rootCmd.AddCommand(cmd)
}

func Wol(cmd *cobra.Command, args []string) {
	tgtMac, _ := cmd.PersistentFlags().GetString("mac")
	tgtIp, _ := cmd.PersistentFlags().GetString("ip")
	macAddr, err := net.ParseMAC(tgtMac)
	if err != nil {
		log.Fatalf("net.ParseMAC error: %v", err)
	}
	// 生成魔术包
	header := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	body := bytes.Repeat(macAddr, 16)
	magicPacket := append(header, body...)

	// 使用udp发送魔术包
	if len(tgtIp) == 0 { // 有目标ip时直接单播, 可以跨越子网
		tgtIp = "255.255.255.255" // 整个局域网广播 (一般路由器不转发, 会限制在本地子网)
		info, err := GatherNetInfo()
		if err != nil {
			log.Println(err)
		} else {
			_, broadcast := GetGwAndBc(info.Ip, info.Mask)
			tgtIp = broadcast // 本地子网广播
		}
	}
	address := fmt.Sprintf("%s:%d", tgtIp, Wol_Port)
	log.Printf("net.Dial %s", address)
	conn, err := net.Dial("udp", address)
	if err != nil {
		log.Fatalf("net.Dial error: %v", err)
	}
	defer conn.Close()

	_, err = conn.Write(magicPacket)
	if err != nil {
		log.Fatalf("conn.Write error: %v", err)
	}

	fmt.Println("WOL packet sent successfully")
}
