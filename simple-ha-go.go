package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var state = "BACKUP"

var listenip = flag.String("listen-ip", "localhost", "the local ip to bind to")
var listenport = flag.Int("listen-port", 29999, "the local port to bind to")
var priority = flag.Int("priority", 0, "the local priority")
var peerip = flag.String("peer-ip", "REQUIRED", "the peer ip to connect to")
var peerport = flag.Int("peer-port", 29999, "the peer port to bind to")
var activescript = flag.String("active-script", "REQUIRED", "the script to launch when switching from backup state to active state")
var backupscript = flag.String("backup-script", "REQUIRED", "the script to launch when switching from active state to backup state")
var verbose = flag.Bool("verbose", false, "be more verbose")
var key = flag.String("key", "OdcejToQuor4", "the shared key between peers")
var retry = flag.Int("retry", 3, "the number of time retrying connecting to the peer when is dead")
var interval = flag.Int("interval", 2, "the interval in second between check to the peer")

func handleRequest(cnx net.Conn) {
	msg := fmt.Sprintf("HA:%s:%s:%d:", state, *key, *priority)
	cnx.Write([]byte(msg))
	cnx.Close()
}

func tcpServer() {
	listenaddr := fmt.Sprintf("%s:%d", *listenip, *listenport)
	listener, err := net.Listen("tcp", listenaddr)
	if err != nil {
		log.Fatal("Error creating socket:", err.Error())
	}
	defer listener.Close()
	log.Printf("[%s]: listening on %s:%d\n", state, *listenip, *listenport)
	for {
		cnx, err := listener.Accept()
		if err != nil {
			log.Fatal("Error accepting on socket:", err.Error())
		}
		go handleRequest(cnx)
	}
}

func checkPeer() int {
	peeraddr := fmt.Sprintf("%s:%d", *peerip, *peerport)
	conn, err := net.Dial("tcp", peeraddr)
	if err != nil {
		if *verbose {
			log.Printf("[%s]: debug, cannot connect to peer\n", state)
		}
		return 0
	} else {
		reader := bufio.NewReader(conn)
		buf := make([]byte, 1024)
		lenbuf, _ := reader.Read(buf)
		if lenbuf > 0 {
			if *verbose {
				log.Printf("[%s]: receveid from peer : %s\n", state, string(buf))
			}
			s := strings.Split(string(buf), ":")
			if len(s) < 4 {
				if *verbose {
					log.Printf("[%s]: debug, malformed response from peer\n", state)
				}
				return 0
			}
			if s[0] != "HA" {
				if *verbose {
					log.Printf("[%s]: debug, malformed response from peer, not found HA\n", state)
				}
				return 0
			}
			if s[2] != *key {
				if *verbose {
					log.Printf("[%s]: debug, received invalid key from peer\n", state)
				}
				return 0
			}
			prio, err := strconv.Atoi(s[3])
			if err != nil {
				if *verbose {
					log.Printf("[%s]: debug, malformed response from peer, priority not numeric\n", state)
				}
				return 0
			}
			return prio
		} else {
			if *verbose {
				log.Printf("[%s]: debug, cannot read from peer\n", state)
			}
			return 0
		}
	}
	return 0
}

func scriptExec(script string) {
	out, err := exec.Command(script).Output()
	if err != nil {
		log.Printf("[%s]: script %s finished with error: %v\n", state, script, err)
	} else {
		log.Printf("[%s]: script %s finished sucessfully\n", state, script)
		log.Printf("%s", out)
	}
}

func main() {

	flag.Parse()
	if *peerip == "REQUIRED" {
		log.Fatal("argument -peer-ip is required")
	}
	if *priority == 0 {
		log.Fatal("argument -priority is required")
	}
	if *activescript == "REQUIRED" {
		log.Fatal("argument -active-script is required")
	}
	if *backupscript == "REQUIRED" {
		log.Fatal("argument -backup-script is required")
	}

	go tcpServer()

	checked := 0

	for {
		peer := checkPeer()
		if state == "BACKUP" {
			if peer == 0 {
				if checked >= *retry {
					log.Printf("[%s]: peer is definitively dead, now becoming ACTIVE", state)
					checked = 0
					state = "ACTIVE"
					scriptExec(*activescript)
				} else {
					checked += 1
					log.Printf("[%s]: peer is dead %d time, retrying\n", state, checked)
				}
			} else if peer < *priority {
				log.Printf("[%s]: peer is alive but with a lower prio, now becoming ACTIVE", state)
				checked = 0
				state = "ACTIVE"
				scriptExec(*activescript)
			} else {
				log.Printf("[%s]: peer is alive with a higher prio, doing nothing\n", state)
				checked = 0
			}
		} else if state == "ACTIVE" {
			if peer == 0 {
				log.Printf("[%s]: peer is dead but we are active, doing nothing\n", state)
			} else if peer > *priority {
				log.Printf("[%s]: peer is alive with a higher prio, becoming BACKUP\n", state)
				checked = 0
				state = "BACKUP"
				scriptExec(*backupscript)
			} else {
				log.Printf("[%s]: peer is alive with a lower prio, doing nothing\n", state)
				checked = 0
			}
		}
		time.Sleep(time.Duration(*interval) * time.Second)
	}
}
