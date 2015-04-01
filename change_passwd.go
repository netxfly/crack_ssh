package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/btcsuite/golangcrypto/ssh"
	"github.com/golang/glog"
	"io"
	"os"
	"strings"
	"time"
)

// help function
func Usage(cmd string) {
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("Usage:")
	fmt.Printf("%s iplist \n", cmd)
	fmt.Println(strings.Repeat("-", 50))
}

// read lime from file and Scan
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

// Scan function
func Scan(slice_iplist []string) {
	for _, iplist := range slice_iplist {
		fmt.Println(iplist)
		t := strings.Split(iplist, ":")
		host := t[0]
		port := t[1]
		user := t[2]
		pass := t[3]
		newpass := createpasswd()[:10]
		ok, result := Crack(host, port, user, pass, newpass)
		glog.Errorf("%v:%v %v %v %v %v %v", host, port, user, pass, newpass, ok, result)
	}
}

// main function
func main() {
	if len(os.Args) != 2 {
		Usage(os.Args[0])
	} else {
		iplist := os.Args[1]
		Usage(os.Args[0])
		Scan(Prepare(iplist))

	}

}

// crack passwd
func Crack(host, port, user, passwd, newpasswd string) (is_ok bool, result string) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(passwd),
		},
	}
	client, err := ssh.Dial("tcp", host+":"+port, config)
	if err != nil {
		result = "Failed to dial: " + err.Error()
		glog.Errorf("%s\n", result)
		return is_ok, result
	}

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		result = "Failed to create session: " + err.Error()
		glog.Errorf("%s\n", result)
		return is_ok, result
	}

	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var b bytes.Buffer
	session.Stdout = &b
	cmd := fmt.Sprintf("/bin/echo \"%s:%s\" | /usr/sbin/chpasswd", user, newpasswd)
	// fmt.Println(cmd)
	if err := session.Run(cmd); err != nil {
		result = "Failed to run: " + err.Error()
		glog.Errorf("%s\n", result)
	} else {
		is_ok = true
		result = b.String()
	}

	return is_ok, result
}

// create random passwd
func createpasswd() string {
	t := time.Now()
	h := md5.New()
	io.WriteString(h, "sina sec")
	io.WriteString(h, t.String())
	passwd := fmt.Sprintf("%x", h.Sum(nil))
	return passwd
}
