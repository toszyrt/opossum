package cmd

import "regexp"

func LooksLikeIpv4(ip string) bool {
	pattern := `^(\d{1,3}\.){3}\d{1,3}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(ip)
}
