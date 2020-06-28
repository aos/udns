package main

import (
	"flag"
	"log"
	"os"
	"time"
)

var (
	port     = flag.String("port", "53", "UDP port to listen on for DNS")
	server   = flag.String("forward-server", "1.1.1.1:53", "forward DNS server")
	zonefile = flag.String("zonefile", "", "zonefile location")
)

func monitorZonefile(zonefile string) {
	fileInfo, err := os.Stat(zonefile)
	if err != nil {
		log.Fatalf("Could not stat file: %s", err)
	}

	t := time.NewTicker(time.Second * 30)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			checkFile, err := os.Stat(zonefile)
			if err != nil {
				log.Fatalf("Could not stat file: %s", err)
			}

			if fileInfo.ModTime() != checkFile.ModTime() {
				log.Printf("zonefile has been modified %s", fileInfo.Name())
				fileInfo = checkFile
			}
		}
	}
}

func main() {
	// flag.Parse()
	// go monitorZonefile()
}
