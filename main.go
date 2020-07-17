package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
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

func monitorZonefile(zone *Zone) {
	t := time.NewTicker(time.Second * 30)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			fileInfo, err := os.Stat(zone.filename)
			if err != nil {
				log.Fatalf("Could not stat file: %s", err)
			}

			if fileInfo.ModTime() != zone.fileLastModTime {
				log.Printf("zone file has been modified on %s", fileInfo.ModTime())
				zone.fileLastModTime = fileInfo.ModTime()

				err = parseRecords(zone)
				if err != nil {
					log.Fatalf("Error parsing zone file: %s", err)
				}
			}
		}
	}
}

func parseRecords(zone *Zone) error {
	zone.mut.Lock()
	defer zone.mut.Unlock()

	data, err := ioutil.ReadFile(zone.filename)
	if err != nil {
		return fmt.Errorf("Error reading zone file: %s", err)
	}

	zp := dns.NewZoneParser(bytes.NewReader(data), "", zone.filename)
	if zp.Err() != nil {
		return zp.Err()
	}

	for rr, ok := zp.Next(); ok; rr, ok = zp.Next() {
		if rr.Header().Rrtype == dns.TypeNS {
			zone.ns = append(zone.ns)
		} else {
			zone.rrs = append(zone.rrs, rr)
		}
	}
	return nil
}

func run(args []string, stdin io.Reader) error {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		port = flags.Int("port", 53, "UDP port to listen on for DNS")
		//server   = flags.String("forward-server", "1.1.1.1:53", "forward DNS server")
		zonefile = flags.String("zonefile", "master.zone", "zone file name")
	)
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	fileInfo, err := os.Stat(*zonefile)
	if err != nil {
		return fmt.Errorf("Could not stat zone file: %s", err)
	}

	zone := Zone{
		filename:        *zonefile,
		fileLastModTime: fileInfo.ModTime(),
		rrs:             []dns.RR{},
		ns:              []dns.NS{},
	}

	err = parseRecords(&zone)
	if err != nil {
		return fmt.Errorf("Error parsing zone file: %s", err)
	}

	go monitorZonefile(&zone)

	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		// do something with dns requests
		// like serve back the matched record(s)
		zone.mut.Lock()

		m := new(dns.Msg)
		m.SetReply(r)
		m.Authoritative = true
		m.Answer = zone.rrs
		w.WriteMsg(m)

		zone.mut.Unlock()
	})

	srv := &dns.Server{Addr: ":" + strconv.Itoa(*port), Net: "udp"}
	if err := srv.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := run(os.Args, os.Stdin); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
