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
	input := schema.InputObject("{{.InputObjName}}", {{.Type}}{})
	{{$name:=.Name}}
	{{range .Maps}}
		input.FieldFunc("{{.FieldName}}", func(target *{{$name}}, source *schemabuilder.Map) error {
			v := source.Value
	
			decodedValue, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				return err
			}
	
			data := make(map[{{.Key}}]{{.Value}})
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
	{{range .Durations}}
	input.FieldFunc("{{.FieldName}}", func(target *{{$name}}, source []*schemabuilder.Duration) {
		array := make([]*duration.Duration, 0 ,len(source))
		for _, s:= range source{
			array = append(array, (*duration.Duration)(s))
		}
		target.{{.Name}} = array
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
	payload := schema.Object("{{.PayloadObjName}}", {{.Name}}{}){{$name:=.Name}}
	{{range .Maps}}
		payload.FieldFunc("{{.FieldName}}", func(ctx context.Context, in *{{$name}}) (*schemabuilder.Map, error) {
			data, err := json.Marshal({{.TargetVal}})
			if err != nil {
				return nil, err
			}
	
			return &schemabuilder.Map{Value:string(data)}, nil
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
	{{range .Durations}}
	payload.FieldFunc("{{.FieldName}}", func(ctx context.Context, in *{{$name}}) []*schemabuilder.Duration {
		array := make([]*schemabuilder.Duration, 0, len(in.{{.Name}}))
		for _, d := range in.{{.Name}}{
			array = append(array, (*schemabuilder.Duration)(d))
		}
		return array
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
			{{range .MapsData}}
			v{{.Name}} := args.{{.Name}}.Value
			decodedValue{{.Name}}, err{{.Name}} := base64.StdEncoding.DecodeString(v{{.Name}})
			if err{{.Name}} != nil {
				return nil,err{{.Name}}
			}
			{{.NewVarName}}Map := make(map[{{.Key}}]{{.Value}})
			if err{{.Name}} := json.Unmarshal(decodedValue{{.Name}}, &{{.NewVarName}}Map); err{{.Name}} != nil {
				return nil,err{{.Name}}
			}{{end}}
			request := &{{.InputName}}{
			{{range .ReturnType}}
			{{.Name}}: {{.Type}},{{end}}
			}
			{{range .Durations}}
			array{{.Name}} := make([]*duration.Duration, 0 ,len(args.{{.Name}}))
			for _, s:= range args.{{.Name}}{
				array{{.Name}} = append(array{{.Name}}, (*duration.Duration)(s))
			}
			request.{{.Name}}=array{{.Name}}
			{{end}}
			{{range .Oneofs}}
			{{$oneOfNameQ:= .Name}}
				{{range .Fields}}
					if args.{{.Name}} != nil{
						request.{{$oneOfNameQ}} = args.{{.Name}}
					}
				{{end}}
			{{end}}
			return client{{"."}}{{.ReturnFunc}}(ctx, request)
		})
	{{end}}
	{{range .Mutations}}
		schema.Mutation().FieldFunc("{{.FieldName}}", func(ctx context.Context, args struct {
			Input {{.InputType}}
		}) ({{.FirstReturnArgType}}, error) {
			request := {{.RequestType}}{
				{{range .RequestFields}}
				{{.}}: args{{"."}}Input{{"."}}{{.}},{{end}}
			}
			{{range .OneOfs}}
			{{$oneOfName:= .Name}}
				{{range .Fields}}
				if args.Input.{{.Name}} != nil{
					request.{{$oneOfName}} = args.Input.{{.Name}}
				}{{end}}{{end}}
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
		target{{"."}}{{.TargetName}} ={{.TargetVal}}
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
		return {{.FieldFuncReturn}}
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
	{{range .Maps}}
		input.FieldFunc("{{.FieldName}}", func(target *{{$name}}Input, source *schemabuilder.Map) error {
			v := source.Value
	
			decodedValue, err := base64.StdEncoding.DecodeString(v)
			if err != nil {
				return err
			}
	
			data := make(map[{{.Key}}]{{.Value}})
			if err := json.Unmarshal(decodedValue, &data); err != nil {
				return err
			}
	
			target.{{.TargetName}} = data
			return nil
		}){{end}}
	{{range .Fields}}
		input.FieldFunc("{{.FieldName}}", func(target *{{$name}}Input, source {{.FuncPara}}) {
			target{{"."}}{{.TargetName}} = {{.TargetVal}}
		})
	{{end}}
	{{range .Durations}}
	input.FieldFunc("{{.Name}}", func(target *{{$name}}Input, source []*schemabuilder.Duration) {
		array := make([]*duration.Duration, 0 ,len(source))
		for _, s:= range source{
			array = append(array, (*duration.Duration)(s))
		}
		target.{{.Name}} = array
	}){{end}}
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
