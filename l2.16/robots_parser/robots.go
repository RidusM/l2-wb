package robots

import (
	"bufio"
	"fmt"
	"strings"
	"time"
)

type RobotsRule struct {
	UserAgent  string
	Disallow   []string
	Allow      []string
	CrawlDelay time.Duration
}

type RobotsParser struct {
	Rules       []RobotsRule
	DefaultRule *RobotsRule
}

func NewRobotsParser(content string) *RobotsParser {
	parser := &RobotsParser{
		Rules: make([]RobotsRule, 0),
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	var currentRule *RobotsRule

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch key {
		case "user-agent":
			if currentRule != nil {
				parser.Rules = append(parser.Rules, *currentRule)
				if currentRule.UserAgent == "*" {
					parser.DefaultRule = currentRule
				}
			}
			currentRule = &RobotsRule{
				UserAgent:  value,
				Disallow:   make([]string, 0),
				Allow:      make([]string, 0),
				CrawlDelay: 0,
			}

		case "disallow":
			if currentRule != nil && value != "" {
				currentRule.Disallow = append(currentRule.Disallow, value)
			}

		case "allow":
			if currentRule != nil && value != "" {
				currentRule.Allow = append(currentRule.Allow, value)
			}

		case "crawl-delay":
			if currentRule != nil {
				var delay float64
				fmt.Sscanf(value, "%f", &delay)
				currentRule.CrawlDelay = time.Duration(delay * float64(time.Second))
			}
		}
	}

	if currentRule != nil {
		parser.Rules = append(parser.Rules, *currentRule)
		if currentRule.UserAgent == "*" {
			parser.DefaultRule = currentRule
		}
	}

	return parser
}

func (rp *RobotsParser) IsAllowed(urlPath, userAgent string) bool {
	var rule *RobotsRule
	userAgent = strings.ToLower(userAgent)

	for i := range rp.Rules {
		if strings.ToLower(rp.Rules[i].UserAgent) == userAgent {
			rule = &rp.Rules[i]
			break
		}
	}

	if rule == nil {
		rule = rp.DefaultRule
	}

	if rule == nil {
		return true
	}

	for _, allowPath := range rule.Allow {
		if strings.HasPrefix(urlPath, allowPath) {
			return true
		}
	}

	for _, disallowPath := range rule.Disallow {
		if disallowPath == "/" {
			return false
		}
		if strings.HasPrefix(urlPath, disallowPath) {
			return false
		}
	}

	return true
}

func (rp *RobotsParser) GetCrawlDelay(userAgent string) time.Duration {
	userAgent = strings.ToLower(userAgent)

	for i := range rp.Rules {
		if strings.ToLower(rp.Rules[i].UserAgent) == userAgent {
			return rp.Rules[i].CrawlDelay
		}
	}

	if rp.DefaultRule != nil {
		return rp.DefaultRule.CrawlDelay
	}

	return 0
}