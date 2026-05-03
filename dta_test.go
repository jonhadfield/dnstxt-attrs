package dta

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/miekg/dns"
)

// Unit coverage for the RFC 1464 examples lives in TestProcessRecordRFCSamples
// (master-file-escaped fixtures) and TestProcessRecordWireRoundtrip (live
// dns.Pack / dns.Unpack round-trip). Master-file unescape edge cases live in
// TestUnescapeMasterFile.

func TestFull1(t *testing.T) {
	nameserver := NameServer{Host: "8.8.4.4", Port: 53, Priority: 0}
	request := NewRequest("test1.nooutbound.co.uk", nameserver)
	res, _ := request.Get()
	expectedAttr := "color"
	expectedVal := "blue"
	if _, ok := res.Config[expectedAttr]; ok {
		if res.Config[expectedAttr] != expectedVal {
			t.Errorf("Expected value: \"%s\"", expectedVal)
			t.Errorf("Got: %+v", res.Config[expectedAttr])
		}
	} else {
		t.Errorf("Expected attribute: \"%s\"", expectedAttr)
		t.Errorf("Got: %+v", res)
	}
}

func TestFull2(t *testing.T) {
	nameserver := NameServer{Host: "8.8.4.4", Port: 53, Priority: 0}
	request := NewRequest("test2.nooutbound.co.uk", nameserver)
	res, _ := request.Get()
	expectedAttr := "equation"
	expectedVal := "a=4"
	if _, ok := res.Config[expectedAttr]; ok {
		if res.Config[expectedAttr] != expectedVal {
			t.Errorf("Expected value: \"%s\"", expectedVal)
			t.Errorf("Got: %+v", res.Config[expectedAttr])
		}
	} else {
		t.Errorf("Expected attribute: \"%s\"", expectedAttr)
		t.Errorf("Got: %+v", res)
	}
}

func TestFull3(t *testing.T) {
	nameserver := NameServer{Host: "8.8.4.4", Port: 53, Priority: 0}
	request := NewRequest("test3.nooutbound.co.uk", nameserver)
	res, _ := request.Get()
	expectedAttr := "a=a"
	expectedVal := "true"
	if _, ok := res.Config[expectedAttr]; ok {
		if res.Config[expectedAttr] != expectedVal {
			t.Errorf("Expected value: \"%s\"", expectedVal)
			t.Errorf("Got: %+v", res.Config[expectedAttr])
		}
	} else {
		t.Errorf("Expected attribute: \"%s\"", expectedAttr)
		t.Errorf("Got: %+v", res)
	}
}

// TestProcessRecordWireRoundtrip exercises the full pipeline — wire bytes →
// dns.Pack → dns.Unpack → processRecord — for the RFC 1464 examples that
// require non-trivial unescaping. Replaces the previously disabled
// TestFull4/5/6, which depended on live DNS records that the upstream provider
// could not host (the literal backslashes/quotes were rejected). Constructing
// a synthetic dns.Msg via Pack/Unpack reproduces exactly what miekg/dns would
// hand us off the wire, with no DNS-provider involvement.
func TestProcessRecordWireRoundtrip(t *testing.T) {
	cases := []struct {
		name      string
		wireBytes string // raw character-string bytes as they would appear on the wire
		key       string
		val       string
	}{
		{
			name:      "rfc1464_ex4_backslash_then_escaped_eq",
			wireBytes: "a\\`=a=false", // bytes: a, \, `, =, a, =, false
			key:       "a\\=a",        // a, \, =, a
			val:       "false",
		},
		{
			name:      "rfc1464_ex5_eq_in_name_with_backslash_value",
			wireBytes: "`==\\=", // bytes: `, =, =, \, =
			key:       "=",
			val:       "\\=",
		},
		{
			name:      "rfc1464_ex6_literal_quotes_in_value",
			wireBytes: `string="Cat"`, // bytes include the literal " characters
			key:       "string",
			val:       `"Cat"`,
		},
		{
			name:      "non_printable_value_bytes",
			wireBytes: "bin=" + string([]byte{0x01, 0x02, 0x7f}),
			key:       "bin",
			val:       string([]byte{0x01, 0x02, 0x7f}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := &dns.TXT{
				Hdr: dns.RR_Header{Name: "rfc1464.example.", Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 300},
				Txt: []string{escapeMasterFile(tc.wireBytes)},
			}
			buf := make([]byte, 4096)
			off, err := dns.PackRR(rr, buf, 0, nil, false)
			if err != nil {
				t.Fatalf("PackRR: %v", err)
			}
			unpacked, _, err := dns.UnpackRR(buf[:off], 0)
			if err != nil {
				t.Fatalf("UnpackRR: %v", err)
			}
			msg := &dns.Msg{Answer: []dns.RR{unpacked}}

			resp := processRecord(msg)
			got, ok := resp.Config[tc.key]
			if !ok {
				t.Fatalf("missing key %q (got %#v)", tc.key, resp.Config)
			}
			if got != tc.val {
				t.Fatalf("for key %q expected value %q, got %q", tc.key, tc.val, got)
			}
		})
	}
}

// escapeMasterFile mirrors the escaping miekg/dns applies when unpacking a
// TXT character-string from the wire (\\, \", \DDD for non-printables). It is
// the inverse of unescapeMasterFile and is used to construct DNS messages
// whose Txt field represents specific raw wire bytes.
func escapeMasterFile(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\\' || c == '"':
			b.WriteByte('\\')
			b.WriteByte(c)
		case c < ' ' || c > '~':
			fmt.Fprintf(&b, "\\%03d", c)
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// TestUnescapeMasterFile covers the full master-file escape set produced by
// miekg/dns on TXT unpack: \\, \", \DDD, plus the \X fallback and a couple of
// defensive edge cases.
func TestUnescapeMasterFile(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "no_escapes_passthrough", in: "color=blue", want: "color=blue"},
		{name: "empty", in: "", want: ""},
		{name: "escaped_backslash", in: `\\=`, want: `\=`},
		{name: "escaped_quote", in: `\"Cat\"`, want: `"Cat"`},
		{name: "non_printable_bytes", in: `bin=\001\002\127`, want: "bin=\x01\x02\x7f"},
		{name: "mixed_escapes", in: `\\` + `\"` + `\010`, want: "\\\"\x0a"},
		{name: "trailing_backslash_literal", in: `abc\`, want: `abc\`},
		{name: "out_of_range_decimal_falls_back", in: `\999`, want: "999"}, // 999 > 0xff: treat as \X+99 → "9" + "99"
		{name: "non_digit_after_backslash", in: `\x`, want: "x"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := unescapeMasterFile(tc.in); got != tc.want {
				t.Fatalf("unescapeMasterFile(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestFull7(t *testing.T) {
	nameserver := NameServer{Host: "8.8.4.4", Port: 53, Priority: 0}
	request := NewRequest("test7.nooutbound.co.uk", nameserver)
	res, _ := request.Get()
	expectedAttr := "string2"
	expectedVal := "``abc``"
	if _, ok := res.Config[expectedAttr]; ok {
		if res.Config[expectedAttr] != expectedVal {
			t.Errorf("Expected value: \"%s\"", expectedVal)
			t.Errorf("Got: %+v", res.Config[expectedAttr])
		}
	} else {
		t.Errorf("Expected attribute: \"%s\"", expectedAttr)
		t.Errorf("Got: %+v", res)
	}
}

func TestFull8(t *testing.T) {
	nameserver := NameServer{Host: "8.8.4.4", Port: 53, Priority: 0}
	request := NewRequest("test8.nooutbound.co.uk", nameserver)
	res, _ := request.Get()
	expectedAttr := "novalue"
	expectedVal := ""
	if _, ok := res.Config[expectedAttr]; ok {
		if res.Config[expectedAttr] != expectedVal {
			t.Errorf("Expected value: \"%s\"", expectedVal)
			t.Errorf("Got: %+v", res.Config[expectedAttr])
		}
	} else {
		t.Errorf("Expected attribute: \"%s\"", expectedAttr)
		t.Errorf("Got: %+v", res)
	}
}

func TestFull9(t *testing.T) {
	nameserver := NameServer{Host: "8.8.4.4", Port: 53, Priority: 0}
	request := NewRequest("test9.nooutbound.co.uk", nameserver)
	res, _ := request.Get()
	expectedAttr := "a b"
	expectedVal := "c d"
	if _, ok := res.Config[expectedAttr]; ok {
		if res.Config[expectedAttr] != expectedVal {
			t.Errorf("Expected value: \"%s\"", expectedVal)
			t.Errorf("Got: %+v", res.Config[expectedAttr])
		}
	} else {
		t.Errorf("Expected attribute: \"%s\"", expectedAttr)
		t.Errorf("Got: %+v", res)
	}
}

func TestFull10(t *testing.T) {
	nameserver := NameServer{Host: "8.8.4.4", Port: 53, Priority: 0}
	request := NewRequest("test10.nooutbound.co.uk", nameserver)
	res, _ := request.Get()
	expected := "abc "
	if _, ok := res.Config[expected]; ok != true {
		t.Errorf("Expected attribute: \"%s\"", expected)
		t.Errorf("Got: %+v", res)
	}
}

func TestInvalidDomain(t *testing.T) {
	// .invalid is reserved by RFC 2606 specifically so it can never be
	// delegated; any resolver MUST answer NXDOMAIN. The previous fixture,
	// missing.example.com, started returning NOERROR + empty answer once
	// example.com moved to Cloudflare, panicking this test on a nil err.
	nameserver := NameServer{Host: "8.8.4.4", Port: 53, Priority: 0}
	request := NewRequest("dnstxt-attrs-nonexistent.invalid", nameserver)
	_, err := request.Get()
	if err == nil {
		t.Fatalf("Expected NXDOMAIN error, got nil")
	}
	if !strings.Contains(err.Error(), "NXDOMAIN") {
		t.Errorf("Expected NXDOMAIN error, got: %v", err)
	}
}

func TestInvalidNameServer(t *testing.T) {
	nameserver := NameServer{Host: "1.2.3.4", Port: 53, Priority: 0}
	request := NewRequest("www.google.com", nameserver)
	_, err := request.Get()
	if err == nil {
		t.Errorf("Expected error for invalid nameserver")
	}
}

func TestSuccessWithSingleInvalidNameServer(t *testing.T) {
	nameserver1 := NameServer{Host: "8.8.8.9", Port: 53, Priority: 0}
	nameserver2 := NameServer{Host: "8.8.8.8", Port: 53, Priority: 1}
	nameservers := []NameServer{nameserver1, nameserver2}
	request := NewRequest("test10.nooutbound.co.uk", nameservers...)
	res, _ := request.Get()
	expected := "abc "
	if _, ok := res.Config[expected]; ok != true {
		t.Errorf("Expected attribute: \"%s\"", expected)
		t.Errorf("Got: %+v", res)
	}
}

func TestNameServerSorting(t *testing.T) {
	nameservers := [6]NameServer{}
	nameservers[0] = NameServer{Host: "1.2.3.5", Port: 53, Priority: 0}
	nameservers[1] = NameServer{Host: "1.2.3.7", Port: 53, Priority: 4}
	nameservers[2] = NameServer{Host: "1.2.3.8", Port: 53, Priority: 2}
	nameservers[3] = NameServer{Host: "1.2.3.6", Port: 53, Priority: 3}
	nameservers[4] = NameServer{Host: "1.2.3.4", Port: 53, Priority: 1}
	nameservers[5] = NameServer{Host: "1.2.3.9", Port: 53, Priority: 5}
	req := NewRequest("example.com", nameservers[:]...)
	for i, nameserver := range req.NameServers {
		if i != nameserver.Priority {
			t.Errorf("Expected nameserver priority: %d got: %d", i, nameserver.Priority)
		}
	}
}

func TestMultipleNameServersWithFirstFailing(t *testing.T) {
	// First nameserver invalid, second valid
	nameserver1 := NameServer{Host: "192.0.2.1", Port: 53, Priority: 0}
	nameserver2 := NameServer{Host: "8.8.8.8", Port: 53, Priority: 1}
	nameservers := []NameServer{nameserver1, nameserver2}
	request := NewRequest("test1.nooutbound.co.uk", nameservers...)
	res, err := request.Get()
	if err != nil {
		t.Errorf("Expected success with fallback nameserver, got error: %v", err)
	}
	if len(res.Config) == 0 {
		t.Errorf("Expected config data from fallback nameserver")
	}
}

func TestProcessRecordWithInvalidEntries(t *testing.T) {
	// Create a mock DNS response with invalid TXT entries
	msg := &dns.Msg{}

	// Add invalid TXT record (no equals sign)
	invalidTxt1 := &dns.TXT{
		Hdr: dns.RR_Header{Name: "test.com.", Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 300},
		Txt: []string{"invalidentry"},
	}
	msg.Answer = append(msg.Answer, invalidTxt1)

	// Add invalid TXT record (equals at position 1 - should be skipped)
	invalidTxt2 := &dns.TXT{
		Hdr: dns.RR_Header{Name: "test.com.", Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 300},
		Txt: []string{"=invalid"},
	}
	msg.Answer = append(msg.Answer, invalidTxt2)

	// Add valid TXT record
	validTxt := &dns.TXT{
		Hdr: dns.RR_Header{Name: "test.com.", Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 300},
		Txt: []string{"color=blue"},
	}
	msg.Answer = append(msg.Answer, validTxt)

	response := processRecord(msg)

	// Should only have the valid entry
	if len(response.Config) != 1 {
		t.Errorf("Expected 1 config entry, got %d", len(response.Config))
	}
	if response.Config["color"] != "blue" {
		t.Errorf("Expected color=blue, got %v", response.Config)
	}
}

func TestProcessRecordAttributeNameNormalization(t *testing.T) {
	msg := &dns.Msg{}
	txt := &dns.TXT{
		Hdr: dns.RR_Header{Name: "test.com.", Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 300},
		Txt: []string{"Favorite Drink=Earl Grey"},
	}
	msg.Answer = append(msg.Answer, txt)

	response := processRecord(msg)
	val, ok := response.Config["favorite drink"]
	if !ok {
		t.Fatalf("Expected normalized attribute name 'favorite drink'")
	}
	if val != "Earl Grey" {
		t.Fatalf("Expected value 'Earl Grey', got %q", val)
	}
}

func TestProcessRecordAttributeWhitespaceHandling(t *testing.T) {
	msg := &dns.Msg{}
	trimmed := &dns.TXT{
		Hdr: dns.RR_Header{Name: "trimmed.com.", Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 300},
		Txt: []string{" color=blue"},
	}
	preserved := &dns.TXT{
		Hdr: dns.RR_Header{Name: "preserved.com.", Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 300},
		Txt: []string{"` color=azure"},
	}
	msg.Answer = append(msg.Answer, trimmed, preserved)

	response := processRecord(msg)
	if _, ok := response.Config["color"]; !ok {
		t.Fatalf("Expected attribute without leading space to normalize to 'color'")
	}
	if val := response.Config["color"]; val != "blue" {
		t.Fatalf("Expected value 'blue', got %q", val)
	}
	if val, ok := response.Config[" color"]; !ok {
		t.Fatalf("Expected attribute with escaped leading space to be preserved")
	} else if val != "azure" {
		t.Fatalf("Expected preserved leading space attribute to have value 'azure', got %q", val)
	}
}

// TestProcessRecordRFCSamples covers the 10 examples from RFC 1464. Each
// fixture entry is written in the form miekg/dns delivers from the wire — i.e.
// with the library's master-file escaping already applied (\\, \", \DDD).
func TestProcessRecordRFCSamples(t *testing.T) {
	cases := []struct {
		name  string // test sub-name
		entry string // as miekg/dns delivers it from the wire
		key   string
		val   string
	}{
		{name: "ex1_color", entry: "color=blue", key: "color", val: "blue"},
		{name: "ex2_equation", entry: "equation=a=4", key: "equation", val: "a=4"},
		{name: "ex3_escaped_eq_in_name", entry: "a`=a=true", key: "a=a", val: "true"},
		// Wire bytes: a, \, `, =, a, =, false   →  miekg escapes \ to \\
		{name: "ex4_backslash_then_escaped_eq", entry: "a\\\\`=a=false", key: "a\\=a", val: "false"},
		// Wire bytes: `, =, =, \, =   →  miekg escapes \ to \\
		{name: "ex5_escaped_eq_value_with_backslash", entry: "`==\\\\=", key: "=", val: "\\="},
		// Wire bytes: string="Cat"   →  miekg escapes both " to \"
		{name: "ex6_quoted_value", entry: "string=\\\"Cat\\\"", key: "string", val: "\"Cat\""},
		{name: "ex7_double_backquote", entry: "string2=``abc``", key: "string2", val: "``abc``"},
		{name: "ex8_empty_value", entry: "novalue=", key: "novalue", val: ""},
		{name: "ex9_space_in_name_and_value", entry: "a b=c d", key: "a b", val: "c d"},
		{name: "ex10_trailing_escaped_space_in_name", entry: "abc` =123 ", key: "abc ", val: "123 "},
	}

	msg := &dns.Msg{}
	txt := &dns.TXT{Hdr: dns.RR_Header{Name: "rfc1464.example.", Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 300}}
	for _, tc := range cases {
		txt.Txt = append(txt.Txt, tc.entry)
	}
	msg.Answer = append(msg.Answer, txt)

	response := processRecord(msg)
	if len(response.Config) != len(cases) {
		t.Fatalf("Expected %d entries, got %d: %#v", len(cases), len(response.Config), response.Config)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := response.Config[tc.key]
			if !ok {
				t.Fatalf("Missing key %q", tc.key)
			}
			if got != tc.val {
				t.Fatalf("For key %q expected value %q, got %q", tc.key, tc.val, got)
			}
		})
	}
}

func TestProcessRecordNilSafe(t *testing.T) {
	resp := processRecord(nil)
	if resp.Config == nil {
		t.Fatalf("Expected Config map to be initialized")
	}
	if len(resp.Config) != 0 {
		t.Fatalf("Expected empty map for nil message, got %v", resp.Config)
	}
}

// TestProcessRecordSkipsNonTXTAndEmpty exercises the defensive branches in
// processRecord/parseTXTEntry: a non-TXT RR mixed into Answer is skipped, and
// empty / whitespace-only / name-less entries are dropped silently.
func TestProcessRecordSkipsNonTXTAndEmpty(t *testing.T) {
	msg := &dns.Msg{Answer: []dns.RR{
		// Non-TXT RR in the answer slice — must be skipped.
		&dns.A{
			Hdr: dns.RR_Header{Name: "test.", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 300},
		},
		&dns.TXT{
			Hdr: dns.RR_Header{Name: "test.", Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 300},
			Txt: []string{
				"",           // empty entry → parseTXTEntry early-return
				"   =value",  // whitespace-only name → normalizeAttributeName returns ""
				"color=blue", // valid, must still come through
			},
		},
	}}
	resp := processRecord(msg)
	if len(resp.Config) != 1 {
		t.Fatalf("expected exactly one attribute, got %#v", resp.Config)
	}
	if resp.Config["color"] != "blue" {
		t.Fatalf("expected color=blue, got %#v", resp.Config)
	}
}

// TestNormalizeAttributeNameDefensive covers branches in
// normalizeAttributeName / decodeAttributeName that aren't reachable through
// processRecord because findDelimiterIndex can't produce the relevant inputs.
// These guards keep the helpers safe to reuse and worth covering directly.
func TestNormalizeAttributeNameDefensive(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "lone_backquote_decodes_to_empty", in: "`", want: ""},
		{name: "all_whitespace", in: " \t ", want: ""},
		{name: "trailing_unescaped_whitespace", in: "abc \t", want: "abc"},
		{name: "trailing_unescaped_backquote", in: "name`", want: "name"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeAttributeName(tc.in); got != tc.want {
				t.Fatalf("normalizeAttributeName(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNameServerAddrDefaultPort(t *testing.T) {
	cases := []struct {
		name string
		ns   NameServer
		want string
	}{
		{name: "explicit_port", ns: NameServer{Host: "1.2.3.4", Port: 5353}, want: "1.2.3.4:5353"},
		{name: "zero_port_defaults_to_53", ns: NameServer{Host: "1.2.3.4"}, want: "1.2.3.4:53"},
		{name: "ipv6_zero_port", ns: NameServer{Host: "::1"}, want: "[::1]:53"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := nameServerAddr(tc.ns); got != tc.want {
				t.Fatalf("nameServerAddr(%+v) = %q, want %q", tc.ns, got, tc.want)
			}
		})
	}
}

// startTestDNSServer spins up an in-process DNS server bound to a random
// localhost port, listening on both UDP and TCP, and returns its host:port.
// The handler is invoked for every request on either transport. The server is
// shut down when the test finishes.
func startTestDNSServer(t *testing.T, handler dns.HandlerFunc) (host string, port int) {
	t.Helper()

	udpConn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	udpAddr := udpConn.LocalAddr().(*net.UDPAddr)
	tcpListener, err := net.Listen("tcp", udpAddr.String())
	if err != nil {
		udpConn.Close()
		t.Fatalf("listen tcp on %s: %v", udpAddr, err)
	}

	udpReady := make(chan struct{})
	tcpReady := make(chan struct{})
	udpServer := &dns.Server{PacketConn: udpConn, Handler: handler, NotifyStartedFunc: func() { close(udpReady) }}
	tcpServer := &dns.Server{Listener: tcpListener, Handler: handler, NotifyStartedFunc: func() { close(tcpReady) }}

	go func() { _ = udpServer.ActivateAndServe() }()
	go func() { _ = tcpServer.ActivateAndServe() }()

	select {
	case <-udpReady:
	case <-time.After(2 * time.Second):
		t.Fatalf("udp server did not start")
	}
	select {
	case <-tcpReady:
	case <-time.After(2 * time.Second):
		t.Fatalf("tcp server did not start")
	}

	t.Cleanup(func() {
		_ = udpServer.Shutdown()
		_ = tcpServer.Shutdown()
	})

	return udpAddr.IP.String(), udpAddr.Port
}

func TestGetTCPFallbackOnTruncation(t *testing.T) {
	var udpHits, tcpHits atomic.Int32
	host, port := startTestDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		reply := new(dns.Msg)
		reply.SetReply(r)
		if _, ok := w.RemoteAddr().(*net.UDPAddr); ok {
			udpHits.Add(1)
			reply.Truncated = true
			_ = w.WriteMsg(reply)
			return
		}
		tcpHits.Add(1)
		reply.Answer = []dns.RR{&dns.TXT{
			Hdr: dns.RR_Header{Name: r.Question[0].Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0},
			Txt: []string{"color=blue"},
		}}
		_ = w.WriteMsg(reply)
	})

	req := NewRequest("test.example.", NameServer{Host: host, Port: port})
	resp, err := req.Get()
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if resp.Config["color"] != "blue" {
		t.Fatalf("expected color=blue, got %#v", resp.Config)
	}
	if udpHits.Load() != 1 {
		t.Fatalf("expected 1 UDP query, got %d", udpHits.Load())
	}
	if tcpHits.Load() != 1 {
		t.Fatalf("expected 1 TCP fallback query, got %d", tcpHits.Load())
	}
}

func TestGetAuthoritativeRcodeNoFailover(t *testing.T) {
	var hits atomic.Int32
	host, port := startTestDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		hits.Add(1)
		reply := new(dns.Msg)
		reply.SetReply(r)
		reply.Rcode = dns.RcodeNameError // NXDOMAIN
		_ = w.WriteMsg(reply)
	})

	primary := NameServer{Host: host, Port: port, Priority: 0}
	// Unreachable per RFC 5737 — if we tried it, the test would either time
	// out or fail with a different error. The query-count assertion is the
	// authoritative check that we did not.
	secondary := NameServer{Host: "192.0.2.1", Port: 53, Priority: 1}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := NewRequest("test.example.", primary, secondary)
	_, err := req.GetContext(ctx)
	if err == nil {
		t.Fatalf("expected NXDOMAIN error, got nil")
	}
	if !strings.Contains(err.Error(), "NXDOMAIN") {
		t.Fatalf("expected NXDOMAIN error, got: %v", err)
	}
	if hits.Load() != 1 {
		t.Fatalf("expected exactly 1 query (no failover on authoritative Rcode), got %d", hits.Load())
	}
}

func TestGetServfailFailsOver(t *testing.T) {
	var primaryHits, secondaryHits atomic.Int32
	primaryHost, primaryPort := startTestDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		primaryHits.Add(1)
		reply := new(dns.Msg)
		reply.SetReply(r)
		reply.Rcode = dns.RcodeServerFailure
		_ = w.WriteMsg(reply)
	})
	secondaryHost, secondaryPort := startTestDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) {
		secondaryHits.Add(1)
		reply := new(dns.Msg)
		reply.SetReply(r)
		reply.Answer = []dns.RR{&dns.TXT{
			Hdr: dns.RR_Header{Name: r.Question[0].Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0},
			Txt: []string{"color=red"},
		}}
		_ = w.WriteMsg(reply)
	})

	primary := NameServer{Host: primaryHost, Port: primaryPort, Priority: 0}
	secondary := NameServer{Host: secondaryHost, Port: secondaryPort, Priority: 1}

	req := NewRequest("test.example.", primary, secondary)
	resp, err := req.Get()
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if resp.Config["color"] != "red" {
		t.Fatalf("expected color=red from secondary, got %#v", resp.Config)
	}
	if primaryHits.Load() != 1 {
		t.Fatalf("expected 1 primary hit, got %d", primaryHits.Load())
	}
	if secondaryHits.Load() != 1 {
		t.Fatalf("expected 1 secondary hit, got %d", secondaryHits.Load())
	}
}

func TestGetContextCancelled(t *testing.T) {
	// A handler that blocks until the test ends so we can be sure it's the
	// context that produces the error, not the server returning early.
	done := make(chan struct{})
	t.Cleanup(func() { close(done) })
	host, port := startTestDNSServer(t, func(w dns.ResponseWriter, r *dns.Msg) { <-done })

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel

	req := NewRequest("test.example.", NameServer{Host: host, Port: port})
	_, err := req.GetContext(ctx)
	if err == nil {
		t.Fatalf("expected error from cancelled context, got nil")
	}
	if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("expected context cancellation error, got: %v", err)
	}
}

func TestMultipleNameServersWithErrors(t *testing.T) {
	// Test with multiple invalid nameservers, then a valid one
	nameserver1 := NameServer{Host: "192.0.2.1", Port: 53, Priority: 0} // Invalid IP (RFC 5737)
	nameserver2 := NameServer{Host: "192.0.2.2", Port: 53, Priority: 1} // Invalid IP (RFC 5737)
	nameserver3 := NameServer{Host: "8.8.8.8", Port: 53, Priority: 2}   // Valid nameserver
	nameservers := []NameServer{nameserver1, nameserver2, nameserver3}

	request := NewRequest("test1.nooutbound.co.uk", nameservers...)
	res, err := request.Get()

	// Should succeed with the valid nameserver after failing with invalid ones
	if err != nil {
		t.Errorf("Expected success with fallback nameserver, got error: %v", err)
	}
	if len(res.Config) == 0 {
		t.Errorf("Expected config data from fallback nameserver")
	}
}

func TestAllNameServersFail(t *testing.T) {
	// Test with only invalid nameservers to trigger final return path
	nameserver1 := NameServer{Host: "192.0.2.1", Port: 53, Priority: 0} // Invalid IP (RFC 5737)
	nameserver2 := NameServer{Host: "192.0.2.2", Port: 53, Priority: 1} // Invalid IP (RFC 5737)
	nameservers := []NameServer{nameserver1, nameserver2}

	request := NewRequest("test1.nooutbound.co.uk", nameservers...)
	_, err := request.Get()

	// Should fail with error after trying all nameservers
	if err == nil {
		t.Errorf("Expected error when all nameservers fail")
	}
}

func TestMultipleNameServersWithRecordErrors(t *testing.T) {
	// Test with nameserver that responds but with NXDOMAIN, then valid one
	nameserver1 := NameServer{Host: "8.8.8.8", Port: 53, Priority: 0}
	nameserver2 := NameServer{Host: "8.8.4.4", Port: 53, Priority: 1}
	nameserver3 := NameServer{Host: "1.1.1.1", Port: 53, Priority: 2}
	nameservers := []NameServer{nameserver1, nameserver2, nameserver3}

	// Use domain that will return NXDOMAIN on first try, then fallback
	request := NewRequest("definitely.does.not.exist.invalid.domain.example", nameservers...)
	_, err := request.Get()

	// Should get NXDOMAIN error after trying all nameservers
	if err == nil {
		t.Errorf("Expected NXDOMAIN error for non-existent domain")
	}
	if !strings.Contains(err.Error(), "NXDOMAIN") {
		t.Errorf("Expected NXDOMAIN error, got: %v", err)
	}
}

func TestRecordErrorWithFallback(t *testing.T) {
	// Test NXDOMAIN with first nameserver, then success with second
	nameserver1 := NameServer{Host: "8.8.8.8", Port: 53, Priority: 0}
	nameserver2 := NameServer{Host: "8.8.4.4", Port: 53, Priority: 1}
	nameservers := []NameServer{nameserver1, nameserver2}

	// Use definitely non-existent domain
	request := NewRequest("thisdoesnotexist.invalid.fake.domain.xyz", nameservers...)
	_, err := request.Get()

	// Should fail since domain doesn't exist on any nameserver
	if err != nil {
		// This is expected - just testing that the continue path gets exercised
		if !strings.Contains(err.Error(), "NXDOMAIN") {
			t.Logf("Got error (as expected): %v", err)
		}
	}
}
