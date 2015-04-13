package main

import (
	"bufio"
	"fmt"
	"github.com/MG-RAST/golib/mgo"
	"os"
	"runtime"
	"strings"
	"time"
)

// Host Info define
type HostInfo struct {
	Host    string
	Port    string
	Dbs     []string
	Is_weak bool
}

// help function
func Usage(cmd string) {
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("Redis Scanner by hartnett x@xsec.io")
	fmt.Println("Usage:")
	fmt.Printf("%s iplist \n", cmd)
	fmt.Println(strings.Repeat("-", 50))
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

// Connect to mongodb
func TestConnect(host_info HostInfo, chan_host_info chan HostInfo) {
	host := host_info.Host
	port := host_info.Port
	is_weak := host_info.Is_weak
	url := fmt.Sprintf("%s:%s", host, port)
	session, err := mgo.DialWithTimeout(url, 2*time.Second)
	if err == nil {
		dbs, err := session.DatabaseNames()
		if err == nil {
			is_weak = true
			host_info.Dbs = dbs
		}
	}
	host_info.Is_weak = is_weak
	chan_host_info <- host_info
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
		host_info := HostInfo{host, port, []string{}, false}

		go TestConnect(host_info, chan_scan_result)
		for runtime.NumGoroutine() > runtime.NumCPU()*200 {
			time.Sleep(10 * time.Microsecond)
		}

	}

	go func() {
		for i := 0; i < cap(chan_scan_result); i++ {
			select {
			case r := <-chan_scan_result:
				if r.Is_weak {
					fmt.Printf("%s:%s is vulnerability, DBs:%s\n", r.Host, r.Port, r.Dbs)
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
