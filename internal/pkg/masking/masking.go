package masking

import "strings"

// MaskName masks a name: returns first 2 chars + "***".
// If name is shorter than 2 chars, returns "***".
func MaskName(name string) string {
	if len(name) == 0 {
		return ""
	}
	runes := []rune(name)
	if len(runes) <= 2 {
		return "***"
	}
	return string(runes[:2]) + "***"
}

// MaskEmail masks an email: returns first 2 chars of local part + "***@" + domain.
// If the email has no "@", returns "***".
func MaskEmail(email string) string {
	if len(email) == 0 {
		return ""
	}
	at := strings.Index(email, "@")
	if at < 0 {
		return "***"
	}
	local := email[:at]
	domain := email[at+1:]
	runes := []rune(local)
	if len(runes) <= 2 {
		return "***@" + domain
	}
	return string(runes[:2]) + "***@" + domain
}

// MaskPhone masks a phone number: returns first 2 digits + "****" + last 4 digits.
// If shorter than 6 chars, returns "***".
func MaskPhone(phone string) string {
	if len(phone) == 0 {
		return ""
	}
	runes := []rune(phone)
	if len(runes) < 6 {
		return "***"
	}
	return string(runes[:2]) + "****" + string(runes[len(runes)-4:])
}
