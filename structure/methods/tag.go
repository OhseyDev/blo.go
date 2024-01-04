package methods

import (
	"github.com/OhseyDev/gospirit/slug"
	"github.com/OhseyDev/gospirit/structure"
	"strings"
)

func GenerateTagsFromCommaString(input string) []structure.Tag {
	output := make([]structure.Tag, 0)
	tags := strings.Split(input, ",")
	for index, _ := range tags {
		tags[index] = strings.TrimSpace(tags[index])
	}
	for _, tag := range tags {
		if tag != "" {
			output = append(output, structure.Tag{Name: []byte(tag), Slug: slug.Generate(tag, "tags")})
		}
	}
	return output
}
