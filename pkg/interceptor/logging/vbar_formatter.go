package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// VBarFormatter formats logs into vertical bar separated style.
type VBarFormatter struct {
	// TimestampFormat sets the format used for marshaling timestamps.
	// The format to use is the same than for time.Format or time.Parse from the standard
	// library.
	// The standard Library already provides a set of predefined format.
	TimestampFormat string

	// DisableHTMLEscape allows disabling html escaping in output
	DisableHTMLEscape bool

	// DataKey allows users to put all the log entry parameters into a nested dictionary at a given key.
	DataKey string

	// FieldMap allows users to customize the names of keys for default fields.
	// As an example:
	// formatter := &JSONFormatter{
	//   	FieldMap: FieldMap{
	// 		 FieldKeyTime:  "@timestamp",
	// 		 FieldKeyLevel: "@level",
	// 		 FieldKeyMsg:   "@message",
	// 		 FieldKeyFunc:  "@caller",
	//    },
	// }
	FieldMap FieldMap

	// CallerPrettyfier can be set by the user to modify the content
	// of the function and file keys in the json data when ReportCaller is
	// activated. If any of the returned value is the empty string the
	// corresponding key will be removed from json fields.
	CallerPrettyfier func(*runtime.Frame) (function string, file string)

	// PrettyPrint will indent all json logs
	PrettyPrint bool
}

// Format renders a single log entry.
func (f *VBarFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields, len(entry.Data)+4)
	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/sirupsen/logrus/issues/137
			data[k] = v.Error()
		case proto.Message:
			// Use protojson to marshal proto message
			raw, err := protojson.Marshal(v)
			if err != nil {
				data[logrus.FieldKeyLogrusError] = err.Error()
			} else {
				data[k] = json.RawMessage(raw)
			}
		case string:
			if _, ok := rawJSONKeys[k]; ok {
				data[k] = json.RawMessage(v)
			} else {
				data[k] = v
			}
		default:
			data[k] = v
		}
	}

	if f.DataKey != "" {
		newData := make(logrus.Fields, 4)
		newData[f.DataKey] = data
		data = newData
	}

	prefixFieldClashes(data, f.FieldMap, entry.HasCaller())

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = time.RFC3339Nano
	}

	if entry.HasCaller() {
		funcVal := entry.Caller.Function
		fileVal := fmt.Sprintf("%s:%d", entry.Caller.File, entry.Caller.Line)
		if f.CallerPrettyfier != nil {
			funcVal, fileVal = f.CallerPrettyfier(entry.Caller)
		}
		if funcVal != "" {
			data[f.FieldMap.resolve(logrus.FieldKeyFunc)] = funcVal
		}
		if fileVal != "" {
			data[f.FieldMap.resolve(logrus.FieldKeyFile)] = fileVal
		}
	}

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	encoder := json.NewEncoder(b)
	encoder.SetEscapeHTML(!f.DisableHTMLEscape)
	if f.PrettyPrint {
		encoder.SetIndent("", "  ")
	}
	b.WriteString("[")
	// Print header
	// - level
	b.WriteString(strings.ToUpper(entry.Level.String()))
	// - topic
	b.WriteString(fmt.Sprintf("|%s.%s.%s.%s", data[ServiceFieldKey], data[ComponentFieldKey], data[MethodFieldKey], data[MethodTypeFieldKey]))
	// - timestamp
	b.WriteString(fmt.Sprintf("|%s %s", entry.Time.Format(timestampFormat), entry.Message))
	// Other fields
	for k, v := range data {
		b.WriteString(fmt.Sprintf("|%s=", k))
		if err := encoder.Encode(v); err != nil {
			return nil, fmt.Errorf("failed to marshal fields to JSON, %w", err)
		}
		// Discard new line by encoder
		b.Truncate(b.Len() - 1)
	}
	b.WriteString("]\n")

	return b.Bytes(), nil
}
