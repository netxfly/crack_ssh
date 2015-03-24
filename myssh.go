package main

import (
	"bytes"
	"fmt"
	"github.com/btcsuite/golangcrypto/ssh"
)

func main() {
	user := "root"
	passwd := "test111"
	host := "www.xsec.io"
	port := "22"
	fmt.Println(Crack(host, port, user, passwd))
}

// crack passwd
func Crack(host, port, user, passwd string) (is_ok bool, result string) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(passwd),
		},
	}
	client, err := ssh.Dial("tcp", host+":"+port, config)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		is_ok = false
		panic("Failed to create session: " + err.Error())
	}
	is_ok = true
	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run("/usr/bin/whoami"); err != nil {
		result = ""
		panic("Failed to run: " + err.Error())
	}
	result = b.String()

	return is_ok, result
}
