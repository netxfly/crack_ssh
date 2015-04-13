package main

import (
	"bufio"
	"fmt"
	"github.com/beego/redigo/redis"
	"os"
	"runtime"
	"strings"
	"time"
)

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
	fmt.Println("Redis Scanner by hartnett x@xsec.io")
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
	c, err := redis.Dial("tcp", host+":"+port)
	// _, err := redis.DialTimeout("tcp", host+":"+port, 2*time.Second, 2*time.Second, 2*time.Second)
	if err == nil {
		s, err := redis.String(c.Do("ping"))
		if err == nil {
			is_vul = true
			reply = s
		}
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
					fmt.Printf("%s:%s is vulnerability, ping's reply: %s\n", r.host, r.port, r.reply)
				}
			case <-time.After(3 * time.Second):
				// fmt.Println("timeout")
				break

			}
			done <- true

		}
	}()

	for i := 0; i < cap(done); i++ {
		<-done
	}

}
