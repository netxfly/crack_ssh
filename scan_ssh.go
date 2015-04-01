package main

import (
	"bufio"
	"fmt"
	"github.com/btcsuite/golangcrypto/ssh"
	"os"
	"runtime"
	"strings"
)

type HostInfo struct {
	host    string
	port    string
	user    string
	pass    string
	is_weak bool
}

var chan_scan_result chan HostInfo //= make(chan HostInfo, 1000000)

// help function
func Usage(cmd string) {
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("Usage:")
	fmt.Printf("%s iplist userdic passwddic\n", cmd)
	fmt.Println(strings.Repeat("-", 50))
}

// read lime from file and Scan
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

// Scan function
func Scan(slice_iplist, slice_user, slice_pass []string) {
	for _, host_port := range slice_iplist {
		fmt.Println(host_port)
		t := strings.Split(host_port, ":")
		host := t[0]
		port := t[1]
		n := len(slice_user) * len(slice_pass) * len(slice_iplist)
		chan_scan_result = make(chan HostInfo, n)

		for _, user := range slice_user {
			for _, passwd := range slice_pass {

				host_info := HostInfo{}
				host_info.host = host
				host_info.port = port
				host_info.user = user
				host_info.pass = passwd
				host_info.is_weak = false

				go Crack(host_info)
			}
		}
	}

}

// main function
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if len(os.Args) != 4 {
		Usage(os.Args[0])
	} else {
		iplist := os.Args[1]
		user_dict := os.Args[2]
		pass_dict := os.Args[3]
		// fmt.Println(Prepare(iplist, user_dict, pass_dict))
		Scan(Prepare(iplist, user_dict, pass_dict))
		// fmt.Println(len(chan_scan_result))
		fmt.Println(runtime.NumGoroutine())

		for h := range chan_scan_result {
			fmt.Println(runtime.NumGoroutine())
			if h.is_weak {
				fmt.Println(h)
				fmt.Println("---------------------------")
			}
		}
	}

}

// crack passwd
func Crack(host_info HostInfo) {
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
	}
	client, err := ssh.Dial("tcp", host+":"+port, config)
	if err != nil {
		chan_scan_result <- host_info
		runtime.Goexit()
		// panic("Failed to dial: " + err.Error())
	}

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	defer session.Close()
	if err != nil {
		chan_scan_result <- host_info
		runtime.Goexit()
	}

	is_ok = true
	host_info.is_weak = is_ok
	chan_scan_result <- host_info

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	// var b bytes.Buffer
	// session.Stdout = &b
	// if err := session.Run("ifconfig | grep 'inet addr'"); err != nil {
	// 	result = ""
	// 	panic("Failed to run: " + err.Error())
	// }
	// result = b.String()

	// return is_ok, result
}
