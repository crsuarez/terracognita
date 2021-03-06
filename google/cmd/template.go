package main

import (
	"io"
	"text/template"

	"github.com/pkg/errors"
)

const (
	// packageTmpl it's the package definition
	packageTmpl = `
	package google
	// Code generated by 'go generate'; DO NOT EDIT
	import (
		"context"

		"github.com/pkg/errors"

		"google.golang.org/api/compute/v1"
		"google.golang.org/api/sqladmin/v1beta4"
		"google.golang.org/api/storage/v1"
	)
	`

	// functionTmpl it's the implementation of a reader function
	functionTmpl = `
	// List{{ .Name }} returns a list of {{ .Name }} within a project {{ if .Zone }}and a zone {{ end }}
	func (r *GCPReader) List{{ .Name}}(ctx context.Context{{ if not .NoFilter }}, filter string {{ end }}) ({{ if .Zone }}map[string]{{end}}[]{{ .API }}.{{ .Resource }}, error) {
		service := {{ .API }}.New{{ .ServiceName}}Service(r.{{ .API }})
		{{ if .Zone }}
		list := make(map[string][]{{ .API }}.{{ .Resource }})
		zones, err := r.getZones()
		if err != nil {
			return nil, errors.Wrap(err, "unable to get zones in region")
		}
		for _, zone := range zones {
		{{ end }}
		resources := make([]{{ .API }}.{{ .Resource }}, 0)
		{{ if .Zone }}
		if err := service.List(r.project, zone).
		{{ else if .Region }}
		if err := service.List(r.project, r.region).
		{{ else }}
		if err := service.List(r.project).
		{{ end }}
		{{ if not .NoFilter }}
			Filter(filter).
		{{ end }}
			MaxResults(int64(r.maxResults)).
			Pages(ctx, func(list *{{ .API }}.{{ .ResourceList }}) error {
				for _, res := range list.Items {
					resources = append(resources, *res)
				}
				return nil
			}); err != nil {
			return nil, errors.Wrap(err, "unable to list {{ .API }} {{ .Resource }} from google APIs")
		}
		{{ if .Zone }}
		list[zone] = resources
		}
		return list, nil
		{{ else }}
		return resources, nil
		{{ end }}
	}
	`
)

var (
	fnTmpl  *template.Template
	pkgTmpl *template.Template
)

func init() {
	var err error
	fnTmpl, err = template.New("template").Parse(functionTmpl)
	if err != nil {
		panic(err)
	}
	pkgTmpl, err = template.New("template").Parse(packageTmpl)
	if err != nil {
		panic(err)
	}
}

// Function is the definition of one of the functions
type Function struct {
	// Resource is the Google name of the entity, like
	// Firewall, Instance, etc.
	// https://godoc.org/google.golang.org/api/compute/v1
	Resource string

	// Zone is used to determine whether the resource is located within google zones or not
	Zone bool

	// Name is the function name to be generated
	// it can be useful if you `Resource` is `SslCertificate`, which is not `go`
	// compliant, `Name` will be `SSLCertificate`, your Function name will be
	// `ListSSLCertificates`
	Name string

	// ServiceName is name of the Google SDK service name
	// If your service is `TargetHttpProxy`, your service name will
	// be `TargetHttpProxies`
	ServiceName string

	// Region is used to determine whether the resource is dedicated to a region or not
	Region bool

	// API is used to determine the
	// google API to use as defined in the Reader
	// for a complete list of API: https://godoc.org/google.golang.org/api
	// ex: compute, storage
	// default goes to `compute`
	API string

	// NoFilter is used to determine if the
	// resource is based on filters or not
	// default goes to `false`
	NoFilter bool

	// ResourceList overrides the default name of
	// the resources list: `resourceList`
	// exemple:
	// for the components Instance, the list struct is `InstanceList`
	// but for the components Bucket, the list struct is `Buckets`
	ResourceList string
}

// Execute uses the fnTmpl to interpolate f
// and write the result to w
func (f Function) Execute(w io.Writer) error {
	if len(f.ResourceList) == 0 {
		f.ResourceList = f.Resource + "List"
	}
	if len(f.API) == 0 {
		f.API = "compute"
	}
	if len(f.Name) == 0 {
		f.Name = f.Resource + "s"
	}
	if len(f.ServiceName) == 0 {
		f.ServiceName = f.Resource + "s"
	}
	if err := fnTmpl.Execute(w, f); err != nil {
		return errors.Wrapf(err, "failed to Execute with Function %+v", f)
	}
	return nil
}
