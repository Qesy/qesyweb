// Package validate provides small, reusable input validators.
package validate

import (
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Qesy/QesyWeb/internal/infra/apperror"
)

var (
	emailDomainPattern = regexp.MustCompile(`^[^\s@]+\.[^\s@]+$`)
	phonePattern       = regexp.MustCompile(`^\+?[1-9][0-9]{7,14}$`)
	chinaMobilePattern = regexp.MustCompile(`^1[3-9][0-9]{9}$`)
	domainLabelPattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
)

// PasswordPolicy describes password rules. Zero values use DefaultPasswordPolicy.
type PasswordPolicy struct {
	MinLength      int
	MaxLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
	NoWhitespace   bool
}

var DefaultPasswordPolicy = PasswordPolicy{
	MinLength:      8,
	MaxLength:      72,
	RequireUpper:   true,
	RequireLower:   true,
	RequireNumber:  true,
	RequireSpecial: true,
	NoWhitespace:   true,
}

func Required(value string) bool {
	return strings.TrimSpace(value) != ""
}

func Same(first, second string) bool {
	return first == second
}

// Email validates a conventional mailbox address and requires a dotted domain.
func Email(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsAny(value, "\r\n") {
		return false
	}
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value {
		return false
	}
	at := strings.LastIndexByte(value, '@')
	return at > 0 && emailDomainPattern.MatchString(value[at+1:])
}

// Phone accepts E.164-like numbers (8-15 digits) and China mainland mobiles.
// Formatting spaces, hyphens and parentheses are ignored.
func Phone(value string) bool {
	normalized := normalizePhone(value)
	return phonePattern.MatchString(normalized)
}

func ChinaMobile(value string) bool {
	normalized := normalizePhone(value)
	return chinaMobilePattern.MatchString(normalized)
}

func PasswordLength(value string, min, max int) bool {
	length := utf8.RuneCountInString(value)
	if min > 0 && length < min {
		return false
	}
	return max <= 0 || length <= max
}

// PasswordComplexity checks the four common character classes. It does not
// enforce length; use PasswordLength or ValidatePassword for the full policy.
func PasswordComplexity(value string) bool {
	var upper, lower, number, special bool
	for _, char := range value {
		switch {
		case unicode.IsUpper(char):
			upper = true
		case unicode.IsLower(char):
			lower = true
		case unicode.IsDigit(char):
			number = true
		case !unicode.IsSpace(char):
			special = true
		}
	}
	return upper && lower && number && special
}

func ValidatePassword(value string, policy PasswordPolicy) error {
	policy = applyPasswordDefaults(policy)
	length := utf8.RuneCountInString(value)
	if length < policy.MinLength {
		return apperror.New("error.common.password.min_length", nil, map[string]any{"min": policy.MinLength})
	}
	if policy.MaxLength > 0 && length > policy.MaxLength {
		return apperror.New("error.common.password.max_length", nil, map[string]any{"max": policy.MaxLength})
	}

	var upper, lower, number, special bool
	for _, char := range value {
		switch {
		case unicode.IsSpace(char):
			if policy.NoWhitespace {
				return apperror.New("error.common.password.not_contain_whitespace", nil)
			}
		case unicode.IsUpper(char):
			upper = true
		case unicode.IsLower(char):
			lower = true
		case unicode.IsDigit(char):
			number = true
		default:
			special = true
		}
	}
	if policy.RequireUpper && !upper {
		return apperror.New("error.common.password.must_contain_uppercase_letter", nil)
	}
	if policy.RequireLower && !lower {
		return apperror.New("error.common.password.must_contain_lowercase_letter", nil)
	}
	if policy.RequireNumber && !number {
		return apperror.New("error.common.password.must_contain_number", nil)
	}
	if policy.RequireSpecial && !special {
		return apperror.New("error.common.password.must_contain_special_character", nil)
	}
	return nil
}

func URL(value string) bool {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	return err == nil && parsed.Scheme != "" && parsed.Host != ""
}

// Domain validates a DNS domain name without performing a DNS lookup.
// An optional http:// or https:// prefix is accepted for form inputs.
func Domain(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.TrimPrefix(value, "https://")
	value = strings.TrimPrefix(value, "http://")
	if value == "" || strings.ContainsAny(value, "\r\n\t /:@") {
		return false
	}
	if net.ParseIP(value) != nil || len(value) > 253 {
		return false
	}
	value = strings.TrimSuffix(value, ".")
	parts := strings.Split(value, ".")
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		if len(part) == 0 || len(part) > 63 || !domainLabelPattern.MatchString(part) {
			return false
		}
	}
	return true
}
func IPv4(value string) bool {
	ip := net.ParseIP(strings.TrimSpace(value))
	return ip != nil && ip.To4() != nil
}

func applyPasswordDefaults(policy PasswordPolicy) PasswordPolicy {
	if policy.MinLength == 0 && policy.MaxLength == 0 && !policy.RequireUpper && !policy.RequireLower && !policy.RequireNumber && !policy.RequireSpecial && !policy.NoWhitespace {
		return DefaultPasswordPolicy
	}
	return policy
}

func normalizePhone(value string) string {
	value = strings.TrimSpace(value)
	var builder strings.Builder
	for _, char := range value {
		switch {
		case char >= '0' && char <= '9':
			builder.WriteRune(char)
		case char == '+' && builder.Len() == 0:
			builder.WriteRune(char)
		case char == ' ' || char == '-' || char == '(' || char == ')':
		default:
			return ""
		}
	}
	return builder.String()
}
