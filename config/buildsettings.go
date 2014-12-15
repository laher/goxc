package config

import (
	"log"

	"github.com/laher/goxc/typeutils"
)

type BuildSettings struct {
	//GoRoot string `json:"-"` //Made *not* settable in settings file. Only at runtime.
	//GoVersion string `json:",omitempty"` //hmm. Should I encourage this?
	Processors    *int                    `json:",omitempty"`
	Race          *bool                   `json:",omitempty"`
	Verbose       *bool                   `json:",omitempty"`
	PrintCommands *bool                   `json:",omitempty"`
	CcFlags       *string                 `json:",omitempty"`
	Compiler      *string                 `json:",omitempty"`
	GccGoFlags    *string                 `json:",omitempty"`
	GcFlags       *string                 `json:",omitempty"`
	InstallSuffix *string                 `json:",omitempty"`
	LdFlags       *string                 `json:",omitempty"`
	LdFlagsXVars  *map[string]interface{} `json:",omitempty"`
	Tags          *string                 `json:",omitempty"`
	ExtraArgs     []string                `json:",omitempty"`
}

func (this BuildSettings) Equals(that BuildSettings) bool {
	return this.Processors == that.Processors &&
		this.Race == that.Race &&
		this.Verbose == that.Verbose &&
		this.PrintCommands == that.PrintCommands &&
		this.CcFlags == that.CcFlags &&
		this.Compiler == that.Compiler &&
		this.GccGoFlags == that.GccGoFlags &&
		this.GcFlags == that.GcFlags &&
		this.InstallSuffix == that.InstallSuffix &&
		this.LdFlags == that.LdFlags &&
		this.Tags == that.Tags &&
		typeutils.StringSliceEquals(this.ExtraArgs, that.ExtraArgs) &&
		((this.LdFlagsXVars == nil && that.LdFlagsXVars == nil) ||
			(this.LdFlagsXVars != nil && that.LdFlagsXVars != nil && typeutils.AreMapsEqual(*this.LdFlagsXVars, *that.LdFlagsXVars)))
}
func (this BuildSettings) IsEmpty() bool {
	bs := BuildSettings{}
	if bs.Equals(this) {
		return true
	}
	//defaults can also be considered 'empty'
	FillBuildSettingsDefaults(&bs)
	return bs.Equals(this)
}

func buildSettingsFromMap(m map[string]interface{}) (*BuildSettings, error) {
	var err error
	bs := BuildSettings{}
	FillBuildSettingsDefaults(&bs)
	for k, v := range m {
		switch k {
		//case "GoRoot":
		//	bs.GoRoot, err = typeutils.ToString(v, k)
		case "Processors":
			var fp float64
			fp, err = typeutils.ToFloat64(v, k)
			if err == nil {
				processors := int(fp)
				bs.Processors = &processors
			}
		case "Race":
			var race bool
			race, err = typeutils.ToBool(v, k)
			if err == nil {
				bs.Race = &race
			}
		case "Verbose":
			var verbose bool
			verbose, err = typeutils.ToBool(v, k)
			if err == nil {
				bs.Verbose = &verbose
			}
		case "PrintCommands":
			var printCommands bool
			printCommands, err = typeutils.ToBool(v, k)
			if err == nil {
				bs.PrintCommands = &printCommands
			}
		case "CcFlags":
			var ccFlags string
			ccFlags, err = typeutils.ToString(v, k)
			if err == nil {
				bs.CcFlags = &ccFlags
			}
		case "Compiler":
			var s string
			s, err = typeutils.ToString(v, k)
			if err == nil {
				bs.Compiler = &s
			}
		case "GccGoFlags":
			var s string
			s, err = typeutils.ToString(v, k)
			if err == nil {
				bs.GccGoFlags = &s
			}
		case "GcFlags":
			var s string
			s, err = typeutils.ToString(v, k)
			if err == nil {
				bs.GcFlags = &s
			}
		case "InstallSuffix":
			var s string
			s, err = typeutils.ToString(v, k)
			if err == nil {
				bs.InstallSuffix = &s
			}
		case "LdFlags":
			var s string
			s, err = typeutils.ToString(v, k)
			if err == nil {
				bs.LdFlags = &s
			}
		case "Tags":
			var s string
			s, err = typeutils.ToString(v, k)
			if err == nil {
				bs.Tags = &s
			}
		case "LdFlagsXVars":
			var xVars map[string]interface{}
			xVars, err = typeutils.ToMap(v, k)
			if err == nil {
				bs.LdFlagsXVars = &xVars
			}
		case "ExtraArgs":
			bs.ExtraArgs, err = typeutils.ToStringSlice(v, k)
		default:
			log.Printf("Warning!! Unrecognised Setting '%s' (value %v)", k, v)
		}
		if err != nil {
			return &bs, err
		}

	}
	return &bs, err
}
