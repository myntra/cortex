package main

import (
	"log"
	"regexp"
	"strings"
)

func regexMatcher(ruleType string, eventType string) bool {
	regexList := regexMaker(ruleType)
	for regex := range regexList {
		var ruleRegex = regexp.MustCompile(regexList[regex])
		if ruleRegex.MatchString(eventType) {
			return true
		}
	}
	return false
}

func regexMaker(ruleType string) []string {

	regexList := []string{}
	ruleRegx := "^" + ruleType + "$"
	regexList = append(regexList, ruleRegx)
	var firstRegex, secondRegex string
	for pos, char := range ruleType {

		if int(char) == 46 {
			nextDotPos := strings.Index(ruleType[pos+1:len(ruleType)], ".")
			if nextDotPos == -1 {
				firstRegex = "^" + ruleType[0:pos] + "\\." + "[a-zA-Z0-9]*" + "$"
				secondRegex = ruleType[0:pos] + "." + ruleType[pos+1:len(ruleType)] + "$"
				regexList = append(regexList, firstRegex)
				regexList = append(regexList, secondRegex)
				continue
			} else {
				firstRegex = "^" + ruleType[0:pos] + "\\." + "." + "*" + ruleType[pos+nextDotPos+1:len(ruleType)] + "$"
				secondRegex = "^" + ruleType[0:pos] + "." + ruleType[pos+1:pos+nextDotPos+1] + "*" + ruleType[pos+nextDotPos+1:len(ruleType)] + "$"
				regexList = append(regexList, firstRegex)
				regexList = append(regexList, secondRegex)
				continue
			}
		}
	}
	return regexList
}

func main() {

	res := regexMatcher("metrics.myntra1.com", "metrics.myntra.com")
	log.Println("Match : ", res)
}
