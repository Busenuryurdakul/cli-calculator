package bookmark

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	ErrNotFound  = errors.New("bookmark not found")
	ErrInvalidID = errors.New("invalid id")
	ErrConflict  = errors.New("bookmark conflict")
)

var idPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// ValidationError is a client input problem (HTTP 400).
type ValidationError struct {
	Msg string
}

func (e *ValidationError) Error() string {
	return e.Msg
}

func ParseID(raw string) (string, error) {
	if !idPattern.MatchString(raw) {
		return "", ErrInvalidID
	}
	return raw, nil
}

func ValidateTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", &ValidationError{Msg: "title is required"}
	}
	if utf8.RuneCountInString(title) > MaxTitleLen {
		return "", &ValidationError{Msg: fmt.Sprintf("title must be at most %d characters", MaxTitleLen)}
	}
	return title, nil
}

func ValidateURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", &ValidationError{Msg: "url is required"}
	}
	if utf8.RuneCountInString(raw) > MaxURLLen {
		return "", &ValidationError{Msg: fmt.Sprintf("url must be at most %d characters", MaxURLLen)}
	}
	u, err := url.Parse(raw)
	if err != nil || !u.IsAbs() || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", &ValidationError{Msg: "url must be an absolute http or https URL"}
	}
	return raw, nil
}

// ValidateTags trims each tag. Empty or whitespace-only tags are rejected.
func ValidateTags(tags []string) ([]string, error) {
	if len(tags) == 0 {
		return []string{}, nil
	}
	if len(tags) > MaxTags {
		return nil, &ValidationError{Msg: fmt.Sprintf("at most %d tags", MaxTags)}
	}
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			return nil, &ValidationError{Msg: "tag must not be empty"}
		}
		if utf8.RuneCountInString(tag) > MaxTagLen {
			return nil, &ValidationError{Msg: fmt.Sprintf("tag must be at most %d characters", MaxTagLen)}
		}
		out = append(out, tag)
	}
	return out, nil
}

func cloneTags(tags []string) []string {
	if len(tags) == 0 {
		return []string{}
	}
	out := make([]string, len(tags))
	copy(out, tags)
	return out
}

func cloneBookmark(b Bookmark) Bookmark {
	b.Tags = cloneTags(b.Tags)
	return b
}
