package gstructs

import (
	"github.com/sagoo-cloud/nexframe/utils/tag"
	"strings"
)

// TagJsonName returns the `json` tag name string of the field.
func (f *Field) TagJsonName() string {
	if jsonTag := f.Tag(tag.Json); jsonTag != "" {
		return strings.Split(jsonTag, ",")[0]
	}
	return ""
}

// TagDefault returns the most commonly used tag `default/d` value of the field.
func (f *Field) TagDefault() string {
	v := f.Tag(tag.Default)
	if v == "" {
		v = f.Tag(tag.DefaultShort)
	}
	return v
}

// TagParam returns the most commonly used tag `param/p` value of the field.
func (f *Field) TagParam() string {
	v := f.Tag(tag.Param)
	if v == "" {
		v = f.Tag(tag.ParamShort)
	}
	return v
}

// TagValid returns the most commonly used tag `valid/v` value of the field.
func (f *Field) TagValid() string {
	v := f.Tag(tag.Valid)
	if v == "" {
		v = f.Tag(tag.ValidShort)
	}
	return v
}

// TagDescription returns the most commonly used tag `description/des/dc` value of the field.
func (f *Field) TagDescription() string {
	v := f.Tag(tag.Description)
	if v == "" {
		v = f.Tag(tag.DescriptionShort)
	}
	if v == "" {
		v = f.Tag(tag.DescriptionShort2)
	}
	return v
}

// TagSummary returns the most commonly used tag `summary/sum/sm` value of the field.
func (f *Field) TagSummary() string {
	v := f.Tag(tag.Summary)
	if v == "" {
		v = f.Tag(tag.SummaryShort)
	}
	if v == "" {
		v = f.Tag(tag.SummaryShort2)
	}
	return v
}

// TagAdditional returns the most commonly used tag `additional/ad` value of the field.
func (f *Field) TagAdditional() string {
	v := f.Tag(tag.Additional)
	if v == "" {
		v = f.Tag(tag.AdditionalShort)
	}
	return v
}

// TagExample returns the most commonly used tag `example/eg` value of the field.
func (f *Field) TagExample() string {
	v := f.Tag(tag.Example)
	if v == "" {
		v = f.Tag(tag.ExampleShort)
	}
	return v
}

// TagIn returns the most commonly used tag `in` value of the field.
func (f *Field) TagIn() string {
	v := f.Tag(tag.In)
	return v
}

// TagPriorityName checks and returns tag name that matches the name item in `gtag.StructTagPriority`.
// It or else returns attribute field Name if it doesn't have a tag name by `gtag.StructsTagPriority`.
func (f *Field) TagPriorityName() string {
	var name = f.Name()
	for _, tagName := range tag.StructTagPriority {
		if tagValue := f.Tag(tagName); tagValue != "" {
			name = tagValue
			break
		}
	}
	return name
}
