package validate

import (
	"regexp"
	"sync"
)

var (
	emailRegex  *regexp.Regexp
	phoneRegex  *regexp.Regexp
	idCardRegex *regexp.Regexp
	regexOnce   sync.Once
)

func compileRegexes() {
	emailRegex = regexp.MustCompile(`^[0-9a-z][_.0-9a-z-]{0,31}@([0-9a-z][0-9a-z-]{0,30}[0-9a-z]\.){1,4}[a-z]{2,4}$`)
	phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)
	idCardRegex = regexp.MustCompile(`(^\d{15}$)|(^\d{18}$)|(^\d{17}(\d|X|x)$)`)
}

// IsEmail 验证是否为电子邮件地址
func IsEmail(email string) bool {
	regexOnce.Do(compileRegexes)
	return emailRegex.MatchString(email)
}

// IsPhone 验证是否为手机号码
func IsPhone(phone string) bool {
	regexOnce.Do(compileRegexes)
	return phoneRegex.MatchString(phone)
}

// IsIDCard 验证是否为身份证号
func IsIDCard(id string) bool {
	regexOnce.Do(compileRegexes)
	return idCardRegex.MatchString(id)
}
