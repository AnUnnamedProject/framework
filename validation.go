// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import (
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type (
	// Rule defines a single rule to check.
	Rule struct {
		Field        string
		Label        string
		Rule         string
		RuleParams   []string
		ErrorMessage string
	}

	// Validator contains the methods for error checking. Values are retreived automatically from Context.
	Validator struct {
		ctx   *Context
		rules []Rule
	}
)

// DefaultErrorMessages contains the rule's default error messages.
var DefaultErrorMessages = map[string]string{
	"Required":     "can not be empty",
	"Min":          "minimum value is %d",
	"MinLength":    "minimum length is %d",
	"Max":          "maximum value is %d",
	"MaxLength":    "maximum length is %d",
	"Exact":        "required value is %d",
	"ExactLength":  "required length is %d",
	"Alpha":        "must contain a valid alpha characters",
	"AlphaNumeric": "must contain a valid alpha numeric characters",
	"Numeric":      "must be a valid numeric value",
	"Float":        "must be a valid float value",
	"URL":          "must be a valid URL",
	"Email":        "must be a valid email address",
	"IP":           "must be a valid ip address",
	"Base64":       "must be a valid base64",
}

// NewValidator returns a new Validator instance
func NewValidator(ctx *Context) *Validator {
	return &Validator{ctx: ctx}
}

// Rules adds multiple rules to the validator.
func (v *Validator) Rules(field string, label string, rules []string, errorMessages []string) {
	for i, rule := range rules {
		v.Rule(field, label, rule, errorMessages[i])
	}
}

// Rule adds a new rule to based on field.
// If the rule needs to pass parameters (f.e. MaxLength of 100) put the value into brackets: MaxLength[100]
func (v *Validator) Rule(field string, label string, rule string, errorMessage string) {
	var params []string

	// If rule contains brackets, split rule params
	if strings.Contains(rule, "[") {
		re := regexp.MustCompile(`(?s)\[(.*)\]`)
		m := re.FindAllStringSubmatch(rule, -1)
		if m != nil {
			params = strings.Split(m[0][1], ",")
		}

		rule = rule[:strings.Index(rule, "[")]
	}

	// No message, sets rule's default message
	if errorMessage == "" {
		errorMessage = DefaultErrorMessages[rule]
	}

	v.rules = append(v.rules, Rule{
		Field:        field,
		Label:        label,
		Rule:         rule,
		RuleParams:   params,
		ErrorMessage: errorMessage,
	})
}

// ValidForm check if all rules from current Request.Form are valid, otherwise return the rule error message.
func (v *Validator) ValidForm() error {
	values := make(map[string]interface{})

	for key, value := range v.ctx.Request.Form {
		values[key] = value[0]
	}

	return v.Valid(values)
}

// ValidJSON check if all rules from current Request.JSON are valid, otherwise return the rule error message.
func (v *Validator) ValidJSON() error {
	values := make(map[string]interface{})

	if strings.Contains(v.ctx.Request.Header.Get("Content-Type"), "application/json") {
		for key, value := range v.ctx.Request.JSON.m {
			if value != nil {
				switch value.data.(type) {
				case bool:
					values[key] = fmt.Sprintf("%t", value.Bool())
				case float64:
					values[key] = fmt.Sprintf("%f", value.Float64())
				case string:
					values[key] = value.String()
				default:
					values[key] = ""
				}
			}
		}
	}

	return v.Valid(values)
}

// Valid check if all rules are valid from provided map[string]string, otherwise return the rule error message.
func (v *Validator) Valid(values map[string]interface{}) error {
	for _, rule := range v.rules {
		method := reflect.ValueOf(v).MethodByName(rule.Rule)

		// Method does not exists ?
		if !method.IsValid() {
			Log.Error(fmt.Errorf("invalid rule: %s", rule.Rule))
			continue
		}

		// Few or too many arguments ?
		if len(rule.RuleParams) < method.Type().NumIn()-1 || len(rule.RuleParams) > method.Type().NumIn()-1 {
			continue
		}

		if values[rule.Field] == nil {
			return fmt.Errorf("%s", rule.ErrorMessage)
		}

		var rst []reflect.Value
		switch method.Type().In(0).String() {
		case "interface {}":
			rst = append(rst, reflect.ValueOf(values[rule.Field]))
		case "int":
			rst = append(rst, reflect.ValueOf(interfaceToInt(values[rule.Field])))
		case "string":
			rst = append(rst, reflect.ValueOf(interfaceToString(values[rule.Field])))
		}

		typ := method.Type()
		for i := 0; i < typ.NumIn(); i++ {
			switch method.Type().In(i).String() {
			case "int":
				param, _ := strconv.Atoi(rule.RuleParams[i-1])
				rst = append(rst, reflect.ValueOf(param))
			}
		}

		res := method.Call(rst)[0]
		if !res.Bool() {
			if len(rule.RuleParams) > 0 {
				return fmt.Errorf(rule.ErrorMessage, rule.RuleParams)
			}
			return fmt.Errorf("%s", rule.ErrorMessage)
		}
	}

	return nil
}

func interfaceToInt(value interface{}) int {
	if value == nil {
		return 0
	}
	if _, ok := value.(string); ok {
		return 0
	}
	if _, ok := value.(bool); ok {
		return 0
	}
	if i, ok := value.(int); ok {
		return i
	}
	if i, ok := value.(uint); ok {
		return int(i)
	}
	if i, ok := value.(int8); ok {
		return int(i)
	}
	if i, ok := value.(uint8); ok {
		return int(i)
	}
	if i, ok := value.(int16); ok {
		return int(i)
	}
	if i, ok := value.(uint16); ok {
		return int(i)
	}
	if i, ok := value.(uint32); ok {
		return int(i)
	}
	if i, ok := value.(int32); ok {
		return int(i)
	}
	if i, ok := value.(int64); ok {
		return int(i)
	}
	if i, ok := value.(uint64); ok {
		return int(i)
	}
	return 0
}

func interfaceToString(value interface{}) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	if _, ok := value.(bool); ok {
		return ""
	}
	if _, ok := value.(int); ok {
		return ""
	}
	if _, ok := value.(uint); ok {
		return ""
	}
	if _, ok := value.(int8); ok {
		return ""
	}
	if _, ok := value.(uint8); ok {
		return ""
	}
	if _, ok := value.(int16); ok {
		return ""
	}
	if _, ok := value.(uint16); ok {
		return ""
	}
	if _, ok := value.(uint32); ok {
		return ""
	}
	if _, ok := value.(int32); ok {
		return ""
	}
	if _, ok := value.(int64); ok {
		return ""
	}
	if _, ok := value.(uint64); ok {
		return ""
	}
	return ""
}

// Required checks if value is not nil, empty, false or 0
func (v *Validator) Required(value interface{}) bool {
	if value == nil {
		return false
	}
	if str, ok := value.(string); ok {
		return str != ""
	}
	if _, ok := value.(bool); ok {
		return true
	}
	if i, ok := value.(int); ok {
		return i != 0
	}
	if i, ok := value.(uint); ok {
		return i != 0
	}
	if i, ok := value.(int8); ok {
		return i != 0
	}
	if i, ok := value.(uint8); ok {
		return i != 0
	}
	if i, ok := value.(int16); ok {
		return i != 0
	}
	if i, ok := value.(uint16); ok {
		return i != 0
	}
	if i, ok := value.(uint32); ok {
		return i != 0
	}
	if i, ok := value.(int32); ok {
		return i != 0
	}
	if i, ok := value.(int64); ok {
		return i != 0
	}
	if i, ok := value.(uint64); ok {
		return i != 0
	}

	return true
}

// Min check if int value is equal or greater than max
func (v *Validator) Min(value int, min int) bool {
	return value >= min
}

// MinLength check if value is equal or greater than min
func (v *Validator) MinLength(value string, min int) bool {
	return len(value) >= min
}

// Max check if int value is lower or equal than max
func (v *Validator) Max(value int, max int) bool {
	return value <= max
}

// MaxLength check if value length is lower or equal than max
func (v *Validator) MaxLength(value string, max int) bool {
	return len(value) <= max
}

// Exact check if value is equal to compare
func (v *Validator) Exact(value int, compare int) bool {
	return value == compare
}

// ExactLength check if value length is equal to length
func (v *Validator) ExactLength(value string, length int) bool {
	return len(value) == length
}

// Alpha check if value contains only alpha (a-z) values
func (v *Validator) Alpha(value string) bool {
	if value == "" {
		return true
	}

	return regexp.MustCompile("^[a-zA-Z]+$").MatchString(value)
}

// AlphaNumeric check if value contains only alphanumeric values
func (v *Validator) AlphaNumeric(value string) bool {
	if value == "" {
		return true
	}

	return regexp.MustCompile("^[a-zA-Z0-9]+$").MatchString(value)
}

// Numeric check if value contains only digits
func (v *Validator) Numeric(value string) bool {
	if value == "" {
		return true
	}

	return regexp.MustCompile("^[-+]?[0-9]+$").MatchString(value)
}

// Float check if value contains only digits
func (v *Validator) Float(value string) bool {
	if value == "" {
		return true
	}

	return regexp.MustCompile("^(?:[-+]?(?:[0-9]+))?(?:\\.[0-9]*)?(?:[eE][\\+\\-]?(?:[0-9]+))?$").MatchString(value)
}

// URL check if value is a valid URL address
func (v *Validator) URL(value string) bool {
	// If empty, len is greater than 2048 (IE) or string is too short
	if value == "" || len(value) > 2048 || len(value) <= 3 {
		return false
	}

	// Parse URL string
	u, err := url.Parse(value)
	if err != nil {
		return false
	}

	// Host can't start with a dot
	if strings.HasPrefix(u.Host, ".") {
		return false
	}

	// Host is empty, check if Path contains the url
	if u.Host == "" && (u.Path != "" && !strings.Contains(u.Path, ".")) {
		return false
	}

	return regexp.MustCompile(`^((ftp|https?):\/\/)?(\S+(:\S*)?@)?((([1-9]\d?|1\d\d|2[01]\d|22[0-3])(\.(1?\d{1,2}|2[0-4]\d|25[0-5])){2}(?:\.([0-9]\d?|1\d\d|2[0-4]\d|25[0-4]))|(([a-zA-Z0-9]([a-zA-Z0-9-]+)?[a-zA-Z0-9]([-\.][a-zA-Z0-9]+)*)|((www\.)?))?(([a-zA-Z\x{00a1}-\x{ffff}0-9]+-?-?)*[a-zA-Z\x{00a1}-\x{ffff}0-9]+)(?:\.([a-zA-Z\x{00a1}-\x{ffff}]{1,}))?))(:(\d{1,5}))?((\/|\?|#)[^\s]*)?$`).MatchString(value)
}

// Email check if value is a valid RFC5322 email address
func (v *Validator) Email(value string) bool {
	if value == "" {
		return false
	}

	if _, err := mail.ParseAddress(value); err != nil {
		return false
	}

	return true
}

// IP check if value is a valid IP address
func (v *Validator) IP(value string) bool {
	if value == "" {
		return false
	}

	check := net.ParseIP(value)

	// Not a valid IPv4 or IPv6 address
	if check.To4() == nil && check.To16() == nil {
		return false
	}

	return true
}

// Base64 check if value is a valid Base64 string
func (v *Validator) Base64(value string) bool {
	if value == "" {
		return false
	}

	return regexp.MustCompile("^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$").MatchString(value)
}
