// AUTOGENERATED FILE: easyjson marshaller/unmarshallers.

package models

import (
	json "encoding/json"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ = json.RawMessage{}
	_ = jlexer.Lexer{}
	_ = jwriter.Writer{}
)

func easyjsonA80d3b19DecodeGithubComTopfreegamesDonationsModels(in *jlexer.Lexer, out *Item) {
	if in.IsNull() {
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "item":
			out.Key = string(in.String())
		case "metadata":
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('{')
				if !in.IsDelim('}') {
					out.Metadata = make(map[string]interface{})
				} else {
					out.Metadata = nil
				}
				for !in.IsDelim('}') {
					key := string(in.String())
					in.WantColon()
					var v1 interface{}
					v1 = in.Interface()
					(out.Metadata)[key] = v1
					in.WantComma()
				}
				in.Delim('}')
			}
		case "limitOfItemsInEachDonationRequest":
			out.LimitOfItemsInEachDonationRequest = int(in.Int())
		case "limitOfItemsPerPlayerDonation":
			out.LimitOfItemsPerPlayerDonation = int(in.Int())
		case "weightPerDonation":
			out.WeightPerDonation = int(in.Int())
		case "updatedAt":
			out.UpdatedAt = int64(in.Int64())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
}
func easyjsonA80d3b19EncodeGithubComTopfreegamesDonationsModels(out *jwriter.Writer, in Item) {
	out.RawByte('{')
	first := true
	_ = first
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"item\":")
	out.String(string(in.Key))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"metadata\":")
	if in.Metadata == nil {
		out.RawString(`null`)
	} else {
		out.RawByte('{')
		v2First := true
		for v2Name, v2Value := range in.Metadata {
			if !v2First {
				out.RawByte(',')
			}
			v2First = false
			out.String(string(v2Name))
			out.RawByte(':')
			if m, ok := v2Value.(json.Marshaler); ok {
				out.Raw(m.MarshalJSON())
			} else {
				out.Raw(json.Marshal(v2Value))
			}
		}
		out.RawByte('}')
	}
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"limitOfItemsInEachDonationRequest\":")
	out.Int(int(in.LimitOfItemsInEachDonationRequest))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"limitOfItemsPerPlayerDonation\":")
	out.Int(int(in.LimitOfItemsPerPlayerDonation))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"weightPerDonation\":")
	out.Int(int(in.WeightPerDonation))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"updatedAt\":")
	out.Int64(int64(in.UpdatedAt))
	out.RawByte('}')
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Item) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonA80d3b19EncodeGithubComTopfreegamesDonationsModels(w, v)
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Item) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonA80d3b19DecodeGithubComTopfreegamesDonationsModels(l, v)
}
