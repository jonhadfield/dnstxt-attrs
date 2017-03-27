package dta

import (
	"github.com/miekg/dns"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	txtString = "TXT"
)

type NameServer struct {
	Priority int
	Host     string
	Port     int
}

type Request struct {
	Domain      string
	NameServers []NameServer
}

type Response struct {
	Config map[string]string
}

func getTxtRecord(domain string, nameservers ...NameServer) (txtRecord *dns.Msg) {
	// Sort nameservers by priority
	sort.Slice(nameservers, func(i, j int) bool { return nameservers[i].Priority < nameservers[j].Priority })

	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeTXT)
	m.RecursionDesired = true
	for _, nameserver := range nameservers {
		record, _, err := c.Exchange(m, net.JoinHostPort(nameserver.Host, strconv.Itoa(nameserver.Port)))
		if record == nil {
			log.Fatalf("*** error: %s\n", err.Error())
			continue
		}

		if record.Rcode != dns.RcodeSuccess {
			log.Fatalf(" *** invalid answer name %s after TXT query for %s\n", os.Args[1], os.Args[1])
			continue
		}
		return record
	}
	return
}

// Extracts the attribute name and
// returns it with the start position of the value
func getAttribute(s string) (a string, valueStart int) {
	// Find first equals sign not preceded by `
	var attributeEnd int
	for i, c := range s {
		if c == '=' {
			if s[i-1] == '`' {
				continue
			} else {
				attributeEnd = i
				valueStart = i + 1
				break
			}
		}
	}
	a = s[1:attributeEnd]
	a = strings.Replace(a, "`=", "=", -1)
	a = strings.Replace(a, "` ", " ", -1)
	a = strings.Replace(a, "\\\\", "\\", -1)
	return
}

// Process quoted values in the extracted value string
func processValue(rawVal string) (processedVal string) {
	// Replace double backticks with single
	processedVal = strings.Replace(rawVal, "``", "`", -1)
	processedVal = strings.Replace(rawVal, "\\\\", "\\", -1)
	return
}

// Extract each TXT entry and return a map of the kv pairs
func processRecord(txtRecord *dns.Msg) (response Response) {
	var config map[string]string
	config = make(map[string]string)
	for _, a := range txtRecord.Answer {
		rawLine := strings.TrimSpace(a.String()[strings.LastIndex(a.String(), txtString)+len(txtString):])
		// Check '=' exists and isn't first char
		equalsPos := strings.Index(rawLine, "=")
		if equalsPos <= 1 {
			continue
		}
		attributeName, valueStart := getAttribute(rawLine)
		config[attributeName] = processValue(rawLine[valueStart : len(rawLine)-1])
	}
	response.Config = config
	return
}

func (request Request) Get() (response Response) {
	record := getTxtRecord(request.Domain, request.NameServers...)
	response = processRecord(record)
	return
}
