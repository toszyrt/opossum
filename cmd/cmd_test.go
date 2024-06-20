package cmd

import (
	"fmt"
	"log"
	"net"
	"testing"
)

func TestGatherNetInfo(t *testing.T) {
	info, err := GatherNetInfo()
	if err != nil {
		log.Fatal(err)
	}
	msg := fmt.Sprintf(`
Name: %s
 Mac: %s
  Ip: %s
Mask: %s`,
		info.Name, info.Mac, info.Ip, info.Mask)
	fmt.Println(msg)
	gateway, broadcast := GetGwAndBc(info.Ip, info.Mask)
	fmt.Println()
	fmt.Printf("gateway: %s\nbroadcast: %s\n", gateway, broadcast)
}

func TestRandom(t *testing.T) {
	for i := 0; i < 10; i++ {
		mac, ip := formatMac(getRandomMac()), net.IP(getRandomIp())
		fmt.Printf("%s, %s\n", mac, ip)
	}
}
