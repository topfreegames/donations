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

func easyjsonF464aa0aDecodeGithubComTopfreegamesDonationsModels(in *jlexer.Lexer, out *DonationRequest) {
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
		case "id":
			out.ID = string(in.String())
		case "item":
			out.Item = string(in.String())
		case "player":
			out.Player = string(in.String())
		case "clan":
			out.Clan = string(in.String())
		case "gameID":
			out.GameID = string(in.String())
		case "donations":
			if in.IsNull() {
				in.Skip()
				out.Donations = nil
			} else {
				in.Delim('[')
				if !in.IsDelim(']') {
					out.Donations = make([]Donation, 0, 1)
				} else {
					out.Donations = []Donation{}
				}
				for !in.IsDelim(']') {
					var v1 Donation
					(v1).UnmarshalEasyJSON(in)
					out.Donations = append(out.Donations, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "createdAt":
			out.CreatedAt = int64(in.Int64())
		case "updatedAt":
			out.UpdatedAt = int64(in.Int64())
		case "finishedAt":
			out.FinishedAt = int64(in.Int64())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
}
func easyjsonF464aa0aEncodeGithubComTopfreegamesDonationsModels(out *jwriter.Writer, in DonationRequest) {
	out.RawByte('{')
	first := true
	_ = first
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"id\":")
	out.String(string(in.ID))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"item\":")
	out.String(string(in.Item))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"player\":")
	out.String(string(in.Player))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"clan\":")
	out.String(string(in.Clan))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"gameID\":")
	out.String(string(in.GameID))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"donations\":")
	if in.Donations == nil {
		out.RawString("null")
	} else {
		out.RawByte('[')
		for v2, v3 := range in.Donations {
			if v2 > 0 {
				out.RawByte(',')
			}
			(v3).MarshalEasyJSON(out)
		}
		out.RawByte(']')
	}
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"createdAt\":")
	out.Int64(int64(in.CreatedAt))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"updatedAt\":")
	out.Int64(int64(in.UpdatedAt))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"finishedAt\":")
	out.Int64(int64(in.FinishedAt))
	out.RawByte('}')
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v DonationRequest) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonF464aa0aEncodeGithubComTopfreegamesDonationsModels(w, v)
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *DonationRequest) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonF464aa0aDecodeGithubComTopfreegamesDonationsModels(l, v)
}
func easyjsonF464aa0aDecodeGithubComTopfreegamesDonationsModels1(in *jlexer.Lexer, out *Donation) {
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
		case "id":
			out.ID = string(in.String())
		case "player":
			out.Player = string(in.String())
		case "amount":
			out.Amount = int(in.Int())
		case "weight":
			out.Weight = int(in.Int())
		case "createdAt":
			out.CreatedAt = int64(in.Int64())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
}
func easyjsonF464aa0aEncodeGithubComTopfreegamesDonationsModels1(out *jwriter.Writer, in Donation) {
	out.RawByte('{')
	first := true
	_ = first
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"id\":")
	out.String(string(in.ID))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"player\":")
	out.String(string(in.Player))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"amount\":")
	out.Int(int(in.Amount))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"weight\":")
	out.Int(int(in.Weight))
	if !first {
		out.RawByte(',')
	}
	first = false
	out.RawString("\"createdAt\":")
	out.Int64(int64(in.CreatedAt))
	out.RawByte('}')
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Donation) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonF464aa0aEncodeGithubComTopfreegamesDonationsModels1(w, v)
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Donation) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonF464aa0aDecodeGithubComTopfreegamesDonationsModels1(l, v)
}
