package uhttp

import (
	"encoding/xml"
	"strings"
)

// xmlMap implements xml.Unmarshaler and can unmarshal arbitrary XML into a
// map[string]any structure. Leaf elements become string values, and elements
// with children become nested maps.
type xmlMap struct {
	data any
}

func (x *xmlMap) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	result, err := unmarshalXMLElement(d)
	if err != nil {
		return err
	}
	x.data = result
	return nil
}

// unmarshalXMLElement reads tokens from the decoder for the current element
// (after its start element has been consumed) until the matching end element.
// It returns a map[string]any if there are child elements, a []map[string]any
// if there are duplicate child element names, or a string if the element
// contains only text.
func unmarshalXMLElement(d *xml.Decoder) (any, error) {
	type entry struct {
		key   string
		value any
	}
	var entries []entry
	seen := make(map[string]bool)
	hasDuplicates := false
	var charData strings.Builder

	for {
		t, err := d.Token()
		if err != nil {
			return nil, err
		}
		switch tt := t.(type) {
		case xml.StartElement:
			child, err := unmarshalXMLElement(d)
			if err != nil {
				return nil, err
			}
			key := tt.Name.Local
			if seen[key] {
				hasDuplicates = true
			}
			seen[key] = true
			entries = append(entries, entry{key: key, value: child})
		case xml.CharData:
			_, err := charData.Write(tt)
			if err != nil {
				return nil, err
			}
		case xml.EndElement:
			if len(entries) == 0 {
				text := strings.TrimSpace(charData.String())
				if text == "" {
					return make(map[string]any), nil
				}
				return text, nil
			}
			if hasDuplicates {
				result := make([]map[string]any, 0, len(entries))
				for _, e := range entries {
					result = append(result, map[string]any{e.key: e.value})
				}
				return result, nil
			}
			result := make(map[string]any, len(entries))
			for _, e := range entries {
				result[e.key] = e.value
			}
			return result, nil
		}
	}
}
