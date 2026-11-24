package server

import (
	"net/http"
	"strconv"
	"strings"
)

// WantsJSON reports whether the client prefers JSON to HTML. Browsers ask for text/html first, and a
// request without Accept gets HTML too, so JSON has to be asked for explicitly.
func WantsJSON(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	if accept == "" {
		return false
	}
	json := preference(accept, "application", "json")
	html := preference(accept, "text", "html")
	if json.q != html.q {
		return json.q > html.q
	}
	// a tie between e.g. application/json and */* goes to the type that was named
	return json.q > 0 && json.specificity > html.specificity
}

type acceptMatch struct {
	q           float64
	specificity int // 2 for type/subtype, 1 for type/*, 0 for */*
}

// preference is how much accept wants the media type, from its most specific matching range.
func preference(accept, typ, subtype string) acceptMatch {
	best := acceptMatch{specificity: -1}
	for _, part := range strings.Split(accept, ",") {
		params := strings.Split(part, ";")
		mediaType := strings.ToLower(strings.TrimSpace(params[0]))
		t, s, ok := strings.Cut(mediaType, "/")
		if !ok {
			continue
		}

		var specificity int
		switch {
		case t == typ && s == subtype:
			specificity = 2
		case t == typ && s == "*":
			specificity = 1
		case t == "*" && s == "*":
			specificity = 0
		default:
			continue
		}
		if specificity > best.specificity {
			best = acceptMatch{q: quality(params[1:]), specificity: specificity}
		}
	}
	if best.specificity < 0 {
		return acceptMatch{}
	}
	return best
}

func quality(params []string) float64 {
	for _, param := range params {
		key, value, _ := strings.Cut(strings.TrimSpace(param), "=")
		if strings.EqualFold(key, "q") {
			q, err := strconv.ParseFloat(value, 64)
			if err != nil || q < 0 || q > 1 {
				return 0
			}
			return q
		}
	}
	return 1
}
