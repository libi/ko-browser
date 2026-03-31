package selector

import (
	"fmt"
	"strconv"
	"strings"
)

type Kind string

const (
	KindDisplayID Kind = "display_id"
	KindCSS       Kind = "css"
	KindXPath     Kind = "xpath"
)

type Selector struct {
	Kind      Kind
	Raw       string
	DisplayID int
	Query     string
}

func Parse(input string) (Selector, error) {
	value := strings.TrimSpace(input)
	if value == "" {
		return Selector{}, fmt.Errorf("selector cannot be empty")
	}

	if displayID, err := strconv.Atoi(value); err == nil {
		if displayID <= 0 {
			return Selector{}, fmt.Errorf("display id must be positive")
		}
		return Selector{Kind: KindDisplayID, Raw: value, DisplayID: displayID}, nil
	}

	if strings.HasPrefix(value, "xpath=") {
		query := strings.TrimSpace(strings.TrimPrefix(value, "xpath="))
		if query == "" {
			return Selector{}, fmt.Errorf("xpath selector cannot be empty")
		}
		return Selector{Kind: KindXPath, Raw: value, Query: query}, nil
	}

	if strings.HasPrefix(value, "//") || strings.HasPrefix(value, "(") {
		return Selector{Kind: KindXPath, Raw: value, Query: value}, nil
	}

	if strings.HasPrefix(value, "css=") {
		query := strings.TrimSpace(strings.TrimPrefix(value, "css="))
		if query == "" {
			return Selector{}, fmt.Errorf("css selector cannot be empty")
		}
		return Selector{Kind: KindCSS, Raw: value, Query: query}, nil
	}

	return Selector{Kind: KindCSS, Raw: value, Query: value}, nil
}
