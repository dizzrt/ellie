{{ range .Errors }}

{{ if .HasComment }}{{ .Comment }}{{ end -}}
func Is{{.CamelValue}}(err error) bool {
	if err == nil {
		return false
	}
	e := errors.FromError(err)
	return e.Reason == {{ .Name }}_{{ .Value }}.String() && e.Code == {{ .Code }}
}

{{ if .HasComment }}{{ .Comment }}{{ end -}}
func {{ .CamelValue }}() *errors.Error {
	 return errors.New({{ .Code }}, {{ .Name }}_{{ .Value }}.String(), "")
}

func {{ .CamelValue }}WithMsg(format string, args ...interface{}) *errors.Error {
	 return errors.New({{ .Code }}, {{ .Name }}_{{ .Value }}.String(), fmt.Sprintf(format, args...))
}

{{- end }}
