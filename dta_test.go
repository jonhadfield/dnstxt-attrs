package dta

import (
	"testing"
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
	request := Request{Domain: "test1.nooutbound.co.uk", NameServers: []NameServer{nameserver}}
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
	request := Request{Domain: "test2.nooutbound.co.uk", NameServers: []NameServer{nameserver}}
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
	request := Request{Domain: "test3.nooutbound.co.uk", NameServers: []NameServer{nameserver}}
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
//	request := Request{Domain:"test4.nooutbound.co.uk", NameServers:[]NameServer{nameserver}}
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
//	request := Request{Domain:"test5.nooutbound.co.uk", NameServers:[]NameServer{nameserver}}
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
//	request := Request{Domain:"test6.nooutbound.co.uk", NameServers:[]NameServer{nameserver}}
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
	request := Request{Domain: "test7.nooutbound.co.uk", NameServers: []NameServer{nameserver}}
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
	request := Request{Domain: "test8.nooutbound.co.uk", NameServers: []NameServer{nameserver}}
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
	nameserver1 := NameServer{Host: "8.8.4.4", Port: 53, Priority: 1}
	nameserver2 := NameServer{Host: "8.8.8.8", Port: 53, Priority: 0}
	request := Request{Domain: "test9.nooutbound.co.uk", NameServers: []NameServer{nameserver1, nameserver2}}
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
	nameserver1 := NameServer{Host: "8.8.4.4", Port: 53, Priority: 1}
	nameserver2 := NameServer{Host: "8.8.8.8", Port: 53, Priority: 0}
	request := Request{Domain: "test10.nooutbound.co.uk", NameServers: []NameServer{nameserver1, nameserver2}}
	res, _ := request.Get()
	expected := "abc "
	if _, ok := res.Config[expected]; ok != true {
		t.Errorf("Expected attribute: \"%s\"", expected)
		t.Errorf("Got: %+v", res)
	}
}

func TestInvalidDomain(t *testing.T) {
	nameserver := NameServer{Host: "8.8.4.4", Port: 53, Priority: 1}
	request := Request{Domain: "missing.example.com", NameServers: []NameServer{nameserver}}
	_, err := request.Get()
	if err == nil {
		t.Errorf("Expected missing domain error")
	}

}
