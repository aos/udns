package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/miekg/dns"
)

func monitorZonefile(zp *dns.ZoneParser) {
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
	if err := run(os.Args, os.Stdin); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader) error {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		port     = flags.String("port", "53", "UDP port to listen on for DNS")
		server   = flags.String("forward-server", "1.1.1.1:53", "forward DNS server")
		zonefile = flags.String("zonefile", "", "zonefile location")
	)
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	if *zonefile == "" {
		return fmt.Errorf("Must specify a zonefile")
	}

	data, err := ioutil.ReadFile(*zonefile)
	if err != nil {
		return fmt.Errorf("Error reading file: %s", err)
	}

	zp := dns.NewZoneParser(bytes.NewReader(data), "", *zonefile)
	//go monitorZonefile(zp)

	return nil
}
