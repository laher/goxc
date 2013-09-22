---
layout: default
title: Downloads
---

NOTE: don't use binaries for goxc. Please use `go get -u github.com/laher/goxc` instead.

{{.AppName}} downloads (version {{.Version}})

{{range $k, $v := .Categories}}### {{$k}}

{{range $v}} * [{{.Text}}]({{.RelativeLink}})
{{end}}
{{end}}

---
{{.ExtraVars.footer}}
