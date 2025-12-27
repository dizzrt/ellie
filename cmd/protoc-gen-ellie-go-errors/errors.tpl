{{ range .Errors }}

{{ if .HasComment }}{{ .Comment }}{{ end }}
func Is{{.CamelValue}}(err error) bool {
    if err == nil {
        return false
    }
    
    se := errors.NewStandardErrorFromError(err)
    return se.Reason() == {{ .Name }}_{{ .Value }}.String() && se.Code() == {{ .Code }}
}

{{ if .HasComment }}{{ .Comment }}{{ end }}
func {{ .CamelValue }}() *errors.StandardError {
    var status *codes.Code = nil
    {{ if ge .Status 0 }}status = errors.StatusPtrFromInt({{ .Status }}) // {{ .StatusName }}{{ end }}

    return errors.NewStandardError(status, {{ .Code }}, {{ .Name }}_{{ .Value }}.String(), "")
}

{{ end }}
