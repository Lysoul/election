package util

import "regexp"

const DateOfBirthRegex = `(Jan(uary)?|Feb(ruary)?|Mar(ch)?|Apr(il)?|May|Jun(e)?|Jul(y)?|Aug(ust)?|Sep(tember)?|Oct(ober)?|Nov(ember)?|Dec(ember)?)\s+\d{1,2},\s+\d{4}`

//IsDobRegex checks string is matched with date of birth regular expression
func IsDateOfBirth(s string) bool {
	matched, err := regexp.MatchString(DateOfBirthRegex, s)

	if err != nil {
		return false
	}

	return matched
}
