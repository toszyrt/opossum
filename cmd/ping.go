package cmd

import (
	"bytes"
	"encoding/binary"
	"log"
	"math"
	"net"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/net/ipv4"
)

const (
	Icmp_Req_Cnt = 1
)

var (
	Test_Icmp_Data = []byte{0x23, 0x33}
)

type ICMP struct {
	Type        uint8
	Code        uint8
	Checksum    uint16
	Identifier  uint16
	SequenceNum uint16
}

func init() {
	var cmd = &cobra.Command{
		Use:  "ping {ip_or_url}",
		Run:  Safely(Ping),
		Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	}
	cmd.PersistentFlags().Uint16("cnt", Icmp_Req_Cnt, "cnt of icmp request")
	rootCmd.AddCommand(cmd)
}

func Ping(cmd *cobra.Command, args []string) {
	// 1. 参数解析
	cnt, _ := cmd.PersistentFlags().GetUint16("cnt")
	if cnt == 0 {
		cnt = 1
	}
	// if len(os.Args) == 3 { // [./opossum ping www.google.com]
	// 	ipOrUrl = os.Args[2]
	// }
	ipOrUrl := args[0]
	targetIp := getIp(ipOrUrl)
	if targetIp != ipOrUrl {
		log.Printf("net.LookupIP [%s] ok: %s", ipOrUrl, targetIp)
	}
	// info, err := GetIfaceInfo()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// laddr := net.IPAddr{IP: net.ParseIP(info.Ip)} // laddr不填从0.0.0.0收包
	raddr := net.IPAddr{IP: net.ParseIP(targetIp)}

	// 2. Dial
	conn, err := net.DialIP("ip4:icmp", nil, &raddr) // mac运行需要sudo
	if err != nil {
		log.Fatalf("net.DialIP error: %v", err)
	}
	defer conn.Close()
	// rawConn, err := conn.SyscallConn()
	// if err != nil {
	// 	log.Fatalf("SyscallConn error: %v", err)
	// }
	// rawConn.Control(func(fd uintptr) {
	// 	if err := syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_TTL, 128); err != nil { // 设置ttl
	// 		log.Fatalf("SetsockoptInt error: %v", err)
	// 	}
	// })
	p := ipv4.NewPacketConn(conn)
	if err := p.SetTTL(128); err != nil {
		log.Fatalf("SetTTL error: %v", err)
	}

	// 3. Ping
	var (
		sumOfCost int64
		minCost   int64 = math.MaxInt64
		maxCost   int64 = math.MinInt64
	)
	for i := uint16(0); i < cnt; i++ {
		icmpPacket := createIcmpPack(i)
		tStart := time.Now()
		if _, err := conn.Write(icmpPacket); err != nil {
			log.Fatalf("conn.Write error: %v", err)
		}
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		reply := make([]byte, 1024)
		n, err := conn.Read(reply)
		if err != nil {
			log.Fatalf("conn.Read error: %v", err) // 接收
		}
		cost := time.Since(tStart).Nanoseconds() / 1e6
		ttl := checkIcmpRep(reply, n)
		log.Printf("%d bytes from %s: icmp_seq=%d ttl=%d time=%d ms\n",
			n, targetIp, i, ttl, cost)
		sumOfCost += cost
		minCost = min(minCost, cost)
		maxCost = max(maxCost, cost)
		time.Sleep(time.Second)
	}
	log.Printf("min/avg/max = %d/%d/%d ms", minCost, sumOfCost/int64(cnt), maxCost)
}

func createIcmpPack(seq uint16) []byte {
	var (
		icmp   ICMP
		buffer bytes.Buffer
	)
	// 填充数据包
	icmp.Type = 8 // 8->echo message  0->reply message
	icmp.Code = 0
	icmp.Checksum = 0
	icmp.Identifier = 0
	icmp.SequenceNum = seq
	// 先写数据
	binary.Write(&buffer, binary.BigEndian, icmp)
	buffer.Write([]byte(Test_Icmp_Data))
	icmp.Checksum = CheckSum(buffer.Bytes())
	// 清空, 加上校验和
	buffer.Reset()
	binary.Write(&buffer, binary.BigEndian, icmp)
	buffer.Write([]byte(Test_Icmp_Data)) // 需要有自定义数据, 不然reply包检验出错, conn.Read超时

	return buffer.Bytes()
}

func checkIcmpRep(reply []byte, n int) uint {
	if n < 20+8 { // IP头20字节 ICMP头8字节
		log.Fatalf("reply packet too short: %d bytes", n)
	}
	ipHeader := reply[:20]
	ttl := ipHeader[8] // 提取ttl
	return uint(ttl)
}

func getIp(input string) string {
	if LooksLikeIpv4(input) {
		return input
	}
	ips, err := net.LookupIP(input) // 解析域名的ip
	if err != nil {
		log.Fatalf("net.LookupIP [%s] error: %v", input, err)
	}
	var ipv4s []net.IP
	for _, ip := range ips {
		if ip.To4() != nil {
			ipv4s = append(ipv4s, ip)
		}
	}
	if len(ipv4s) == 0 {
		log.Fatalf("get [%s] ip failed", input)
	}
	target := ipv4s[0].String()
	return target
}

func CheckSum(data []byte) uint16 {
	var (
		sum    uint32
		length int = len(data)
		index  int
	)
	for length > 1 {
		sum += uint32(data[index])<<8 + uint32(data[index+1])
		index += 2
		length -= 2
	}
	if length > 0 {
		sum += uint32(data[index])
	}
	sum += (sum >> 16)

	return uint16(^sum)
}
