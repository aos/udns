package main

import (
	"bytes"
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

// Zone contains all the zone resource records parsed from file
type Zone struct {
	filename        string
	fileLastModTime time.Time
	rrs             []dns.RR
	ns              []dns.RR
	mut             sync.RWMutex
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

	// Reset resource records and parse again with new data
	zone.rrs = zone.rrs[:0]
	zone.ns = zone.ns[:0]

	for rr, ok := zp.Next(); ok; rr, ok = zp.Next() {
		zone.rrs = append(zone.rrs, rr)

		if rr.Header().Rrtype == dns.TypeNS {
			zone.ns = append(zone.ns, rr)
		}
	}
	return nil
}

func resolve(server, fqdn string, rrType uint16) []dns.RR {
	m := new(dns.Msg)
	m.Id = dns.Id()
	m.SetQuestion(fqdn, rrType)
	m.RecursionDesired = true

	in, err := dns.Exchange(m, server)
	if err != nil {
		log.Printf("ERROR: unable to resolve %s\n", err)
		return []dns.RR{}
	}

	return in.Answer
}

func run(args []string, stdin io.Reader) error {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		address  = flags.String("address", "", "address to listen on")
		port     = flags.String("port", "8053", "UDP port to listen on for DNS")
		server   = flags.String("forward-server", "1.1.1.1:53", "forward DNS server")
		zonefile = flags.String("zonefile", "master.zone", "zone file name")
	)
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	fileInfo, err := os.Stat(*zonefile)
	if err != nil {
		return fmt.Errorf("Could not stat zone file: %s", err)
	}

	// Set local resolver address if not specified
	if *address == "" {
		*address = "127.0.0.1"
	}

	zone := Zone{
		filename:        *zonefile,
		fileLastModTime: fileInfo.ModTime(),
		rrs:             []dns.RR{},
		ns:              []dns.RR{},
	}

	err = parseRecords(&zone)
	if err != nil {
		return fmt.Errorf("Error parsing zone file: %s", err)
	}

	go monitorZonefile(&zone)

	dns.HandleFunc(".", func(w dns.ResponseWriter, req *dns.Msg) {
		zone.mut.RLock()
		defer zone.mut.RUnlock()

		m := new(dns.Msg)
		m.SetReply(req)
		m.Authoritative = true
		m.RecursionAvailable = true

		for _, q := range req.Question {
			answers := []dns.RR{}

			for _, rr := range zone.rrs {
				rh := rr.Header()

				// 1. handle CNAMEs
				// should call resolve function here (with localhost)
				if q.Name == rh.Name && (rh.Rrtype == dns.TypeCNAME || q.Qtype == dns.TypeCNAME) {
					answers = append(answers, rr)

					for _, a := range resolve(*address+":"+*port, rr.(*dns.CNAME).Target, q.Qtype) {
						answers = append(answers, a)
					}
				}

				// 2. handle everything else
				if q.Name == rh.Name && q.Qtype == rh.Rrtype && q.Qclass == rh.Class {
					answers = append(answers, rr)
				}
			}

			// if we can't find the answer, then recursively
			// resolve with forward DNS server
			if len(answers) == 0 && *server != "" {
				for _, a := range resolve(*server, q.Name, q.Qtype) {
					answers = append(answers, a)
				}
			} else {
				// Set local name server as authority
				m.Ns = zone.ns
			}

			m.Answer = append(m.Answer, answers...)
		}
		w.WriteMsg(m)
	})

	srv := &dns.Server{Addr: *address + ":" + *port, Net: "udp"}
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
