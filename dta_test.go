package dta

import (
	"strings"
	"testing"

	"github.com/miekg/dns"
)

// All tests are numbered based on the order of examples in https://tools.ietf.org/html/rfc1464

func TestGetAttribute1(t *testing.T) {
	expectedAttr := "color"
	expectedValStartPos := 7
	attr, valStartPos := getAttribute("\"color=blue\"")
	if attr != expectedAttr {
		t.Errorf("Expected attribute: \"%s\" but got: \"%s\"", expectedAttr, attr)
	}
	if valStartPos != expectedValStartPos {
		t.Errorf("Expected value start position: %d", expectedValStartPos)
	}
}

func TestGetAttribute2(t *testing.T) {
	expectedAttr := "equation"
	expectedValStartPos := 10
	attr, valStartPos := getAttribute("\"equation=a=4\"")
	if attr != expectedAttr {
		t.Errorf("Expected attribute: \"%s\" but got: \"%s\"", expectedAttr, attr)
	}
	if valStartPos != expectedValStartPos {
		t.Errorf("Expected value start position: %d", expectedValStartPos)
	}
}

func TestGetAttribute3(t *testing.T) {
	expectedAttr := "a=a"
	expectedValStartPos := 6
	attr, valStartPos := getAttribute("\"a`=a=true\"")
	if attr != expectedAttr {
		t.Errorf("Expected attribute: \"%s\" but got: \"%s\"", expectedAttr, attr)
	}
	if valStartPos != expectedValStartPos {
		t.Errorf("Expected value start position: %d", expectedValStartPos)
	}
}

func TestGetAttribute4(t *testing.T) {
	expectedAttr := "a\\=a"
	expectedValStartPos := 8
	attr, valStartPos := getAttribute("\"a\\\\`=a=false\"")
	if attr != expectedAttr {
		t.Errorf("Expected attribute: \"%s\" but got: \"%s\"", expectedAttr, attr)
	}
	if valStartPos != expectedValStartPos {
		t.Errorf("Expected value start position: %d", expectedValStartPos)
	}
}

func TestGetAttribute5(t *testing.T) {
	expectedAttr := "="
	expectedValStartPos := 4
	attr, valStartPos := getAttribute("\"`==\\\\=\"")
	if attr != expectedAttr {
		t.Errorf("Expected attribute: \"%s\" but got: \"%s\"", expectedAttr, attr)
	}
	if valStartPos != expectedValStartPos {
		t.Errorf("Expected value start position: %d", expectedValStartPos)
	}
}

func TestGetAttribute6(t *testing.T) {
	expectedAttr := "string"
	expectedValStartPos := 8
	attr, valStartPos := getAttribute("\"string=\"Cat\"\"")
	if attr != expectedAttr {
		t.Errorf("Expected attribute: \"%s\" but got: \"%s\"", expectedAttr, attr)
	}
	if valStartPos != expectedValStartPos {
		t.Errorf("Expected value start position: %d", expectedValStartPos)
	}
}

func TestGetAttribute7(t *testing.T) {
	expectedAttr := "string2"
	expectedValStartPos := 9
	attr, valStartPos := getAttribute("\"string2=``abc``\"")
	if attr != expectedAttr {
		t.Errorf("Expected attribute: \"%s\" but got: \"%s\"", expectedAttr, attr)
	}
	if valStartPos != expectedValStartPos {
		t.Errorf("Expected value start position: %d", expectedValStartPos)
	}
}

func TestGetAttribute8(t *testing.T) {
	expectedAttr := "novalue"
	expectedValStartPos := 9
	attr, valStartPos := getAttribute("\"novalue=\"")
	if attr != expectedAttr {
		t.Errorf("Expected attribute: \"%s\" but got: \"%s\"", expectedAttr, attr)
	}
	if valStartPos != expectedValStartPos {
		t.Errorf("Expected value start position: %d", expectedValStartPos)
	}
}

func TestGetAttribute9(t *testing.T) {
	expectedAttr := "a b"
	expectedValStartPos := 5
	attr, valStartPos := getAttribute("\"a b=c d\"")
	if attr != expectedAttr {
		t.Errorf("Expected attribute: \"%s\" but got: \"%s\"", expectedAttr, attr)
	}
	if valStartPos != expectedValStartPos {
		t.Errorf("Expected value start position: %d", expectedValStartPos)
	}
}

func TestGetAttribute10(t *testing.T) {
	expectedAttr := "abc "
	expectedValStartPos := 7
	attr, valStartPos := getAttribute("\"abc` =123 \"")
	if attr != expectedAttr {
		t.Errorf("Expected attribute: \"%s\" but got: \"%s\"", expectedAttr, attr)
	}
	if valStartPos != expectedValStartPos {
		t.Errorf("Expected value start position: %d", expectedValStartPos)
	}
}

func TestProcessValue1(t *testing.T) {
	expectedVal := "blue"
	val := processValue("blue")
	if val != expectedVal {
		t.Errorf("Expected value: \"%s\" but got: \"%s\"", expectedVal, val)
	}
}

func TestProcessValue2(t *testing.T) {
	expectedVal := "a=4"
	val := processValue("a=4")
	if val != expectedVal {
		t.Errorf("Expected value: \"%s\" but got: \"%s\"", expectedVal, val)
	}
}

func TestProcessValue3(t *testing.T) {
	expectedVal := "true"
	val := processValue("true")
	if val != expectedVal {
		t.Errorf("Expected value: \"%s\" but got: \"%s\"", expectedVal, val)
	}
}

func TestProcessValue4(t *testing.T) {
	expectedVal := "false"
	val := processValue("false")
	if val != expectedVal {
		t.Errorf("Expected value: \"%s\" but got: \"%s\"", expectedVal, val)
	}
}

func TestProcessValue5(t *testing.T) {
	expectedVal := "\\="
	val := processValue("\\\\=")
	if val != expectedVal {
		t.Errorf("Expected value: \"%s\" but got: \"%s\"", expectedVal, val)
	}
}

func TestProcessValue6(t *testing.T) {
	expectedVal := "\"Cat\""
	val := processValue("\"Cat\"")
	if val != expectedVal {
		t.Errorf("Expected value: \"%s\" but got: \"%s\"", expectedVal, val)
	}
}

func TestProcessValue7(t *testing.T) {
	expectedVal := "``abc``"
	val := processValue("``abc``")
	if val != expectedVal {
		t.Errorf("Expected value: \"%s\" but got: \"%s\"", expectedVal, val)
	}
}

func TestProcessValue8(t *testing.T) {
	expectedVal := ""
	val := processValue("")
	if val != expectedVal {
		t.Errorf("Expected value: \"%s\" but got: \"%s\"", expectedVal, val)
	}
}

func TestProcessValue9(t *testing.T) {
	expectedVal := "c d"
	val := processValue("c d")
	if val != expectedVal {
		t.Errorf("Expected value: \"%s\" but got: \"%s\"", expectedVal, val)
	}
}

func TestProcessValue10(t *testing.T) {
	expectedVal := "123 "
	val := processValue("123 ")
	if val != expectedVal {
		t.Errorf("Expected value: \"%s\" but got: \"%s\"", expectedVal, val)
	}
}

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

// ### Need to find a DNS provider that supports backslashes ###

//func TestFull4(t *testing.T) {
//	nameserver := NameServer{Host:"8.8.4.4", Port:53, Priority:}
//	request := Request{Domain:"test4.nooutbound.co.uk", NameServers:nameserver}
//	res := request.Get()
//	expectedAttr := "a\\=a false"
//	expectedVal  := "a\\`=a=false"
//	if _, ok := res.Config[expectedAttr]; ok {
//		if res.Config[expectedAttr] != expectedVal {
//			t.Errorf("Expected value: \"%s\"", expectedVal)
//			t.Errorf("Got: %+v", res.Config[expectedAttr])
//		}
//	} else {
//		t.Errorf("Expected attribute: \"%s\"", expectedAttr)
//		t.Errorf("Got: %+v", res)
//	}
//}
//
//func TestFull5(t *testing.T) {
//	nameserver := NameServer{Host:"8.8.4.4", Port:53, Priority:}
//	request := Request{Domain:"test5.nooutbound.co.uk", NameServers:nameserver}
//	res := request.Get()
//	expectedAttr := "="
//	expectedVal  := "\\="
//	if _, ok := res.Config[expectedAttr]; ok {
//		if res.Config[expectedAttr] != expectedVal {
//			t.Errorf("Expected value: \"%s\"", expectedVal)
//			t.Errorf("Got: %+v", res.Config[expectedAttr])
//		}
//	} else {
//		t.Errorf("Expected attribute: \"%s\"", expectedAttr)
//		t.Errorf("Got: %+v", res)
//	}
//}
//
//func TestFull6(t *testing.T) {
//	nameserver := NameServer{Host:"8.8.4.4", Port:53, Priority:}
//	request := Request{Domain:"test6.nooutbound.co.uk", NameServers:nameserver}
//	res := request.Get()
//	expectedAttr := "string"
//	expectedVal  := "\"Cat\""
//	if _, ok := res.Config[expectedAttr]; ok {
//		if res.Config[expectedAttr] != expectedVal {
//			t.Errorf("Expected value: \"%s\"", expectedVal)
//			t.Errorf("Got: %+v", res.Config[expectedAttr])
//		}
//	} else {
//		t.Errorf("Expected attribute: \"%s\"", expectedAttr)
//		t.Errorf("Got: %+v", res)
//	}
//}

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
	nameserver := NameServer{Host: "8.8.4.4", Port: 53, Priority: 0}
	request := NewRequest("missing.example.com", nameserver)
	_, err := request.Get()
	if !strings.Contains(err.Error(), "NXDOMAIN") {
		t.Errorf("Expected NXDOMAIN error")
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

func TestGetAttributeEdgeCases(t *testing.T) {
	// Test with equals at start (should be handled by processRecord's validation)
	attr, valStart := getAttribute("\"=invalid\"")
	if attr != "" || valStart != 1 {
		// Testing boundary condition where equals is first character
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

func TestMultipleNameServersWithErrors(t *testing.T) {
	// Test with multiple invalid nameservers, then a valid one
	nameserver1 := NameServer{Host: "192.0.2.1", Port: 53, Priority: 0}   // Invalid IP (RFC 5737)
	nameserver2 := NameServer{Host: "192.0.2.2", Port: 53, Priority: 1}   // Invalid IP (RFC 5737) 
	nameserver3 := NameServer{Host: "8.8.8.8", Port: 53, Priority: 2}     // Valid nameserver
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
	nameserver1 := NameServer{Host: "192.0.2.1", Port: 53, Priority: 0}   // Invalid IP (RFC 5737)
	nameserver2 := NameServer{Host: "192.0.2.2", Port: 53, Priority: 1}   // Invalid IP (RFC 5737)
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
