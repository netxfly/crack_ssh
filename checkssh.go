package main

import (
	"bufio"
	_ "bytes"
	"fmt"
	"github.com/btcsuite/golangcrypto/ssh"
	"os"
	"strings"
)

// read lime from file and Scan
func Scan(path string) {
	inFile, _ := os.Open(path)
	defer inFile.Close()
	scanner := bufio.NewScanner(inFile)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		ip_port := scanner.Text()
		fmt.Println("start to check: ", ip_port)
		slice_ip := strings.Split(ip_port, ":")
		host := slice_ip[0]
		port := slice_ip[1]
		user := "root"
		passwd := "aisin_gioro"
		ret := Crack(host, port, user, passwd)
		if ret {
			fmt.Printf("%s is not company's ssh\n", host)
		}
	}
}

func Usage(cmd string) {
	fmt.Println(strings.Repeat("-", 100))
	fmt.Println("Ssh version Check by haifeng11")
	fmt.Println("Usage:")
	fmt.Printf("%s iplist_file\n", cmd)
	fmt.Println(strings.Repeat("-", 100))
}

// main function
func main() {
	// fmt.Println(os.Args)
	if len(os.Args) != 2 {
		Usage(os.Args[0])

	} else {
		filepath := os.Args[1]
		Scan(filepath)
	}
}

// crack passwd
func Crack(host, port, user, passwd string) (is_ok bool) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(passwd),
		},
	}
	_, err := ssh.Dial("tcp", host+":"+port, config)
	if err != nil {
		is_ok = false
		// panic("Failed to dial: " + err.Error())
	}
	is_ok = true
	return is_ok
}
