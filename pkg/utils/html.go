package utils

import (
	"strings"

	"github.com/kennygrant/sanitize"
)

var (
	IgnoreTags = []string{"title", "script", "style", "iframe", "frame", "frameset", "noframes", "noembed", "embed", "applet", "object", "base"}	
	DefaultTags = []string{"h1", "h2", "h3", "h4", "h5", "h6", "div", "hr", "span", "p", "br", "b", "i", "strong", "em", "ol", "ul", "li", "a", "img", "pre", "code", "blockquote", "article", "section"}
	DefaultAttributes = []string{"id", "class", "src", "href", "title", "alt", "name", "rel"}
)

func CleanHtmlBody(s string) string {
	newmessage, err := sanitize.HTMLAllowing(s, DefaultTags, DefaultAttributes)
	if err != nil {
		panic(err)
	}
	newmessage = strings.Replace(newmessage, " class=\"MsoNormal\"", "", -1)
	return newmessage
}

func RemoveAllTags(s string) string {
	newmessage, err := sanitize.HTMLAllowing(s, []string{}, []string{})
	if err != nil {
		panic(err)
	}
	return newmessage
}