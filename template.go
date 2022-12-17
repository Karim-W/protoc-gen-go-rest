package main

import (
	"bytes"
	"strings"
	"text/template"
)

var httpTemplate = `
{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}

type {{.ServiceType}}UseCases interface {
	MatchError(err error) (int, any)
{{range .Methods}}
	{{.Name}}Request(ctx *gin.Context, req *{{.Request}}) (int,*{{.Reply}}, error)
}
{{end}}



type {{.ServiceType}}Impl struct {
	svc {{.ServiceType}}UseCases
}

func Create{{.ServiceType}}() rest.Controller {
	return &{{.ServiceType}}Impl{}
}

func (ctrl *{{.ServiceType}}Impl) SetupRoutes(version int,rg *gin.RouterGroup) {
{{range .Methods}}
	rg.{{.Method}}("{{.Path}}", ctrl.{{.Name}}Handler)
{{- end}}
}



{{range .Methods}}
func (ctrl *{{$svrType}}Impl) {{.Name}}Handler(ctx *gin.Context){
	var req {{.Request}}
	{{- if .HasBody}}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	{{- end}}
	{{- if not (eq .Body "")}}
	if err := rest.BindReqQuery(ctx, &req); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	{{- else}}
		if err := rest.BindReqQuery(ctx,&req{{.Body}}); err != nil {
    	ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}
  {{- end}}
  {{- if .HasVars}}
	if err := rest.BindReqVars(ctx,&req); err != nil {
    	ctx.JSON(400, gin.H{"error": err.Error()})
			return
	}
	{{- end}}
	code,resp, err := ctrl.svc.{{.Name}}Request(ctx, &req)
	if err != nil {
		ctx.JSON(ctrl.svc.MatchError(err))
		return
	}
	ctx.JSON(code, resp)
}
{{end}}


type {{.ServiceType}}HTTPClient interface {
{{- range .MethodSets}}
	{{.Name}}(ctx context.Context, req *{{.Request}}, opts *stdlib.ClientOptions) (code int,rsp *{{.Reply}}, err error) 
{{- end}}
}
	
type {{.ServiceType}}HTTPClientImpl struct{
	cc stdlib.Client
	baseURL string
}
	
func New{{.ServiceType}}HTTPClient (client stdlib.Client,baseUrl string) {{.ServiceType}}HTTPClient {
	return &{{.ServiceType}}HTTPClientImpl{
		cc: client,
		baseURL: baseUrl,
	}
}
{{range .MethodSets}}
// {{.Method}} {{.Path}}
{{- if .HasVars}}
// Requires path variables to be passed in opts.PathParams
{{- end}}
{{- if .HasBody}}
// pass {{.Request}} as body
{{- end}}
func (c *{{$svrType}}HTTPClientImpl) {{.Name}}(ctx context.Context, in *{{.Request}}, opts *stdlib.ClientOptions) (int,*{{.Reply}}, error) {
	var out {{.Reply}}
	{{if .HasBody -}}
	code,err := c.cc.Invoke(ctx, "{{.Method}}", c.baseURL+"{{.Path}}", opts ,in{{.Body}}, &out{{.ResponseBody}})
	{{else -}} 
  code,err := c.cc.Invoke(ctx, "{{.Method}}", c.baseURL+"{{.Path}}", opts ,nil, &out{{.ResponseBody}})
	{{end -}}
	if err != nil {
		return code,nil, err
	}
	return code,&out, err
}
{{end}}
`

type serviceDesc struct {
	ServiceType string // Greeter
	ServiceName string // helloworld.Greeter
	Metadata    string // api/helloworld/helloworld.proto
	Methods     []*methodDesc
	MethodSets  map[string]*methodDesc
}

type methodDesc struct {
	// method
	Name         string
	OriginalName string // The parsed original name
	Num          int
	Request      string
	Reply        string
	// http_rule
	Path          string
	AltPath       string
	Method        string
	HasVars       bool
	HasBody       bool
	HasPositional bool
	Body          string
	ResponseBody  string
}

func (s *serviceDesc) execute() string {
	s.MethodSets = make(map[string]*methodDesc)
	for _, m := range s.Methods {
		s.MethodSets[m.Name] = m
	}
	buf := new(bytes.Buffer)
	tmpl, err := template.New("http").Parse(strings.TrimSpace(httpTemplate))
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(buf, s); err != nil {
		panic(err)
	}
	return strings.Trim(buf.String(), "\r\n")
}
