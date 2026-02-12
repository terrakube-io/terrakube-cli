package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/kataras/tablewriter"
	"gopkg.in/yaml.v3"
)

// Render writes data to w in the specified format.
// Supported formats: json, yaml, table, tsv, none.
func Render(w io.Writer, data any, format string) error {
	switch format {
	case "json":
		return renderJSON(w, data)
	case "yaml":
		return renderYAML(w, data)
	case "table":
		return renderTable(w, data)
	case "tsv":
		return renderTSV(w, data)
	case "none":
		return nil
	default:
		return fmt.Errorf("unsupported output format: %q", format)
	}
}

func renderJSON(w io.Writer, data any) error {
	b, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", b)
	return err
}

func renderYAML(w io.Writer, data any) error {
	b, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal yaml: %w", err)
	}
	_, err = w.Write(b)
	return err
}

func renderTable(w io.Writer, data any) error {
	rows, headers := extractRows(data)
	if len(rows) == 0 {
		return nil
	}
	table := tablewriter.NewWriter(w)
	table.SetHeader(headers)
	table.AppendBulk(rows)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCaption(true, " ")
	table.SetCenterSeparator("|")
	table.Render()
	return nil
}

func renderTSV(w io.Writer, data any) error {
	rows, _ := extractRows(data)
	for _, row := range rows {
		if _, err := fmt.Fprintln(w, strings.Join(row, "\t")); err != nil {
			return err
		}
	}
	return nil
}

// extractRows converts a struct or slice of structs into string rows and headers.
// ID is always the first column. Fields tagged with jsonapi "relation,..." are skipped.
func extractRows(data any) ([][]string, []string) {
	v := reflect.ValueOf(data)
	headers := []string{"ID"}
	var rows [][]string

	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i).Interface()
			d := reflect.Indirect(reflect.ValueOf(elem))
			row := []string{d.FieldByName("ID").String()}
			row, headers = appendFields(d, row, headers, i == 0)
			rows = append(rows, row)
		}
	} else {
		d := reflect.Indirect(v)
		row := []string{d.FieldByName("ID").String()}
		row, headers = appendFields(d, row, headers, true)
		rows = append(rows, row)
	}
	return rows, headers
}

func appendFields(d reflect.Value, row []string, headers []string, buildHeaders bool) ([]string, []string) {
	for j := 0; j < d.NumField(); j++ {
		field := d.Type().Field(j)
		if field.Name == "ID" || !field.IsExported() {
			continue
		}
		tag := field.Tag.Get("jsonapi")
		if strings.HasPrefix(tag, "relation,") {
			continue
		}
		if buildHeaders {
			headers = append(headers, field.Name)
		}
		row = append(row, formatFieldValue(d.Field(j)))
	}
	return row, headers
}

func formatFieldValue(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Bool:
		return fmt.Sprintf("%t", v.Bool())
	case reflect.Pointer:
		if v.IsNil() {
			return ""
		}
		elem := v.Elem()
		switch elem.Kind() {
		case reflect.String:
			return elem.String()
		case reflect.Bool:
			return fmt.Sprintf("%t", elem.Bool())
		default:
			return fmt.Sprintf("%v", elem.Interface())
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	default:
		return v.String()
	}
}
