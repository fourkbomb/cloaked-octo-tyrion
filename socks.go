/*
   Copyright 2014 Simon Shields

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
package main

import (
	"fmt"
	"net"
	"bufio"
	"os"
	"bytes"
	"strings"
	"encoding/binary"
	"strconv"
)

func genRequest(dns string, port uint16) []byte {
	reader := bytes.NewBufferString(dns)
	bts := reader.Bytes()
	res := []byte{0x05, 0x01, 0x00, 0x03, byte(len(bts))}
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, port)
	res = append(res, bts...)
	return append(res, b...) // TODO ports other than 80
}

func sendStuff(conn net.Conn) {
	stdin := bufio.NewReader(os.Stdin)
	for true {
		text, _ := stdin.ReadString('\n')
		fmt.Fprintf(conn, strings.TrimSpace(text) + "\r\n")
	}
}

func main() {
	if len(os.Args) < 3 {
		println("Usage: " + os.Args[0] + " <domain name> <port> [proxyhost [proxyport]]")
		println("Proxy host defaults to localhost, and proxy port defaults to 1080")
		os.Exit(1)
	}
	host := "localhost"
	portnum := "1080"
	if len(os.Args) >= 4 {
		host = os.Args[3]
		if (len(os.Args) >= 5) {
			portnum = os.Args[4]
		}
	}
	conn, err := net.Dial("tcp", host + ":" + portnum)
	if err != nil {
		fmt.Println("Damn")
		fmt.Println(err)
		os.Exit(1)
	}
	conn.Write([]byte{0x05, 0x01, 0x00}) // SOCKS 5, no auth
	response := make([]byte, 2)
	conn.Read(response)
	if response[0] != 0x05 {
		fmt.Println("Not SOCKS 5")
		fmt.Println("It's actually SOCKS %d", response[0])
		os.Exit(2)
	}

	if response[1] == 0xFF {
		fmt.Println("No acceptable auth methods")
		os.Exit(2)
	}

	// try and open a TCP connection to google or something
	port, err := strconv.ParseInt(os.Args[2], 10, 16)
	if err != nil {
		println("Invalid port " + os.Args[2])
		os.Exit(3)
	}
	toSend := genRequest(os.Args[1], uint16(port))
	conn.Write(toSend)
	fmt.Println("Trying " + os.Args[1] + ":" + os.Args[2] + "...")
	response = make([]byte, 4)
	conn.Read(response)
	msg := "Unknown error"
	switch response[1] {
	case 0x00:
		msg = "Success"
	case 0x01:
		msg = "General SOCKS Server Failure"
	case 0x02:
		msg = "Connection not allowed by ruleset"
	case 0x03:
		msg = "Network unreachable"
	case 0x04:
		msg = "Host unreachable"
	case 0x05:
		msg = "Connection refused"
	case 0x06:
		msg = "TTL expired"
	case 0x07:
		msg = "Command not supported"
	case 0x08:
		msg = "Address type not supported"
	}
	if response[1] != 0 {
		fmt.Println("Error:",msg)
		os.Exit(1)
	}
	println("Connected!")
	// now do some reading and writing
	go sendStuff(conn)
	reader := bufio.NewReader(conn)
	for true {
		line, err := reader.ReadString('\n')
		if err != nil {
			println("Lost connection to the SOCKS server")
			println(err)
			os.Exit(3)
		}
		line = strings.TrimSpace(line)
		println(line)
	}

}
