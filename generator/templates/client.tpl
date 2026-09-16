package {{.Package}}

import (
	"context"
	mrpc "mrpc/runtime"
)

{{range $serviceName, $service := .Services}}

type {{$serviceName}}Client struct {
	client *mrpc.Client
}

func New{{$serviceName}}Client (client *mrpc.Client) (*{{$serviceName}}Client) {
	return &{{$serviceName}}Client{
		client: client,
	}
}

{{range $methodName, $method := $service.Methods}}
{{if $method.ServerStream}}

	type {{$serviceName}}{{$methodName}}Client struct {
		stream *mrpc.ClientStream
	}

	func (c *{{$serviceName}}Client) {{$methodName}}(ctx context.Context, req *{{$method.Request}}) (*{{$serviceName}}{{$methodName}}Client, error) {
		stream, err := c.client.NewClientStream("{{$serviceName}}", "{{$methodName}}", req)
		if err != nil {
			return nil, err
		}	
		return &{{$serviceName}}{{$methodName}}Client{
			stream: stream,
		}, nil
	}

	func (s *{{$serviceName}}{{$methodName}}Client) Recv() (*{{$method.Response}},error) {
		var resp {{$method.Response}}
		err := s.stream.Recv(&resp) 
		if err != nil {
			return nil, err
		}
		return &resp, nil
	}
{{else}}
	func (c *{{$serviceName}}Client) {{$methodName}}(ctx context.Context, req *{{$method.Request}})(*{{$method.Response}},error){

		var resp {{$method.Response}}
		err := c.client.CallUnary(ctx, "{{$serviceName}}", "{{$methodName}}", req, &resp)
		if err != nil {
			return nil, err
		}
		return &resp, nil
	}
{{end}}
{{end}}
{{end}}