package pvd

import (
	"os"
	"text/template"
)

type ReturnStatus struct {
	Code    int
	Message string
}

func (rs ReturnStatus) Exit() {
	tmpl, err := template.New("exit_tmpl").Parse("Return code:\t{{.Code}}\nMessage    :\t{{.Message}}\n")
	if err != nil {
		os.Exit(rs.Code)
	} else {
		if rs.Code == 0 {
			err = tmpl.Execute(os.Stdout, rs)
			if err != nil {
				os.Exit(rs.Code)
			}
		} else {
			err = tmpl.Execute(os.Stderr, rs)
			if err != nil {
				os.Exit(rs.Code)
			}
		}
	}
}
