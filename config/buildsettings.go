package config


import (
	"github.com/laher/goxc/typeutils"
	"log"
)

type BuildSettings struct {
	GoRoot string `json:",omitempty"`
	Processors int `json:"omitempty"`
	Race bool `json:",omitempty"`
	Verbose bool `json:",omitempty"`
	PrintCommands bool `json:",omitempty"`
	CcFlags string `json:",omitempty"`
	Compiler string `json:",omitempty"`
	GccGoFlags string `json:",omitempty"`
	GcFlags string `json:",omitempty"`
	InstallSuffix string `json:",omitempty"`
	LdFlags string `json:",omitempty"`
	LdFlagsXVars *map[string]interface{} `json:",omitempty"`
	Tags string `json:",omitempty"`
	ExtraArgs []string `json:",omitempty"`
}

func BuildSettingsDefault() BuildSettings {
	bs := BuildSettings{}
	bs.LdFlagsXVars = &map[string]interface{}{"TimeNow" : "main.BUILD_DATE", "Version" : "main.VERSION" }
	return bs
}

func buildSettingsFromMap(m map[string]interface{}) (*BuildSettings, error) {
	var err error
	bs := BuildSettingsDefault()
	for k, v := range m {
		switch k {
		case "GoRoot":
			bs.GoRoot, err = typeutils.ToString(v, k)
		case "Processors":
			bs.Processors, err = typeutils.ToInt(v, k)
		case "Race":
			bs.Race, err = typeutils.ToBool(v, k)
		case "Verbose":
			bs.Verbose, err = typeutils.ToBool(v, k)
		case "PrintCommands":
			bs.PrintCommands, err = typeutils.ToBool(v, k)
		case "CcFlags":
			bs.CcFlags, err = typeutils.ToString(v, k)
		case "Compiler":
			bs.Compiler, err = typeutils.ToString(v, k)
		case "GccGoFlags":
			bs.GccGoFlags, err = typeutils.ToString(v, k)
		case "GcFlags":
			bs.GcFlags, err = typeutils.ToString(v, k)
		case "InstallSuffix":
			bs.InstallSuffix, err = typeutils.ToString(v, k)
		case "LdFlags":
			bs.LdFlags, err = typeutils.ToString(v, k)
		case "Tags":
			bs.Tags, err = typeutils.ToString(v, k)
		case "LdFlagsXVars":
			var xVars map[string]interface{}
			xVars, err = typeutils.ToMap(v, k)
			bs.LdFlagsXVars = &xVars
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
