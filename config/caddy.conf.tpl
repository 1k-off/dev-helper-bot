{{ .domain }} {
        {{if eq .basicauth "Restricted"}}
        basicauth /* {
                demo $2a$12$Y.lglYtJKk89gqdK0pWiPurj5pzsmUccJgHlOMLcLJ5IMN4DZcrHG # demo demo
        }
        {{end}}
        reverse_proxy {
                to {{ .scheme }}://{{ .ip }}:{{ .port }}
                {{if eq .scheme "https"}}
                transport http {
                  tls
                  tls_insecure_skip_verify
                }
                {{end}}
        }
}