package main

import "text/template"
import "log"

func getEnumTemplate() *template.Template {

	tmpl := `
func Register{{.Name}}(schema *schemabuilder.Schema){
{{$name:=.Name}}
	schema.Enum({{.Name}}(0), map[string]interface{}{
		{{range .Values}}	"{{.Value}}": {{$name}}({{.Index}}),{{"\n"}}{{end}}
	})
}
`

	t, err := template.New("enum").Parse(tmpl)
	if err != nil {
		log.Fatal("Parse: ", err)
		panic(err)
	}

	return t
}

func getUnionStructTemplate() *template.Template {

	tmpl := `
{{range .}}
type {{.UnionName}} struct {
	schemabuilder.Union
	{{range .UnionFields}}
		{{.}}{{end}}
}
{{end}}
`

	t, err := template.New("UnionStruct").Parse(tmpl)
	if err != nil {
		log.Fatal("Parse: ", err)
		panic(err)
	}

	return t
}

func getInputTemplate() *template.Template {

	tmpl := `
func RegisterInput{{.Name}}(schema *schemabuilder.Schema) {
	input := schema.InputObject("{{.InputObjName}}", {{.Name}}{})
	{{$name:=.Name}}
	{{range .Maps}}
		input.FieldFunc("{{.FieldName}}", func(target *{{$name}}, source *schemabuilder.Map) error {
			v := string({{.TargetVal}})
	
			decodedValue, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				return err
			}
	
			data := make(map[{{.Key}}]*{{.Value}})
			if err := json.Unmarshal(decodedValue, &data); err != nil {
				return err
			}
	
			target.{{.TargetName}} = data
			return nil
		}){{end}}
	{{range .Fields}}
	input.FieldFunc("{{.FieldName}}", func(target *{{$name}}, source {{.FuncPara}}) {
		target.{{.TargetName}} = {{.TargetVal}}
	}){{end}}
}
`

	t, err := template.New("InputType").Parse(tmpl)
	if err != nil {
		log.Fatal("Parse: ", err)
		panic(err)
	}

	return t
}
func getPayloadTemplate() *template.Template {

	tmpl := `
func RegisterPayload{{.Name}}(schema *schemabuilder.Schema) {
	payload := schema.Object("{{.Name}}", {{.Name}}{}){{$name:=.Name}}
	{{range .Maps}}
		payload.FieldFunc("{{.FieldName}}", func(ctx context.Context, in *{{$name}}) (*schemabuilder.Map, error) {
			data, err := json.Marshal({{.TargetVal}})
			if err != nil {
				return nil, err
			}
	
			encodedValue := base64.StdEncoding.EncodeToString(data)
			return (*schemabuilder.Map)(&encodedValue), nil
		}){{end}}
	{{range .UnionObjects}}
	payload.FieldFunc("{{.FieldName}}", func(ctx context.Context, in *{{$name}}) {{.FuncReturn}} {
		switch v := in{{"."}}{{.SwitchName}}{{"."}}(type) {
		{{range .Fields}}
		case *{{.CaseName}}:
			return {{.ReturnType}}{
				{{.CaseName}}: v,
			}
		{{end}}
		}
		return nil
	})
	{{end}}
	{{range .Fields}}
	payload.FieldFunc("{{.FieldName}}", func(ctx context.Context, in *{{$name}}) {{.FuncPara}} {
		return {{.TargetVal}}
	}){{end}}
}
`

	t, err := template.New("Payload").Parse(tmpl)
	if err != nil {
		log.Fatal("Parse: ", err)
		panic(err)
	}

	return t
}

func getServiceTemplate() *template.Template {

	tmpl := `
func Register{{.Name}}Operations(schema *schemabuilder.Schema, client {{.Name}}Client) {
	{{range .Queries}}
		schema.Query().FieldFunc("{{.FieldName}}", func(ctx context.Context, args struct {
		{{range .InType}}
		{{.Name}} {{.Type}}{{end}}
		}) ({{.FirstReturnArgType}}, error) {
			return client{{"."}}{{.ReturnFunc}}(ctx, &{{.InputName}}{
			{{range .ReturnType}}
			{{.Name}}: {{.Type}},{{end}}
			})
		})
	{{end}}
	{{range .Mutations}}
		schema.Mutation().FieldFunc("{{.FieldName}}", func(ctx context.Context, args struct {
			Input {{.InputType}}
		}) ({{.FirstReturnArgType}}, error) {
			request := {{.RequestType}}{
				{{range .RequestFields}}
				{{.}}: args{{"."}}Input{{"."}}{{.}},
				{{end}}
			}
			response, err := client{{"."}}{{.ResponseType}}(ctx, request)
			return {{.ReturnType}}{
				Payload:          response,
				ClientMutationId: args.Input.ClientMutationId,
			}, err
		})
	{{end}}
}
`

	t, err := template.New("service").Parse(tmpl)
	if err != nil {
		log.Fatal("Parse: ", err)
		panic(err)
	}

	return t
}

func getServiceStructInputTemplate() *template.Template {

	tmpl := `
{{range .}}
type {{.RpcName}}Input struct {
	{{range .InputFields}}
		{{.Name}}  {{.Type}}{{end}}
	ClientMutationId string
}
{{end}}
`

	t, err := template.New("serviceStruct").Parse(tmpl)
	if err != nil {
		log.Fatal("Parse: ", err)
		panic(err)
	}

	return t
}

func getServiceStructPayloadTemplate() *template.Template {

	tmpl := `
{{range .}}
type {{.Name}}Payload struct {
	Payload          {{.ReturnType}}
	ClientMutationId string
}
{{end}}
`

	t, err := template.New("ServiceStructPayload").Parse(tmpl)
	if err != nil {
		log.Fatal("Parse: ", err)
		panic(err)
	}

	return t
}

func getInitTemplate() *template.Template {

	tmpl := `
func init() {
{{range $key, $value := . }}
	{{$key}}(gtypes.Schema){{end}}
}
`

	t, err := template.New("Init").Parse(tmpl)
	if err != nil {
		log.Fatal("Parse: ", err)
		panic(err)
	}

	return t
}

func getOneofInputTemplate() *template.Template {

	tmpl := `
{{range .}}
func RegisterInput{{.Name}}(schema *schemabuilder.Schema) {
	input := schema.InputObject("{{.SchemaObjectPara}}", {{.Name}}{})
	input.FieldFunc("{{.FieldFuncPara}}", func(target *{{.Name}}, source *{{.FieldFuncSecondParaFuncPara}}) {
		target{{"."}}{{.TargetName}} = *source
	})
}
{{end}}
`

	t, err := template.New("oneof").Parse(tmpl)
	if err != nil {
		log.Fatal("Parse: ", err)
		panic(err)
	}

	return t
}

func getOneofPayloadTemplate() *template.Template {

	tmpl := `
{{range .}}
func RegisterPayload{{.Name}}(schema *schemabuilder.Schema) {
	payload := schema.Object("{{.Name}}", {{.Name}}{})
	payload.FieldFunc("{{.FieldFuncPara}}", func(ctx context.Context, in *{{.Name}}) {{.FieldFuncSecondFuncReturn}} {
		return in{{"."}}{{.FieldFuncReturn}}
	})
}
{{end}}
`

	t, err := template.New("oneofPayload").Parse(tmpl)
	if err != nil {
		log.Fatal("Parse: ", err)
		panic(err)
	}

	return t
}

func getServiceStructInputFuncTemplate() *template.Template {

	tmpl := `
{{range .}}
func RegisterInput{{.Name}}Input(schema *schemabuilder.Schema) {
	input := schema.InputObject("{{.Name}}Input", {{.Name}}Input{}) {{$name:=.Name}}
	{{range .Fields}}
		input.FieldFunc("{{.FieldName}}", func(target *{{$name}}Input, source {{.FuncPara}}) {
			target{{"."}}{{.TargetName}} = {{.TargetVal}}
		})
	{{end}}
	input.FieldFunc("clientMutationId", func(target *{{.Name}}Input, source *string) {
		target.ClientMutationId = *source
	})
}
{{end}}
`

	t, err := template.New("inputMutationStruct").Parse(tmpl)
	if err != nil {
		log.Fatal("Parse: ", err)
		panic(err)
	}

	return t
}

func getServiceStructPayloadFuncTemplate() *template.Template {

	tmpl := `
{{range .}}
	func RegisterPayload{{.Name}}Payload(schema *schemabuilder.Schema) {
		payload := schema.Object("{{.Name}}Payload", {{.Name}}Payload{}){{$name:= .Name}}
		payload.FieldFunc("payload", func(ctx context.Context, in *{{.Name}}Payload) {{.ReturnType}} {
			return in.Payload
		})
		payload.FieldFunc("clientMutationId", func(ctx context.Context, in *{{.Name}}Payload) string {
			return in.ClientMutationId
		})
	}
{{end}}
`

	t, err := template.New("inputMutationStruct").Parse(tmpl)
	if err != nil {
		log.Fatal("Parse: ", err)
		panic(err)
	}

	return t
}
