package dta

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

const (
	txtString = "TXT"
)

type NameServer struct {
	Priority int
	Host     string
	Port     int
}

type NameServers []NameServer

type request struct {
	Domain      string
	NameServers []NameServer
}

type Response struct {
	Config map[string]string
}

type PrioritySorter []NameServer

func (a PrioritySorter) Len() int           { return len(a) }
func (a PrioritySorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a PrioritySorter) Less(i, j int) bool { return a[i].Priority < a[j].Priority }

func NewRequest(domain string, ns ...NameServer) (req request) {
	// Sort nameservers by priority
	sort.Sort(PrioritySorter(ns))
	req = request{domain, ns}
	return
}

func getTxtRecord(domain string, nameservers ...NameServer) (txtRecord *dns.Msg, err error) {
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeTXT)
	m.RecursionDesired = true
	nameserverCount := len(nameservers)
	for i, nameserver := range nameservers {
		record, _, exchangeErr := c.Exchange(m, net.JoinHostPort(nameserver.Host, strconv.Itoa(nameserver.Port)))
		// If there was a DNS error
		if exchangeErr != nil {
			// and we're out of name servers to try, return the error
			if i+1 >= nameserverCount {
				err = fmt.Errorf("%s", exchangeErr)
				return
			} else {
				continue
			}
		}

		// If there was a record error
		if record.Rcode != dns.RcodeSuccess {
			// and we're out of name servers to try, return the error
			if i+1 >= nameserverCount {
				err = fmt.Errorf("%s", dns.RcodeToString[record.Rcode])
				return
			} else {
				continue
			}
		} else {
			return record, err
		}
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
		if strings.Index(rawLine, "=") <= 1 {
			continue
		}
		attributeName, valueStart := getAttribute(rawLine)
		config[attributeName] = processValue(rawLine[valueStart : len(rawLine)-1])
	}
	response.Config = config
	return
}

func (req request) Get() (response Response, err error) {
	record, err := getTxtRecord(req.Domain, req.NameServers...)
	if err == nil {
		response = processRecord(record)
	}
	return
}
