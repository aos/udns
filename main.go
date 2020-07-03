package main

import (
	"bytes"
	"errors"
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
				log.Printf("zone file has been modified %s", fileInfo.Name())
				fileInfo = checkFile
			}
		}
	}
}

func run(args []string, stdin io.Reader) error {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		port     = flags.String("port", "53", "UDP port to listen on for DNS")
		server   = flags.String("forward-server", "1.1.1.1:53", "forward DNS server")
		zonefile = flags.String("zone file", "default.zone", "zone file name")
	)
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	if *zonefile == "" {
		return errors.New("Must specify a zone file")
	}

	go monitorZonefile(*zonefile)

	rrs, err := parseZonefile(*zonefile)
	if err != nil {
		return fmt.Errorf("Error parsing zone file: %s", err)
	}

	dns.HandleFunc(".", func(w dns.ResponseWriter, m *dns.Msg) {
		// do something with dns requestion
		// like serve back the matched record(s)
	})

	return nil
}

func parseZonefile(zonefile string) ([]dns.RR, error) {
	data, err := ioutil.ReadFile(*zonefile)
	if err != nil {
		return fmt.Errorf("Error reading zone file: %s", err)
	}

	rrs := []dns.RR{}

	zp := dns.NewZoneParser(bytes.NewReader(data), "", *zonefile)
	if zp.Err() != nil {
		return nil, zp.Err()
	}

	for rr, ok := zp.Next(); ok; rr, ok = zp.Next() {
		rrs = append(rrs, rr)
	}

	return rrs, nil
}

func main() {
	if err := run(os.Args, os.Stdin); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
