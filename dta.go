package dta

import (
	"context"
	"errors"
	"net"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/miekg/dns"
)

const (
	quoteChar = '`'

	defaultDNSPort = 53
	// ednsBufferSize is advertised in the OPT pseudo-RR so authoritative
	// servers can return responses larger than 512 bytes without setting the
	// truncation bit. RFC 1464 attribute sets routinely exceed 512 bytes.
	ednsBufferSize = 4096
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

func NewRequest(domain string, ns ...NameServer) Request {
	sort.Slice(ns, func(i, j int) bool { return ns[i].Priority < ns[j].Priority })
	return Request{Domain: domain, NameServers: ns}
}

// Get queries the configured nameservers using context.Background(). Use
// GetContext to bind the query to a deadline or cancellation signal.
func (req Request) Get() (Response, error) {
	return req.GetContext(context.Background())
}

// GetContext queries the configured nameservers, returning the first
// successful response. Network errors and SERVFAIL trigger failover to the
// next nameserver; authoritative non-success Rcodes (NXDOMAIN, REFUSED, ...)
// are returned immediately because every nameserver will return the same
// answer. Truncated UDP responses are automatically retried over TCP for the
// same nameserver. Context cancellation aborts the query immediately.
func (req Request) GetContext(ctx context.Context) (Response, error) {
	record, err := getTxtRecord(ctx, req.Domain, req.NameServers...)
	if err != nil {
		return Response{}, err
	}
	return processRecord(record), nil
}

// nameServerAddr returns the host:port string for a NameServer, defaulting
// Port to 53 when unset (zero) so callers can write NameServer{Host: "..."}
// without remembering the port.
func nameServerAddr(ns NameServer) string {
	port := ns.Port
	if port == 0 {
		port = defaultDNSPort
	}
	return net.JoinHostPort(ns.Host, strconv.Itoa(port))
}

func getTxtRecord(ctx context.Context, domain string, nameservers ...NameServer) (*dns.Msg, error) {
	udp := &dns.Client{Net: "udp"}
	tcp := &dns.Client{Net: "tcp"}

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeTXT)
	m.RecursionDesired = true
	m.SetEdns0(ednsBufferSize, false)

	var lastErr error
	for _, ns := range nameservers {
		addr := nameServerAddr(ns)
		record, _, err := udp.ExchangeContext(ctx, m, addr)
		if err == nil && record.Truncated {
			// RFC 7766: client SHOULD retry over TCP when TC=1.
			record, _, err = tcp.ExchangeContext(ctx, m, addr)
		}
		if err != nil {
			if ctx.Err() != nil {
				return nil, err
			}
			lastErr = err
			continue
		}
		if record.Rcode == dns.RcodeSuccess {
			return record, nil
		}
		rcodeErr := errors.New(dns.RcodeToString[record.Rcode])
		if record.Rcode == dns.RcodeServerFailure {
			// SERVFAIL is transient; another nameserver might succeed.
			lastErr = rcodeErr
			continue
		}
		// Authoritative non-success — every nameserver will say the same.
		return nil, rcodeErr
	}
	return nil, lastErr
}

// processRecord parses a DNS TXT response into the RFC 1464 attribute map.
func processRecord(txtRecord *dns.Msg) Response {
	config := make(map[string]string)
	if txtRecord == nil {
		return Response{Config: config}
	}
	for _, rr := range txtRecord.Answer {
		txtRR, ok := rr.(*dns.TXT)
		if !ok {
			continue
		}
		for _, entry := range txtRR.Txt {
			if attr, value, ok := parseTXTEntry(entry); ok {
				config[attr] = value
			}
		}
	}
	return Response{Config: config}
}

// parseTXTEntry parses a single TXT character-string as delivered by miekg/dns.
// It first reverses the master-file escaping applied by the library, then
// applies the RFC 1464 attribute=value layer (backquote-escaped delimiter,
// whitespace trimming, lowercased name).
func parseTXTEntry(entry string) (string, string, bool) {
	if entry == "" {
		return "", "", false
	}
	entry = unescapeMasterFile(entry)
	delimiter := findDelimiterIndex(entry)
	if delimiter <= 0 {
		return "", "", false
	}
	name := normalizeAttributeName(entry[:delimiter])
	if name == "" {
		return "", "", false
	}
	return name, entry[delimiter+1:], true
}

// unescapeMasterFile reverses the escaping miekg/dns applies to TXT
// character-strings on unpack:
//
//	\\        -> \
//	\"        -> "
//	\DDD      -> byte with decimal value DDD (0–255), DDD always 3 digits
//	\X (any other char) -> X
//
// A trailing lone backslash is treated as a literal '\' for safety.
func unescapeMasterFile(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		if s[i] != '\\' {
			b.WriteByte(s[i])
			i++
			continue
		}
		if i+1 >= len(s) {
			b.WriteByte('\\')
			i++
			continue
		}
		if i+3 < len(s) && isASCIIDigit(s[i+1]) && isASCIIDigit(s[i+2]) && isASCIIDigit(s[i+3]) {
			v := int(s[i+1]-'0')*100 + int(s[i+2]-'0')*10 + int(s[i+3]-'0')
			if v <= 0xff {
				b.WriteByte(byte(v))
				i += 4
				continue
			}
		}
		b.WriteByte(s[i+1])
		i += 2
	}
	return b.String()
}

func isASCIIDigit(b byte) bool { return b >= '0' && b <= '9' }

// findDelimiterIndex finds the byte offset of the first unquoted '=' separating
// an RFC 1464 attribute name from its value. The accent-grave (`) escapes the
// next character (per RFC 1464); backslash escapes are NOT recognized here and
// must be resolved by the caller before invoking this function.
func findDelimiterIndex(s string) int {
	prevQuote := false
	for idx, r := range s {
		if r == '=' && !prevQuote {
			return idx
		}
		if prevQuote {
			prevQuote = false
			continue
		}
		if r == quoteChar {
			prevQuote = true
		}
	}
	return -1
}

// normalizeAttributeName decodes the RFC 1464 backquote-escape layer, trims
// unescaped leading/trailing space/tab (per RFC 1464 "spaces and tabs"), and
// lowercases the result for case-insensitive lookup. Master-file escapes are
// expected to have been resolved before this is called.
func normalizeAttributeName(raw string) string {
	if raw == "" {
		return ""
	}
	runes, escaped := decodeAttributeName(raw)
	if len(runes) == 0 {
		return ""
	}
	start := 0
	for start < len(runes) && isAttributeWhitespace(runes[start]) && !escaped[start] {
		start++
	}
	end := len(runes) - 1
	for end >= start && isAttributeWhitespace(runes[end]) && !escaped[end] {
		end--
	}
	if start > end {
		return ""
	}
	return strings.ToLower(string(runes[start : end+1]))
}

func decodeAttributeName(raw string) ([]rune, []bool) {
	var (
		runes   []rune
		escaped []bool
	)
	for len(raw) > 0 {
		r, size := utf8.DecodeRuneInString(raw)
		if r == quoteChar {
			raw = raw[size:]
			if len(raw) == 0 {
				break
			}
			next, nextSize := utf8.DecodeRuneInString(raw)
			runes = append(runes, next)
			escaped = append(escaped, true)
			raw = raw[nextSize:]
			continue
		}
		runes = append(runes, r)
		escaped = append(escaped, false)
		raw = raw[size:]
	}
	return runes, escaped
}

func isAttributeWhitespace(r rune) bool {
	return r == ' ' || r == '\t'
}
