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
	"sync"
	"time"

	"github.com/miekg/dns"
)

// Zone contains all the information necessary to serve DNS requests
type Zone struct {
	filename        string
	fileLastModTime time.Time
	rrs             []dns.RR
	ns              []dns.NS
	mut             sync.Mutex
}

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

func parseRecords(zone *Zone) error {
	zone.mut.Lock()

	data, err := ioutil.ReadFile(zone.filename)
	if err != nil {
		return fmt.Errorf("Error reading zone file: %s", err)
	}

	zp := dns.NewZoneParser(bytes.NewReader(data), "", zone.filename)
	if zp.Err() != nil {
		return zp.Err()
	}

	for rr, ok := zp.Next(); ok; rr, ok = zp.Next() {
		zone.rrs = append(zone.rrs, rr)
	}

	zone.mut.Unlock()
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

	fileInfo, err := os.Stat(zonefile)
	if err != nil {
		log.Fatalf("Could not stat file: %s", err)
	}

	zone := Zone{
		filename:        *zonefile,
		fileLastModTime: fileInfo.ModTime(),
		rrs:             []dns.RR{},
		ns:              []dns.NS{},
	}

	err := parseRecords(&zone)
	if err != nil {
		return fmt.Errorf("Error parsing zone file: %s", err)
	}

	go monitorZonefile(*zonefile)

	dns.HandleFunc(".", func(w dns.ResponseWriter, m *dns.Msg) {
		// do something with dns requestion
		// like serve back the matched record(s)
	})

	return nil
}

func main() {
	if err := run(os.Args, os.Stdin); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
