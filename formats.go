package jsonschema

import (
	"errors"
	"net"
	"net/mail"
	"net/url"
	"regexp/syntax"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/kaptinlin/jsonpointer"
	"github.com/yosida95/uritemplate/v3"
	"golang.org/x/net/idna"
)

var errInvalidResourceIdentifier = errors.New("invalid resource identifier")

var idnProfile = idna.New(
	idna.ValidateForRegistration(),
	idna.StrictDomainName(true),
	idna.VerifyDNSLength(true),
	idna.BidiRule(),
)

// Formats is a registry of functions, which know how to validate
// a specific format.
//
// New Formats can be registered by adding to this map. Key is format name,
// value is function that knows how to validate that format.
var Formats = map[string]func(any) bool{
	"date-time":             IsDateTime,
	"date":                  IsDate,
	"time":                  IsTime,
	"duration":              IsDuration,
	"period":                IsPeriod,
	"hostname":              IsHostname,
	"email":                 IsEmail,
	"idn-hostname":          IsIDNHostname,
	"idn-email":             IsIDNEmail,
	"ip-address":            IsIPV4,
	"ipv4":                  IsIPV4,
	"ipv6":                  IsIPV6,
	"uri":                   IsURI,
	"iri":                   IsIRI,
	"uri-reference":         IsURIReference,
	"uriref":                IsURIReference,
	"iri-reference":         IsIRIReference,
	"uri-template":          IsURITemplate,
	"json-pointer":          IsJSONPointer,
	"relative-json-pointer": IsRelativeJSONPointer,
	"uuid":                  IsUUID,
	"regex":                 IsRegex,
}

// IsDateTime tells whether given string is a valid date representation
// as defined by RFC 3339, section 5.6.
//
// see https://datatracker.ietf.org/doc/html/rfc3339#section-5.6, for details
func IsDateTime(v any) bool {
	s, ok := v.(string)
	if !ok {
		return true
	}
	if len(s) < 20 { // yyyy-mm-ddThh:mm:ssZ
		return false
	}
	if s[10] != 'T' && s[10] != 't' {
		return false
	}
	return IsDate(s[:10]) && IsTime(s[11:])
}

// IsDate tells whether given string is a valid full-date production
// as defined by RFC 3339, section 5.6.
//
// see https://datatracker.ietf.org/doc/html/rfc3339#section-5.6, for details
func IsDate(v any) bool {
	s, ok := v.(string)
	if !ok {
		return true
	}
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

// IsTime tells whether given string is a valid full-time production
// as defined by RFC 3339, section 5.6.
//
// see https://datatracker.ietf.org/doc/html/rfc3339#section-5.6, for details
func IsTime(v any) bool {
	str, ok := v.(string)
	if !ok {
		return true
	}

	// golang time package does not support leap seconds.
	// so we are parsing it manually here.

	// Expect hh:mm:ss at the start.
	if len(str) < 9 || str[2] != ':' || str[5] != ':' {
		return false
	}
	isInRange := func(str string, minVal, maxVal int) (int, bool) {
		n, err := strconv.Atoi(str)
		if err != nil {
			return 0, false
		}
		if n < minVal || n > maxVal {
			return 0, false
		}
		return n, true
	}
	var h, m, s int
	if h, ok = isInRange(str[0:2], 0, 23); !ok {
		return false
	}
	if m, ok = isInRange(str[3:5], 0, 59); !ok {
		return false
	}
	if s, ok = isInRange(str[6:8], 0, 60); !ok {
		return false
	}
	str = str[8:]

	// parse secfrac if present
	if str[0] == '.' {
		// dot following more than one digit
		str = str[1:]
		var numDigits int
		for str != "" {
			if str[0] < '0' || str[0] > '9' {
				break
			}
			numDigits++
			str = str[1:]
		}
		if numDigits == 0 {
			return false
		}
	}

	if len(str) == 0 {
		return false
	}

	if str[0] == 'z' || str[0] == 'Z' {
		if len(str) != 1 {
			return false
		}
	} else {
		// time-numoffset
		// +hh:mm
		// 012345
		if len(str) != 6 || str[3] != ':' {
			return false
		}

		var sign int
		switch str[0] {
		case '+':
			sign = -1
		case '-':
			sign = +1
		default:
			return false
		}

		var zh, zm int
		ok := false
		if zh, ok = isInRange(str[1:3], 0, 23); !ok {
			return false
		}
		if zm, ok = isInRange(str[4:6], 0, 59); !ok {
			return false
		}

		// apply timezone offset
		hm := (h*60 + m) + sign*(zh*60+zm)
		if hm < 0 {
			hm += 24 * 60
		}
		h, m = hm/60, hm%60
	}

	// check leapsecond
	if s == 60 { // leap second
		if h != 23 || m != 59 {
			return false
		}
	}

	return true
}

// IsDuration tells whether given string is a valid duration format
// from the ISO 8601 ABNF as given in Appendix A of RFC 3339.
//
// see https://datatracker.ietf.org/doc/html/rfc3339#appendix-A, for details
func IsDuration(v any) bool {
	s, ok := v.(string)
	if !ok {
		return true
	}
	if len(s) == 0 || s[0] != 'P' {
		return false
	}
	s = s[1:]
	parseUnits := func() (units string, ok bool) {
		for len(s) > 0 && s[0] != 'T' {
			digits := false
			for len(s) != 0 {
				if s[0] < '0' || s[0] > '9' {
					break
				}
				digits = true
				s = s[1:]
			}
			if !digits || len(s) == 0 {
				return units, false
			}
			units += s[:1]
			s = s[1:]
		}
		return units, true
	}
	units, ok := parseUnits()
	if !ok {
		return false
	}
	if units == "W" {
		return len(s) == 0 // P_W
	}
	if len(units) > 0 {
		if !containsOrderedUnits(units, "YMD") {
			return false
		}
		if len(s) == 0 {
			return true // "P" dur-date
		}
	}
	if len(s) == 0 || s[0] != 'T' {
		return false
	}
	s = s[1:]
	units, ok = parseUnits()
	return ok && len(s) == 0 && len(units) > 0 && containsOrderedUnits(units, "HMS")
}

func containsOrderedUnits(units, allowed string) bool {
	return strings.Contains(allowed, units)
}

// IsPeriod tells whether given string is a valid period format
// from the ISO 8601 ABNF as given in Appendix A of RFC 3339.
//
// see https://datatracker.ietf.org/doc/html/rfc3339#appendix-A, for details
func IsPeriod(v any) bool {
	s, ok := v.(string)
	if !ok {
		return true
	}
	before, after, ok0 := strings.Cut(s, "/")
	if !ok0 {
		return false
	}
	start, end := before, after
	if IsDateTime(start) {
		return IsDateTime(end) || IsDuration(end)
	}
	return IsDuration(start) && IsDateTime(end)
}

// IsHostname tells whether given string is a valid representation
// for an Internet host name, as defined by RFC 1034 section 3.1 and
// RFC 1123 section 2.1.
//
// See https://en.wikipedia.org/wiki/Hostname#Restrictions_on_valid_host_names, for details.
func IsHostname(v any) bool {
	s, ok := v.(string)
	if !ok {
		return true
	}
	// entire hostname (including the delimiting dots but not a trailing dot) has a maximum of 253 ASCII characters
	s = strings.TrimSuffix(s, ".")
	if len(s) > 253 {
		return false
	}

	// Hostnames are composed of series of labels concatenated with dots, as are all domain names
	for label := range strings.SplitSeq(s, ".") {
		// Each label must be from 1 to 63 characters long
		if labelLen := len(label); labelLen < 1 || labelLen > 63 {
			return false
		}

		// labels must not start with a hyphen
		// RFC 1123 section 2.1: restriction on the first character
		// is relaxed to allow either a letter or a digit
		if first := label[0]; first == '-' {
			return false
		}

		// must not end with a hyphen
		if label[len(label)-1] == '-' {
			return false
		}

		// labels may contain only the ASCII letters 'a' through 'z' (in a case-insensitive manner),
		// the digits '0' through '9', and the hyphen ('-')
		for _, c := range label {
			if valid := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || (c == '-'); !valid {
				return false
			}
		}
	}

	return true
}

// IsEmail tells whether the given string has valid ASCII email address syntax.
func IsEmail(v any) bool {
	s, ok := v.(string)
	if !ok {
		return true
	}
	return isEmail(s, false)
}

// IsIDNHostname tells whether the given string has valid internationalized
// hostname syntax.
func IsIDNHostname(v any) bool {
	s, ok := v.(string)
	if !ok {
		return true
	}
	if !utf8.ValidString(s) {
		return false
	}

	s = strings.TrimSuffix(s, ".")
	if s == "" {
		return false
	}

	ascii, err := idnProfile.ToASCII(s)
	return err == nil && IsHostname(ascii)
}

// IsIDNEmail tells whether the given string has valid internationalized email
// address syntax.
func IsIDNEmail(v any) bool {
	s, ok := v.(string)
	if !ok {
		return true
	}
	return isEmail(s, true)
}

func isEmail(s string, international bool) bool {
	if !utf8.ValidString(s) || len(s) > 254 {
		return false
	}
	if !international && !isASCII(s) {
		return false
	}

	at := strings.LastIndexByte(s, '@')
	if at <= 0 || at == len(s)-1 || at > 64 {
		return false
	}
	domain := s[at+1:]
	switch {
	case len(domain) >= 2 && domain[0] == '[' && domain[len(domain)-1] == ']':
		ip := domain[1 : len(domain)-1]
		if ipv6, ok := strings.CutPrefix(ip, "IPv6:"); ok {
			if !IsIPV6(ipv6) {
				return false
			}
		} else if !IsIPV4(ip) {
			return false
		}
	case international:
		if !IsIDNHostname(domain) {
			return false
		}
	default:
		if !IsHostname(domain) {
			return false
		}
	}

	_, err := mail.ParseAddress(s)
	return err == nil
}

func isASCII(s string) bool {
	for i := range len(s) {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}

// IsIPV4 tells whether given string is a valid representation of an IPv4 address
// according to the "dotted-quad" ABNF syntax as defined in RFC 2673, section 3.2.
func IsIPV4(v any) bool {
	s, ok := v.(string)
	if !ok {
		return true
	}
	groups := strings.Split(s, ".")
	if len(groups) != 4 {
		return false
	}
	for _, group := range groups {
		n, err := strconv.Atoi(group)
		if err != nil {
			return false
		}
		if n < 0 || n > 255 {
			return false
		}
		if len(group) > 1 && group[0] == '0' {
			return false // leading zeroes should be rejected, as they are treated as octals
		}
	}
	return true
}

// IsIPV6 tells whether given string is a valid representation of an IPv6 address
// as defined in RFC 2373, section 2.2.
func IsIPV6(v any) bool {
	s, ok := v.(string)
	if !ok {
		return true
	}
	if !strings.Contains(s, ":") {
		return false
	}
	return net.ParseIP(s) != nil
}

// IsURI tells whether given string is valid URI, according to RFC 3986.
func IsURI(v any) bool {
	s, ok := v.(string)
	if !ok {
		return true
	}
	u, err := parseResourceIdentifier(s, false)
	return err == nil && u.IsAbs()
}

// IsIRI tells whether the given string is an Internationalized Resource
// Identifier according to RFC 3987.
func IsIRI(v any) bool {
	s, ok := v.(string)
	if !ok {
		return true
	}
	u, err := parseResourceIdentifier(s, true)
	return err == nil && u.IsAbs()
}

func parseResourceIdentifier(s string, allowIRI bool) (*url.URL, error) {
	if !validIdentifierCharacters(s, allowIRI) {
		return nil, errInvalidResourceIdentifier
	}
	parseInput, hasIPvFuture := normalizeIPvFutureHost(s)
	u, err := url.Parse(parseInput)
	if err != nil {
		return nil, err
	}

	if hasIPvFuture {
		return u, nil
	}
	hostname := u.Hostname()
	bracketedHost := strings.HasPrefix(u.Host, "[") && strings.Contains(u.Host, "]")
	if bracketedHost {
		if !IsIPV6(hostname) && !isIPvFuture(hostname) {
			return nil, ErrInvalidIPv6
		}
	} else if strings.Contains(hostname, ":") {
		return nil, ErrIPv6AddressFormat
	}
	return u, nil
}

func normalizeIPvFutureHost(s string) (string, bool) {
	start := strings.IndexByte(s, '[')
	if start < 0 {
		return s, false
	}
	endOffset := strings.IndexByte(s[start+1:], ']')
	if endOffset < 0 {
		return s, false
	}
	end := start + endOffset + 1
	if !isIPvFuture(s[start+1 : end]) {
		return s, false
	}
	return s[:start+1] + "::1" + s[end:], true
}

func validIdentifierCharacters(s string, allowIRI bool) bool {
	if !utf8.ValidString(s) {
		return false
	}

	hierPart, fragment, hasFragment := strings.Cut(s, "#")
	if hasFragment && (!validIdentifierComponent(fragment, ":@/?", allowIRI, false) || strings.Contains(fragment, "#")) {
		return false
	}
	hierPart, query, hasQuery := strings.Cut(hierPart, "?")
	if hasQuery && !validIdentifierComponent(query, ":@/?", allowIRI, true) {
		return false
	}

	location := hierPart
	colon := strings.IndexByte(hierPart, ':')
	slash := strings.IndexByte(hierPart, '/')
	if colon >= 0 && (slash < 0 || colon < slash) {
		if !validScheme(hierPart[:colon]) {
			return false
		}
		location = hierPart[colon+1:]
	}

	if strings.HasPrefix(location, "//") {
		authority, path, hasPath := strings.Cut(strings.TrimPrefix(location, "//"), "/")
		if !validIdentifierComponent(authority, ":@[]", allowIRI, false) {
			return false
		}
		if hasPath {
			location = "/" + path
		} else {
			location = ""
		}
	}
	return validIdentifierComponent(location, ":@/", allowIRI, false)
}

func validScheme(s string) bool {
	if s == "" || !isASCIIAlpha(rune(s[0])) {
		return false
	}
	for i := 1; i < len(s); i++ {
		if !isASCIIAlpha(rune(s[i])) && !isASCIIDigit(rune(s[i])) && !strings.ContainsRune("+.-", rune(s[i])) {
			return false
		}
	}
	return true
}

func validIdentifierComponent(s, extraASCII string, allowIRI, allowPrivate bool) bool {
	const commonASCII = "-._~!$&'()*+,;="
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == '%' {
			if i+2 >= len(s) || !isHex(s[i+1]) || !isHex(s[i+2]) {
				return false
			}
			i += 3
			continue
		}

		if r < utf8.RuneSelf {
			if !isASCIIAlpha(r) && !isASCIIDigit(r) &&
				!strings.ContainsRune(commonASCII, r) && !strings.ContainsRune(extraASCII, r) {
				return false
			}
			i += size
			continue
		}

		if allowIRI && (isUCSChar(r) || allowPrivate && isIPrivate(r)) {
			i += size
			continue
		}
		return false
	}
	return true
}

func isASCIIAlpha(b rune) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
}

func isASCIIDigit(b rune) bool {
	return b >= '0' && b <= '9'
}

func isHex(b byte) bool {
	return b >= '0' && b <= '9' || b >= 'a' && b <= 'f' || b >= 'A' && b <= 'F'
}

func isUCSChar(r rune) bool {
	return r >= 0xA0 && r <= 0xD7FF ||
		r >= 0xF900 && r <= 0xFDCF ||
		r >= 0xFDF0 && r <= 0xFFEF ||
		r >= 0x10000 && r <= 0x1FFFD ||
		r >= 0x20000 && r <= 0x2FFFD ||
		r >= 0x30000 && r <= 0x3FFFD ||
		r >= 0x40000 && r <= 0x4FFFD ||
		r >= 0x50000 && r <= 0x5FFFD ||
		r >= 0x60000 && r <= 0x6FFFD ||
		r >= 0x70000 && r <= 0x7FFFD ||
		r >= 0x80000 && r <= 0x8FFFD ||
		r >= 0x90000 && r <= 0x9FFFD ||
		r >= 0xA0000 && r <= 0xAFFFD ||
		r >= 0xB0000 && r <= 0xBFFFD ||
		r >= 0xC0000 && r <= 0xCFFFD ||
		r >= 0xD0000 && r <= 0xDFFFD ||
		r >= 0xE1000 && r <= 0xEFFFD
}

func isIPrivate(r rune) bool {
	return r >= 0xE000 && r <= 0xF8FF ||
		r >= 0xF0000 && r <= 0xFFFFD ||
		r >= 0x100000 && r <= 0x10FFFD
}

func isIPvFuture(s string) bool {
	if len(s) < 4 || s[0] != 'v' && s[0] != 'V' {
		return false
	}
	dot := strings.IndexByte(s, '.')
	if dot < 2 || dot == len(s)-1 {
		return false
	}
	for i := 1; i < dot; i++ {
		if !isHex(s[i]) {
			return false
		}
	}
	const addressCharacters = "-._~!$&'()*+,;=:"
	for i := dot + 1; i < len(s); i++ {
		if !isASCIIAlpha(rune(s[i])) && !isASCIIDigit(rune(s[i])) &&
			!strings.ContainsRune(addressCharacters, rune(s[i])) {
			return false
		}
	}
	return true
}

// IsURIReference tells whether given string is a valid URI Reference
// (either a URI or a relative-reference), according to RFC 3986.
func IsURIReference(v any) bool {
	s, ok := v.(string)
	if !ok {
		return true
	}
	_, err := parseResourceIdentifier(s, false)
	return err == nil
}

// IsIRIReference tells whether the given string is an IRI reference according
// to RFC 3987.
func IsIRIReference(v any) bool {
	s, ok := v.(string)
	if !ok {
		return true
	}
	_, err := parseResourceIdentifier(s, true)
	return err == nil
}

// IsURITemplate tells whether given string is a valid URI Template
// according to RFC6570.
func IsURITemplate(v any) bool {
	s, ok := v.(string)
	if !ok {
		return true
	}
	_, err := uritemplate.New(s)
	return err == nil
}

// IsJSONPointer tells whether given string is a valid JSON Pointer.
//
// Note: It returns false for JSON Pointer URI fragments.
func IsJSONPointer(v any) bool {
	s, ok := v.(string)
	if !ok {
		return true
	}
	// Empty string is a valid JSON Pointer (points to the whole document)
	if s == "" {
		return true
	}
	_, err := jsonpointer.Parse(s)
	return err == nil
}

// IsRelativeJSONPointer tells whether given string is a valid Relative JSON Pointer.
//
// see https://tools.ietf.org/html/draft-handrews-relative-json-pointer-01#section-3
func IsRelativeJSONPointer(v any) bool {
	s, ok := v.(string)
	if !ok {
		return true
	}
	if s == "" {
		return false
	}
	switch {
	case s[0] == '0':
		s = s[1:]
	case s[0] >= '0' && s[0] <= '9':
		for s != "" && s[0] >= '0' && s[0] <= '9' {
			s = s[1:]
		}
	default:
		return false
	}
	return s == "#" || IsJSONPointer(s)
}

// IsUUID tells whether given string is a valid uuid format
// as specified in RFC4122.
//
// see https://datatracker.ietf.org/doc/html/rfc4122#page-4, for details
func IsUUID(v any) bool {
	s, ok := v.(string)
	if !ok {
		return true
	}
	parseHex := func(n int) bool {
		for n > 0 {
			if len(s) == 0 {
				return false
			}
			hex := (s[0] >= '0' && s[0] <= '9') || (s[0] >= 'a' && s[0] <= 'f') || (s[0] >= 'A' && s[0] <= 'F')
			if !hex {
				return false
			}
			s = s[1:]
			n--
		}
		return true
	}
	groups := []int{8, 4, 4, 4, 12}
	for i, numDigits := range groups {
		if !parseHex(numDigits) {
			return false
		}
		if i == len(groups)-1 {
			break
		}
		if len(s) == 0 || s[0] != '-' {
			return false
		}
		s = s[1:]
	}
	return len(s) == 0
}

// IsRegex tells whether the given string is a regular expression supported by
// the interoperable JSON Schema subset and Go's RE2 syntax.
func IsRegex(v any) bool {
	pattern, ok := v.(string)
	if !ok {
		return true
	}
	for {
		_, err := syntax.Parse(pattern, syntax.Perl)
		if err == nil {
			return true
		}

		var syntaxErr *syntax.Error
		if !errors.As(err, &syntaxErr) || syntaxErr.Code != syntax.ErrInvalidRepeatSize {
			return false
		}

		normalized, ok := normalizeCountedRepeat(pattern, syntaxErr.Expr)
		if !ok {
			return false
		}
		pattern = normalized
	}
}

// normalizeCountedRepeat removes regexp/syntax's 1000-copy implementation
// limit. The next parse still decides every other aspect of regex syntax.
func normalizeCountedRepeat(pattern, repeat string) (string, bool) {
	if len(repeat) < 3 || repeat[0] != '{' || repeat[len(repeat)-1] != '}' {
		return "", false
	}

	min, max, hasMax := strings.Cut(repeat[1:len(repeat)-1], ",")
	if !decimalDigits(min) || hasMax && max != "" && !decimalDigits(max) {
		return "", false
	}
	if hasMax && max != "" && decimalLess(max, min) {
		return "", false
	}

	normalized := strings.ReplaceAll(pattern, repeat, "{1}")
	return normalized, normalized != pattern
}

func decimalDigits(s string) bool {
	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return s != ""
}

func decimalLess(a, b string) bool {
	a = strings.TrimLeft(a, "0")
	b = strings.TrimLeft(b, "0")
	if len(a) != len(b) {
		return len(a) < len(b)
	}
	return a < b
}
