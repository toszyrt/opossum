package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/spf13/cobra"
)

func init() {
	var cmd = &cobra.Command{
		Use: "arp-listen",
		Run: Safely(ArpListen),
	}
	rootCmd.AddCommand(cmd)
}

func ArpListen(cmd *cobra.Command, args []string) {
	info, err := GatherNetInfo()
	if err != nil {
		log.Fatal(err)
	}
	ifaceName := info.Name

	handle, err := pcap.OpenLive(ifaceName, 65536, true, pcap.BlockForever)
	if err != nil {
		log.Fatalf("pcap.OpenLive failed: %v", err)
	}
	defer handle.Close()
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		arpLayer := packet.Layer(layers.LayerTypeARP) // 过滤arp layer
		if arpLayer == nil {
			continue
		}
		arp, _ := arpLayer.(*layers.ARP)
		ArpPrint(arp)

		// 收到面向本机的arp请求, 尝试回包
		if arp.Operation == layers.ARPRequest &&
			net.IP(arp.DstProtAddress).String() == info.Ip {
			ArpReply(handle)
		}
	}
}

// ArpReply 回复请求 (和系统自带干扰? 暂不实现)
func ArpReply(handle *pcap.Handle) {

}

// ArpPrint 格式化打印arp包
func ArpPrint(arp *layers.ARP) {
	if arp == nil {
		return
	}
	arpWrapper := &ARPWrapper{arp}
	j, err := json.MarshalIndent(arpWrapper, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(j))
}

// ARPWrapper arp包装器
type ARPWrapper struct {
	*layers.ARP
}

func (w *ARPWrapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		AddrType          layers.LinkType     `json:"AddrType"`
		Protocol          layers.EthernetType `json:"Protocol"`
		HwAddressSize     uint8               `json:"HwAddressSize"`
		ProtAddressSize   uint8               `json:"ProtAddressSize"`
		Operation         uint16              `json:"Operation"`
		SourceHwAddress   string              `json:"SourceHwAddress"`
		SourceProtAddress net.IP              `json:"SourceProtAddress"`
		DstHwAddress      string              `json:"DstHwAddress"`
		DstProtAddress    net.IP              `json:"DstProtAddress"`
	}{
		AddrType:          w.AddrType,
		Protocol:          w.Protocol,
		HwAddressSize:     w.HwAddressSize,
		ProtAddressSize:   w.ProtAddressSize,
		Operation:         w.Operation,
		SourceHwAddress:   formatMac(w.SourceHwAddress),
		SourceProtAddress: net.IP(w.SourceProtAddress),
		DstHwAddress:      formatMac(w.DstHwAddress),
		DstProtAddress:    net.IP(w.DstProtAddress),
	})
}

func formatMac(mac []byte) string {
	// 有前导0的2为16进制数
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
}
