package {{.Package}}


import (
	"context"

	"encoding/json"

	mrpc "github.com/yansss/mrpc/runtime"
)

{{range $serviceName,$service := .Services}}

type {{$serviceName}} interface {

{{range $methodName,$method := $service.Methods}}

{{if $method.ServerStream}}
	{{$methodName}}(ctx context.Context, req *{{$method.Request}}, stream *mrpc.StreamServer) error
{{else}}
	{{$methodName}}(ctx context.Context, req *{{$method.Request}}) (*{{$method.Response}}, error)

{{end}}

{{end}}
}

func Register{{$serviceName}}(registry *mrpc.Registry, service {{$serviceName}}) error {

{{range $methodName,$method := $service.Methods}}

{{if $method.ServerStream}}
	registry.Register("{{$serviceName}}", "{{$methodName}}",
		mrpc.MethodHandler{
		StreamHandler:	func(ctx context.Context, data json.RawMessage, stream *mrpc.StreamServer) error {
				req := &{{$method.Request}}{}
				err := json.Unmarshal(data, req)
				if err != nil {
					return err
				}
				return service.{{$methodName}}(ctx, req, stream)
			},
		},
	)
{{else}}
	registry.Register("{{$serviceName}}", "{{$methodName}}",
		mrpc.MethodHandler{
		UnaryHandler:	func(ctx context.Context, data json.RawMessage) (any, error,) {
				req := &{{$method.Request}}{}
				err := json.Unmarshal(data, req)
				if err != nil {
					return nil,err
				}
				return service.{{$methodName}}(ctx, req)
			},
		},
	)
{{end}}

{{end}}
	return nil
}



{{end}}