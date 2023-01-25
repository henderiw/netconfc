package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/henderiw/netconfc/netconf"
	"golang.org/x/crypto/ssh"
)

func main() {
	var address string
	var username string
	var password string
	var keyfile string
	var get string
	var period int

	flag.StringVar(&address, "address", "172.20.20.2:830", "Address")
	flag.StringVar(&username, "user", "admin", "Username")
	flag.StringVar(&password, "pass", "admin", "Password")
	flag.StringVar(&get, "get", "/port", "Get")
	flag.IntVar(&period, "period", 3, "Period")
	flag.Parse()

	var client netconf.Client
	var err error
	if len(keyfile) == 0 {
		client, err = netconf.DialSSHWithPassword(address, username, password, ssh.InsecureIgnoreHostKey())
	} else {
		var key []byte
		key, err = os.ReadFile(keyfile)
		if err == nil {
			var signer ssh.Signer
			signer, err = ssh.ParsePrivateKey(key)
			if err == nil {
				client, err = netconf.DialSSHWithPublicKey(address, username, signer, ssh.InsecureIgnoreHostKey())
			}
		}
	}
	if err != nil {
		log.Println(err.Error())
		os.Exit(2)
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		log.Println(err.Error())
		os.Exit(3)
	}
	request := netconf.Get{Filter: &netconf.Filter{Type: "xpath", Select: get}}
	response := netconf.RPCReplyData{}
	for len(get) > 0 {
		if err := session.Call(&request, &response); err != nil {
			log.Println(err)
		}

		fmt.Printf("%s\n\n", response.Data.InnerXML)

		select {
		case <-time.After(time.Duration(period) * time.Second):
		}
	}

	session.Close()
	client.Close()

}
