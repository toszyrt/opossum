package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/spf13/cobra"
)

func init() {
	var cmd = &cobra.Command{
		Use: "head",
		Run: Safely(Head),
	}
	rootCmd.AddCommand(cmd)
}

func Head(cmd *cobra.Command, args []string) {
	// Connect to the server
	conn, err := net.Dial("tcp", "www.baidu.com:http")
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	// Create the HTTP HEAD request
	request := "HEAD / HTTP/1.1\r\n" +
		"Host: www.baidu.com\r\n" +
		"Connection: close\r\n" +
		"\r\n"

	// Send the request
	_, err = conn.Write([]byte(request))
	if err != nil {
		log.Fatalf("conn.Write error: %v", err)
	}

	// Read the response
	reader := bufio.NewReader(conn)
	output, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("io.ReadAll error: %v", err)
	}
	fmt.Println(string(output))
}
