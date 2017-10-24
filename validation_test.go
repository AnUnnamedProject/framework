// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRule
// TODO:
func TestRule(t *testing.T) {

}

// TestRules
// TODO:
func TestRules(t *testing.T) {

}

func TestValidForm(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	c := NewContext(NewResponseWriter(httptest.NewRecorder()), &Request{Request: r})
	c.ParseForm()

	c.Request.Form.Add("required", "1")

	valid := NewValidator(c)
	valid.Rule("required", "required", "Required", "")
	err := valid.ValidForm()
	if err != nil {
		t.Error(err)
	}
}

func TestValidJSON(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Add("Content-Type", "application/json")

	c := NewContext(NewResponseWriter(httptest.NewRecorder()), &Request{Request: r})

	c.Request.JSON = &JSONData{
		m: make(map[string]*JSONValue),
	}

	c.Request.JSON.Set("required", "1")
	c.Request.JSON.Set("required2", 1)
	c.Request.JSON.Set("required4", 1.2)
	c.Request.JSON.Set("required3", true)

	valid := NewValidator(c)
	valid.Rule("required", "required", "Required", "")
	err := valid.ValidJSON()
	if err != nil {
		t.Error(err)
	}
}

func TestRequired(t *testing.T) {
	t.Parallel()

	valid := &Validator{}

	var tests = []struct {
		value    interface{}
		expected bool
	}{
		{nil, false},
		{"", false},
		{"ok", true},
		{true, true},
		{false, true},
		{0, false},
		{1, true},
	}
	for _, test := range tests {
		actual := valid.Required(test.value)
		if actual != test.expected {
			t.Errorf("Required(%q) is expected to be %v, got %v", test.value, test.expected, actual)
		}
	}
}

func TestMin(t *testing.T) {
	t.Parallel()

	valid := &Validator{}

	var tests = []struct {
		value    int
		length   int
		expected bool
	}{
		{-1, 0, false},
		{1, 0, true},
	}
	for _, test := range tests {
		actual := valid.Min(test.value, test.length)
		if actual != test.expected {
			t.Errorf("Min(%q) is expected to be %v, got %v", test.value, test.expected, actual)
		}
	}
}

func TestMinLength(t *testing.T) {
	t.Parallel()

	valid := &Validator{}

	var tests = []struct {
		value    string
		length   int
		expected bool
	}{
		{"", 1, false},
		{"ok", 2, true},
	}
	for _, test := range tests {
		actual := valid.MinLength(test.value, test.length)
		if actual != test.expected {
			t.Errorf("MinLength(%q) is expected to be %v, got %v", test.value, test.expected, actual)
		}
	}
}

func TestMax(t *testing.T) {
	t.Parallel()

	valid := &Validator{}

	var tests = []struct {
		value    int
		length   int
		expected bool
	}{
		{1, 0, false},
		{-1, 0, true},
	}
	for _, test := range tests {
		actual := valid.Max(test.value, test.length)
		if actual != test.expected {
			t.Errorf("Max(%q) is expected to be %v, got %v", test.value, test.expected, actual)
		}
	}
}

func TestMaxLength(t *testing.T) {
	t.Parallel()

	valid := &Validator{}

	var tests = []struct {
		value    string
		length   int
		expected bool
	}{
		{"", 1, true},
		{"ok", 1, false},
		{"ok", 2, true},
	}
	for _, test := range tests {
		actual := valid.MaxLength(test.value, test.length)
		if actual != test.expected {
			t.Errorf("MaxLength(%q) is expected to be %v, got %v", test.value, test.expected, actual)
		}
	}
}

func TestExact(t *testing.T) {
	t.Parallel()

	valid := &Validator{}

	var tests = []struct {
		value    int
		length   int
		expected bool
	}{
		{0, 1, false},
		{1, 1, true},
	}
	for _, test := range tests {
		actual := valid.Exact(test.value, test.length)
		if actual != test.expected {
			t.Errorf("Exact(%q) is expected to be %v, got %v", test.value, test.expected, actual)
		}
	}
}

func TestExactLength(t *testing.T) {
	t.Parallel()

	valid := &Validator{}

	var tests = []struct {
		value    string
		length   int
		expected bool
	}{
		{"", 1, false},
		{"OK", 2, true},
	}
	for _, test := range tests {
		actual := valid.ExactLength(test.value, test.length)
		if actual != test.expected {
			t.Errorf("ExactLength(%q) is expected to be %v, got %v", test.value, test.expected, actual)
		}
	}
}

func TestAlpha(t *testing.T) {
	t.Parallel()

	valid := &Validator{}

	var tests = []struct {
		value    string
		expected bool
	}{
		{"\n", false},
		{"\r", false},
		{"", true},
		{"abc1", false},
		{"abc", true},
		{"ABC", true},
		{"123", false},
		{"0", false},
		{"123.123", false},
		{" ", false},
		{".", false},
		{"-", false},
	}
	for _, test := range tests {
		actual := valid.Alpha(test.value)
		if actual != test.expected {
			t.Errorf("Alpha(%q) is expected to be %v, got %v", test.value, test.expected, actual)
		}
	}
}

func TestAlphaNumeric(t *testing.T) {
	t.Parallel()

	valid := &Validator{}

	var tests = []struct {
		value    string
		expected bool
	}{
		{"\n", false},
		{"\r", false},
		{"", true},
		{"abc1", true},
		{"abc", true},
		{"ABC", true},
		{"123", true},
		{"0", true},
		{"123.123", false},
		{" ", false},
		{".", false},
		{"-", false},
	}
	for _, test := range tests {
		actual := valid.AlphaNumeric(test.value)
		if actual != test.expected {
			t.Errorf("AlphaNumeric(%q) is expected to be %v, got %v", test.value, test.expected, actual)
		}
	}
}

func TestNumeric(t *testing.T) {
	t.Parallel()

	valid := &Validator{}

	var tests = []struct {
		value    string
		expected bool
	}{
		{"\n", false},
		{"\r", false},
		{"", true},
		{"abc1", false},
		{"abc", false},
		{"ABC", false},
		{"123", true},
		{"0", true},
		{"123.123", false},
		{" ", false},
		{".", false},
		{"-", false},
	}
	for _, test := range tests {
		actual := valid.Numeric(test.value)
		if actual != test.expected {
			t.Errorf("Numeric(%q) is expected to be %v, got %v", test.value, test.expected, actual)
		}
	}
}

func TestFloat(t *testing.T) {
	t.Parallel()

	valid := &Validator{}

	var tests = []struct {
		value    string
		expected bool
	}{
		{"\n", false},
		{"\r", false},
		{"", true},
		{"abc1", false},
		{"abc", false},
		{"ABC", false},
		{"123", true},
		{"0", true},
		{"123.123", true},
		{" ", false},
		{".", true},
		{"-", false},
		{"-20", true},
	}
	for _, test := range tests {
		actual := valid.Float(test.value)
		if actual != test.expected {
			t.Errorf("Float(%q) is expected to be %v, got %v", test.value, test.expected, actual)
		}
	}
}

func TestURL(t *testing.T) {
	t.Parallel()

	valid := &Validator{}

	var tests = []struct {
		value    string
		expected bool
	}{
		{"", false},
		{"http://example.com#", true},
		{"http://example.com", true},
		{"https://example.com", true},
		{"example.com", true},
		{"http://example.info/", true},
		{"http://example.org/", true},
		{"http://example.ORG", true},
		{"http://example.org:80/", true},
		{"ftp://example.com/", true},
		{"ftp.example.com", true},
		{"http://user:pass@www.example.com/", true},
		{"http://user:pass@www.example.com/path", true},
		{"http://127.0.0.1/", true},
		{"http://localhost:8000/", true},
		{"http://example.com/?foo=bar&bar=foo", true},
		{"invalid.", false},
		{".com", false},
		{"mailto:me@example.com", true},
		{"/local/dir", false},
		{"http://example .org", false},
		{"example", false},
		{"http://.example.com", false},
		{"http://cant-end-with-hyphen-.example.com", false},
		{"http://-cant-start-with-hyphen.example.com", false},
		{"http://www.domain-can-contain-dashes.com", true},
	}
	for _, test := range tests {
		actual := valid.URL(test.value)
		if actual != test.expected {
			t.Errorf("URL(%q) is expected to be %v, got %v", test.value, test.expected, actual)
		}
	}
}

func TestEmail(t *testing.T) {
	t.Parallel()

	valid := &Validator{}

	var tests = []struct {
		value    string
		expected bool
	}{
		{"", false},
		{"x@x.x", true},
		{"me@example.com", true},
		{"me+alias@example.com", true},
		{"me@example.co.uk", true},
		{"me@example.travel", true},
		{"invalid@", false},
		{"invalid.com", false},
		{"@invalid.com", false},
	}
	for _, test := range tests {
		actual := valid.Email(test.value)
		if actual != test.expected {
			t.Errorf("Email(%q) is expected to be %v, got %v", test.value, test.expected, actual)
		}
	}
}

func TestIP(t *testing.T) {
	t.Parallel()

	valid := &Validator{}

	var tests = []struct {
		value    string
		expected bool
	}{
		{"", false},
		{"127.0.0.1", true},
		{"::1", true},
		{"255.255.255.256", false},
	}
	for _, test := range tests {
		actual := valid.IP(test.value)
		if actual != test.expected {
			t.Errorf("IP(%q) is expected to be %v, got %v", test.value, test.expected, actual)
		}
	}
}

func TestBase64(t *testing.T) {
	t.Parallel()

	valid := &Validator{}

	var tests = []struct {
		value    string
		expected bool
	}{
		{"", false},
		{"12345", false},
		{"SGVsbG8=", true},
	}
	for _, test := range tests {
		actual := valid.Base64(test.value)
		if actual != test.expected {
			t.Errorf("Base64(%q) is expected to be %v, got %v", test.value, test.expected, actual)
		}
	}
}
