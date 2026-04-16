package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/imesart/apple-ads-cli/internal/columnname"
	"github.com/imesart/apple-ads-cli/internal/fieldmeta"
	"gopkg.in/yaml.v3"
)

// OrderedField is a selected output field with a stable output name and header.
type OrderedField struct {
	Key    string
	Header string
}

// OrderedData is a filtered object or collection with stable field order.
type OrderedData struct {
	Fields []OrderedField
	Rows   [][]any
	Single bool
}

type selectedPath struct {
	Key    string
	Header string
	Actual []string
}

// SelectFields filters structured output to the requested fields.
func SelectFields(data any, fieldList string, entityIDName string) (any, error) {
	fieldSpecs := parseFieldList(fieldList)
	if len(fieldSpecs) == 0 {
		return data, nil
	}

	root, err := toGeneric(data)
	if err != nil {
		return nil, err
	}
	root = unwrapDataEnvelope(root)

	switch v := root.(type) {
	case map[string]any:
		paths, err := resolveSelectedPaths(v, fieldSpecs, entityIDName)
		if err != nil {
			return nil, err
		}
		return OrderedData{
			Fields: orderedFields(paths),
			Rows:   [][]any{buildOrderedRow(v, paths)},
			Single: true,
		}, nil
	case []any:
		if len(v) == 0 {
			return OrderedData{
				Fields: unresolvedFields(fieldSpecs),
				Rows:   nil,
				Single: false,
			}, nil
		}
		sample, ok := v[0].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("--fields requires object or collection output")
		}
		paths, err := resolveSelectedPaths(sample, fieldSpecs, entityIDName)
		if err != nil {
			return nil, err
		}
		rows := make([][]any, 0, len(v))
		for _, item := range v {
			obj, ok := item.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("--fields requires object or collection output")
			}
			rows = append(rows, buildOrderedRow(obj, paths))
		}
		return OrderedData{
			Fields: orderedFields(paths),
			Rows:   rows,
			Single: false,
		}, nil
	default:
		return nil, fmt.Errorf("--fields requires object or collection output")
	}
}

func parseFieldList(fieldList string) []string {
	parts := strings.Split(fieldList, ",")
	fields := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			fields = append(fields, part)
		}
	}
	return fields
}

func toGeneric(data any) (any, error) {
	if raw, ok := data.(json.RawMessage); ok {
		var out any
		if err := UnmarshalUseNumber(raw, &out); err != nil {
			return nil, err
		}
		return out, nil
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var out any
	if err := UnmarshalUseNumber(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func unwrapDataEnvelope(v any) any {
	obj, ok := v.(map[string]any)
	if !ok {
		return v
	}
	data, ok := obj["data"]
	if !ok {
		return v
	}
	if data == nil {
		return []any{}
	}
	switch data.(type) {
	case map[string]any, []any:
		return data
	default:
		return v
	}
}

func resolveSelectedPaths(sample map[string]any, fieldSpecs []string, entityIDName string) ([]selectedPath, error) {
	kind := fieldmeta.KindFromEntityIDName(entityIDName)
	paths := make([]selectedPath, 0, len(fieldSpecs))
	for _, spec := range fieldSpecs {
		if spec == "" {
			continue
		}
		lookup := spec
		if target, ok := fieldmeta.AliasTarget(kind, spec); ok {
			lookup = target
		}
		actual, err := resolvePath(sample, strings.Split(lookup, "."))
		if err != nil {
			return nil, fmt.Errorf("unknown field %q", spec)
		}
		paths = append(paths, selectedPath{
			Key:    strings.TrimSpace(spec),
			Header: renderPathHeader(strings.Split(strings.TrimSpace(spec), ".")),
			Actual: actual,
		})
	}
	return paths, nil
}

func resolvePath(obj map[string]any, segments []string) ([]string, error) {
	actual := make([]string, 0, len(segments))
	current := obj
	for i, seg := range segments {
		key, ok := matchKey(current, seg)
		if !ok {
			return nil, fmt.Errorf("unknown field")
		}
		actual = append(actual, key)
		if i == len(segments)-1 {
			break
		}
		next, ok := current[key].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("unknown field")
		}
		current = next
	}
	return actual, nil
}

func matchKey(obj map[string]any, segment string) (string, bool) {
	want := normalizeSegment(segment)
	for key := range obj {
		if normalizeSegment(key) == want {
			return key, true
		}
	}
	return "", false
}

func normalizeSegment(s string) string {
	return columnname.Compact(columnname.FromField(strings.TrimSpace(s)))
}

func orderedFields(paths []selectedPath) []OrderedField {
	fields := make([]OrderedField, 0, len(paths))
	for _, path := range paths {
		fields = append(fields, OrderedField{
			Key:    path.Key,
			Header: path.Header,
		})
	}
	return fields
}

func unresolvedFields(fieldSpecs []string) []OrderedField {
	fields := make([]OrderedField, 0, len(fieldSpecs))
	for _, spec := range fieldSpecs {
		actual := strings.Split(spec, ".")
		fields = append(fields, OrderedField{
			Key:    spec,
			Header: renderPathHeader(actual),
		})
	}
	return fields
}

func renderPathHeader(actual []string) string {
	parts := make([]string, 0, len(actual))
	for _, part := range actual {
		parts = append(parts, columnname.FromField(part))
	}
	return strings.Join(parts, ".")
}

func buildOrderedRow(obj map[string]any, paths []selectedPath) []any {
	row := make([]any, 0, len(paths))
	for _, path := range paths {
		row = append(row, extractPathValue(obj, path.Actual))
	}
	return row
}

func extractPathValue(obj map[string]any, actual []string) any {
	var current any = obj
	for _, seg := range actual {
		next, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = next[seg]
	}
	return current
}

func (o OrderedData) MarshalJSON() ([]byte, error) {
	if o.Single {
		if len(o.Rows) == 0 {
			return []byte("null"), nil
		}
		return marshalOrderedRowJSON(o.Fields, o.Rows[0])
	}

	var buf bytes.Buffer
	buf.WriteByte('[')
	for i, row := range o.Rows {
		if i > 0 {
			buf.WriteByte(',')
		}
		raw, err := marshalOrderedRowJSON(o.Fields, row)
		if err != nil {
			return nil, err
		}
		buf.Write(raw)
	}
	buf.WriteByte(']')
	return buf.Bytes(), nil
}

func marshalOrderedRowJSON(fields []OrderedField, row []any) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, field := range fields {
		if i > 0 {
			buf.WriteByte(',')
		}
		key, err := json.Marshal(field.Key)
		if err != nil {
			return nil, err
		}
		val, err := json.Marshal(row[i])
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteByte(':')
		buf.Write(val)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (o OrderedData) MarshalYAML() (any, error) {
	if o.Single {
		if len(o.Rows) == 0 {
			return nil, nil
		}
		return orderedYAMLNode(o.Fields, o.Rows[0]), nil
	}

	seq := &yaml.Node{Kind: yaml.SequenceNode}
	for _, row := range o.Rows {
		seq.Content = append(seq.Content, orderedYAMLNode(o.Fields, row))
	}
	return seq, nil
}

func orderedYAMLNode(fields []OrderedField, row []any) *yaml.Node {
	node := &yaml.Node{Kind: yaml.MappingNode}
	for i, field := range fields {
		node.Content = append(node.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: field.Key})
		node.Content = append(node.Content, yamlNodeFor(row[i]))
	}
	return node
}

func yamlNodeFor(v any) *yaml.Node {
	var node yaml.Node
	if err := node.Encode(v); err != nil {
		return &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%v", v)}
	}
	return &node
}
