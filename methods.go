package main

import (
	"bytes"
	"strings"

	"github.com/golang/protobuf/proto"
	pgs "github.com/lyft/protoc-gen-star"
	pbt "go.appointy.com/protoc-gen-jaal/schema"
)

type Value struct {
	Value string
	Index int
}
type enum struct {
	Name   string
	Values []Value
}
type MsgFields struct {
	TargetName string
	FieldName  string
	FuncPara   string
	TargetVal  string
}
type UnionObject struct {
	UnionName   string
	UnionFields []string
}
type InputClass struct {
	Name         string
	InputObjName string
	Fields       []MsgFields
}

func scalarMap(scalar string) string {
	switch scalar {
	case "BOOL":
		return "bool"
	case "INT32":
		return "int32"
	case "INT64":
		return "int64"
	case "UINT32":
		return "uint32"
	case "UINT64":
		return "uint64"
	case "SINT32":
		return "int32"
	case "SINT64":
		return "int64"
	case "FIXED32":
		return "uint32"
	case "FIXED64":
		return "uint64"
	case "SFIXED32":
		return "int32"
	case "SFIXED64":
		return "int64"
	case "FLOAT":
		return "float32"
	case "DOUBLE":
		return "float64"
	case "STRING":
		return "string"
	case "BYTES":
		return "[]byte"

	}
	return ""
}
func fieldElementType(valKey pgs.FieldTypeElem) string {
	switch valKey.ProtoType().Proto().String() {
	case "TYPE_MESSAGE":
		obj := valKey.Embed()
		return obj.Name().String()
	case "TYPE_ENUM":
		enum := valKey.Enum()
		return enum.Name().String()
	default:
		tType := strings.Split(valKey.ProtoType().Proto().String(), "_")
		return scalarMap(tType[len(tType)-1])
	}
}

type PayloadFields struct {
	FieldName string
	FuncPara  string
	TargetVal string
}
type OneOfFields struct {
	CaseName   string
	ReturnType string
}
type UnionObjectPayload struct {
	FieldName  string
	FuncReturn string
	SwitchName string
	Fields     []OneOfFields
}
type Payload struct {
	Name         string
	UnionObjects []UnionObjectPayload
	Fields       []PayloadFields
}

func (m *jaalModule) EnumType(enumData pgs.Enum, imports map[string]string, initFunctionsName map[string]bool) (string, error) {

	enumval := enum{Name: enumData.Name().UpperCamelCase().String()}

	initFunctionsName["Register"+enumval.Name] = true

	for i, val := range enumData.Values() {

		enumval.Values = append(enumval.Values, Value{Value: val.Name().String(), Index: i})

	}

	tmp := getEnumTemplate()
	buf := &bytes.Buffer{}

	if err := tmp.Execute(buf, enumval); err != nil {

		return "", err

	}

	return buf.String(), nil
}

type Oneof struct {
	Name                        string
	SchemaObjectPara            string
	FieldFuncPara               string
	FieldFuncSecondParaFuncPara string
	TargetName                  string
}

func (m *jaalModule) OneofInputType(inputData pgs.Message, imports map[string]string, initFunctionsName map[string]bool) (string, error) {
	var oneOfArr []Oneof
	for _, oneof := range inputData.OneOfs() {
		for _, fields := range oneof.Fields() {

			name := fields.Message().Name().UpperCamelCase().String() + "_" + fields.Name().UpperCamelCase().String()
			initFunctionsName["RegisterInput"+name] = true
			schemaObjectPara := fields.Message().Name().LowerCamelCase().String() + fields.Name().UpperCamelCase().String()
			fieldFuncPara := fields.Name().LowerCamelCase().String()
			targetName := fields.Name().UpperCamelCase().String()
			fieldFuncSecondParaFuncPara := m.RPCFieldType(fields)
			oneOfArr = append(oneOfArr, Oneof{Name: name, SchemaObjectPara: schemaObjectPara, FieldFuncPara: fieldFuncPara, TargetName: targetName, FieldFuncSecondParaFuncPara: fieldFuncSecondParaFuncPara})
		}
	}

	tmp := getOneofInputTemplate()
	buf := &bytes.Buffer{}

	if err := tmp.Execute(buf, oneOfArr); err != nil {

		return "", err

	}

	return buf.String(), nil
}

type OneofPayload struct {
	Name                      string
	SchemaObjectPara          string
	FieldFuncPara             string
	FieldFuncSecondFuncReturn string
	FieldFuncReturn           string
}

func (m *jaalModule) OneofPayloadType(inputData pgs.Message, imports map[string]string, initFunctionsName map[string]bool) (string, error) {
	var oneOfArr []OneofPayload
	for _, oneof := range inputData.OneOfs() {
		for _, fields := range oneof.Fields() {

			name := fields.Message().Name().UpperCamelCase().String() + "_" + fields.Name().UpperCamelCase().String()
			initFunctionsName["RegisterPayload"+name] = true
			schemaObjectPara := fields.Message().Name().LowerCamelCase().String() + fields.Name().UpperCamelCase().String()
			fieldFuncPara := fields.Name().LowerCamelCase().String()
			fieldFuncSecondFuncReturn := m.RPCFieldType(fields)
			fieldFuncReturn := fields.Name().UpperCamelCase().String()
			oneOfArr = append(oneOfArr, OneofPayload{Name: name, SchemaObjectPara: schemaObjectPara, FieldFuncPara: fieldFuncPara, FieldFuncReturn: fieldFuncReturn, FieldFuncSecondFuncReturn: fieldFuncSecondFuncReturn})
		}
	}

	tmp := getOneofPayloadTemplate()
	buf := &bytes.Buffer{}

	if err := tmp.Execute(buf, oneOfArr); err != nil {

		return "", err

	}

	return buf.String(), nil
}

func (m *jaalModule) UnionStruct(inputData pgs.Message, imports map[string]string, PossibleReqObjects map[string]bool, initFunctionsName map[string]bool) (string, error) {
	var unionObjects []UnionObject
	for _, oneof := range inputData.OneOfs() {

		unionName := "Union"
		msgName := oneof.Message().Name().UpperCamelCase().String()
		unionName += msgName
		unionName += oneof.Name().UpperCamelCase().String()

		var unionField []string

		for _, fields := range oneof.Fields() {

			unionField = append(unionField, "*"+fields.Message().Name().UpperCamelCase().String()+"_"+fields.Name().UpperCamelCase().String())

		}

		unionObjects = append(unionObjects, UnionObject{UnionName: unionName, UnionFields: unionField})

	}

	tmp := getUnionStructTemplate()
	buf := &bytes.Buffer{}

	if err := tmp.Execute(buf, unionObjects); err != nil {

		return "", err

	}

	return buf.String(), nil

}

func (m *jaalModule) InputType(inputData pgs.Message, imports map[string]string, PossibleReqObjects map[string]bool, initFunctionsName map[string]bool) (string, error) {

	msg := InputClass{Name: inputData.Name().UpperCamelCase().String()}
	if PossibleReqObjects[inputData.Name().String()] {
		msg.InputObjName = InputAppend(inputData.Name().UpperCamelCase().String())
	}else{
		msg.InputObjName = inputData.Name().UpperCamelCase().String()
	}

	initFunctionsName["RegisterInput"+msg.Name] = true
	for _, oneof := range inputData.OneOfs() {

		unionName := "Union"
		msgName := oneof.Message().Name().UpperCamelCase().String()
		unionName += msgName
		unionName += oneof.Name().UpperCamelCase().String()

		for _, fields := range oneof.Fields() {

			msg.Fields = append(msg.Fields, MsgFields{TargetName: oneof.Name().UpperCamelCase().String(), FieldName: oneof.Message().Name().LowerCamelCase().String() + fields.Name().UpperCamelCase().String(), FuncPara: "*" + fields.Message().Name().UpperCamelCase().String() + "_" + fields.Name().UpperCamelCase().String(), TargetVal: "source"})

		}

	}

	for _, fields := range inputData.NonOneOfFields() {

		msgArg := ""
		tVal := ""
		flag := true
		flag2 := true
		flag3 := true

		if fields.Type().IsRepeated() {
			msgArg += "[]"
			flag = false
			if !fields.Type().Element().IsEmbed() {
				flag3 = false
			}
		}

		if flag3 {
			msgArg += "*"
		}

		if strings.ToLower(fields.Name().String()) == "id" {

			msgArg += "schemabuilder.ID"
			tVal += "source.Value"
			flag = false
			flag2 = false

		} else if fields.Type().IsMap() {

			msgArg += "map["
			msgArg += fieldElementType(fields.Type().Key())
			msgArg += "]"
			msgArg += fieldElementType(fields.Type().Element())

		} else if fields.Descriptor().GetType().String() == "TYPE_MESSAGE" {

			if fields.Type().IsEmbed() && fields.Type().Embed().File().Descriptor().Options != nil && fields.Type().Embed().File().Descriptor().Options.GoPackage != nil {
				if inputData.Package().ProtoName().String() != fields.Type().Embed().Package().ProtoName().String() {

					go_pkg := m.GetGoPackage(fields.Type().Embed().File())
					msgArg += go_pkg
					msgArg += "."
				}
			}

			tmsg := strings.Split(fields.Descriptor().GetTypeName(), ".")
			msgArg += tmsg[len(tmsg)-1]

			flag = false

		} else if fields.Descriptor().GetType().String() == "TYPE_ENUM" {

			tmsg := strings.Split(fields.Descriptor().GetTypeName(), ".")
			msgArg += tmsg[len(tmsg)-1]
		} else { // todo till here
			tmsg := strings.Split(fields.Descriptor().GetType().String(), "_")
			msgArg += scalarMap(tmsg[len(tmsg)-1])

		}

		if flag {

			tVal += "*"

		}

		if flag2 {

			tVal += "source"

		}

		targetName := fields.Name().UpperCamelCase().String()
		fieldName := fields.Name().LowerCamelCase().String()
		msg.Fields = append(msg.Fields, MsgFields{TargetName: targetName, FieldName: fieldName, FuncPara: msgArg, TargetVal: tVal})

	}

	tmp := getInputTemplate()
	buf := &bytes.Buffer{}

	if err := tmp.Execute(buf, msg); err != nil {

		return "", err

	}

	return buf.String(), nil
}

func (m *jaalModule) PayloadType(payloadData pgs.Message, imports map[string]string, initFunctionsName map[string]bool) (string, error) {

	msg := Payload{Name: payloadData.Name().UpperCamelCase().String()}
	initFunctionsName["RegisterPayload"+msg.Name] = true
	for _, oneof := range payloadData.OneOfs() {

		var oneofFields []OneOfFields

		for _, fields := range oneof.Fields() {

			caseName := fields.Message().Name().UpperCamelCase().String() + "_" + fields.Name().UpperCamelCase().String()
			returnType := "&Union" + oneof.Message().Name().UpperCamelCase().String() + oneof.Name().UpperCamelCase().String()
			oneofFields = append(oneofFields, OneOfFields{CaseName: caseName, ReturnType: returnType})

		}

		funcpara := "*Union" + oneof.Message().Name().UpperCamelCase().String() + oneof.Name().UpperCamelCase().String()
		fieldName := oneof.Name().LowerCamelCase().String()
		switchName := oneof.Name().UpperCamelCase().String()
		msg.UnionObjects = append(msg.UnionObjects, UnionObjectPayload{FieldName: fieldName, SwitchName: switchName, FuncReturn: funcpara, Fields: oneofFields})

	}

	for _, fields := range payloadData.NonOneOfFields() {

		msgArg := ""
		tVal := ""

		if fields.Type().IsRepeated() {

			msgArg += "[]"

		}

		if strings.ToLower(fields.Name().String()) == "id" {

			msgArg += "schemabuilder.ID"
			tVal += "schemabuilder."
			tVal += strings.ToUpper(fields.Name().String())
			tVal += "{Value: in."
			tVal += fields.Name().UpperCamelCase().String()
			tVal += "}"

		} else if fields.Type().IsMap() {

			msgArg += "map["
			msgArg += fieldElementType(fields.Type().Key())
			msgArg += "]"
			msgArg += fieldElementType(fields.Type().Element())
			tVal += "in."
			tVal += fields.Name().UpperCamelCase().String()

		} else if fields.Descriptor().GetType().String() == "TYPE_MESSAGE" {
			msgArg += "*"
			if fields.Type().IsEmbed() && fields.Type().Embed().File().Descriptor().Options != nil && fields.Type().Embed().File().Descriptor().Options.GoPackage != nil {
				if payloadData.Package().ProtoName().String() != fields.Type().Embed().Package().ProtoName().String() {
					go_pkg := m.GetGoPackage(fields.Type().Embed().File())
					msgArg += go_pkg
					msgArg += "."
				}
			}

			tTypeArr := strings.Split(fields.Descriptor().GetTypeName(), ".")
			msgArg += tTypeArr[len(tTypeArr)-1]

			tVal += "in."
			tVal += fields.Name().UpperCamelCase().String()

		} else if fields.Descriptor().GetType().String() == "TYPE_ENUM" {

			tTypeArr := strings.Split(fields.Descriptor().GetTypeName(), ".")
			msgArg += tTypeArr[len(tTypeArr)-1]
			tVal += "in."
			tVal += fields.Name().UpperCamelCase().String()

		} else {
			tTypeArr := strings.Split(fields.Descriptor().GetType().String(), "_")
			msgArg += scalarMap(tTypeArr[len(tTypeArr)-1])
			tVal += "in."
			tVal += fields.Name().UpperCamelCase().String()

		}

		fieldName := fields.Name().LowerCamelCase().String()
		msg.Fields = append(msg.Fields, PayloadFields{FieldName: fieldName, FuncPara: msgArg, TargetVal: tVal})

	}

	tmp := getPayloadTemplate()
	buf := &bytes.Buffer{}

	if err := tmp.Execute(buf, msg); err != nil {
		return "", err
	}

	return buf.String(), nil
}

type Query struct {
	FieldName          string
	InType             string
	FirstReturnArgType string
	ReturnFunc         string
}
type Mutation struct {
	FieldName          string
	InputType          string
	FirstReturnArgType string
	RequestType        string
	RequestFields      []string
	ResponseType       string
	ReturnType         string
}
type Service struct {
	Name      string
	Queries   []Query
	Mutations []Mutation
}

func InputAppend(str string) string {
	if strings.HasSuffix(strings.ToLower(str), "req") {
		str = str[:len(str)-3]
		str += "Input"
		return str
	} else if strings.HasSuffix(strings.ToLower(str), "request") {
		str = str[:len(str)-7]
		str += "Input"
		return str
	} else {
		return str + "Input"
	}
}

func (m *jaalModule) ServiceInput(service pgs.Service) (string, error) {
	var varQuery []Query
	var varMutation []Mutation

	for _, rpc := range service.Methods() {

		opt := rpc.Descriptor().GetOptions()
		x, err := proto.GetExtension(opt, pbt.E_Schema)
		if opt == nil {
			continue
		}
		if err != nil {
			if err == proto.ErrMissingExtension {
				continue
			}

			return "", err

		}

		option := *x.(*pbt.MethodOptions)

		if option.GetMutation() == "" {
			fieldName := option.GetQuery()
			inType := "*" + rpc.Input().Name().UpperCamelCase().String()
			firstReturnArgType := "*" + rpc.Output().Name().UpperCamelCase().String()
			returnFunc := rpc.Name().UpperCamelCase().String()
			varQuery = append(varQuery, Query{FieldName: fieldName, InType: inType, FirstReturnArgType: firstReturnArgType, ReturnFunc: returnFunc})
		} else if option.GetQuery() == "" {
			fieldName := option.GetMutation()
			inputType := "*" + rpc.Name().UpperCamelCase().String() + "Input"
			firstReturnArgType := "*" + rpc.Name().UpperCamelCase().String() + "Payload"
			returnType := "&" + rpc.Name().UpperCamelCase().String() + "Payload"
			requestType := "&" + rpc.Input().Name().UpperCamelCase().String()
			var requestFields []string
			for _, fields := range rpc.Input().Fields() {
				requestFields = append(requestFields, fields.Name().UpperCamelCase().String())
			}
			responseType := rpc.Name().UpperCamelCase().String()
			varMutation = append(varMutation, Mutation{FieldName: fieldName, InputType: inputType, FirstReturnArgType: firstReturnArgType, RequestType: requestType, RequestFields: requestFields, ResponseType: responseType, ReturnType: returnType})

		}
	}
	name := service.Name().UpperCamelCase().String()
	varService := Service{Name: name, Queries: varQuery, Mutations: varMutation}
	tmp := getServiceTemplate()
	buf := &bytes.Buffer{}

	if err := tmp.Execute(buf, varService); err != nil {
		return "", err
	}

	return buf.String(), nil

}

type InputField struct {
	Name string
	Type string
}
type InputServiceStruct struct {
	RpcName     string
	InputFields []InputField
}

func (m *jaalModule) RPCFieldType(field pgs.Field) string {
	switch field.Descriptor().GetType().String() {
	case "TYPE_MESSAGE":
		if field.Type().IsRepeated() {
			tTypeArr := strings.Split(*field.Descriptor().TypeName, ".")
			typeName := tTypeArr[len(tTypeArr)-1]
			return "*" + typeName
		} else if field.Type().IsMap() {
			return "*[" + fieldElementType(field.Type().Key()) + "]" + fieldElementType(field.Type().Element())
		}
		obj := field.Type().Embed()
		return "*" + obj.Name().String()
	case "TYPE_ENUM":
		if field.Type().IsRepeated() {
			tTypeArr := strings.Split(*field.Descriptor().TypeName, ".")
			typeName := tTypeArr[len(tTypeArr)-1]
			return typeName
		}
		enum := field.Type().Enum()
		return enum.Name().String()
	default:
		tTypeArr := strings.Split(field.Descriptor().GetType().String(), "_")
		scalarType := tTypeArr[len(tTypeArr)-1]
		return scalarMap(scalarType)
	}
}
func (m *jaalModule)checkImportedField(service pgs.Service,field pgs.Field)string{

	if service.Package().ProtoName().String() != field.Package().ProtoName().String(){
		goPkg:= m.GetGoPackage(field.File()) //.Type().Embed().File()
		return goPkg+ "."
	}
	return ""
}
func (m *jaalModule) ServiceStructInput(service pgs.Service) (string, error) {
	var inputServiceStruct []InputServiceStruct
	for _, rpc := range service.Methods() {
		opt := rpc.Descriptor().GetOptions()
		if opt == nil {
			continue
		}
		x, err := proto.GetExtension(opt, pbt.E_Schema)
		if err != nil {
			if err == proto.ErrMissingExtension {
				continue
			}
			return "", err
		}
		option := *x.(*pbt.MethodOptions)
		if option.GetMutation() == "" {
			continue
		}
		tInputServiceSTruct := InputServiceStruct{RpcName: rpc.Name().UpperCamelCase().String()}
		for _, ipField := range rpc.Input().Fields() {
			name := ipField.Name().UpperCamelCase().String()
			ttype:= m.RPCFieldType(ipField)

			go_pkg:=""
			if ipField.Type().IsEmbed() && ipField.Type().Embed().File().Descriptor().Options != nil && ipField.Type().Embed().File().Descriptor().Options.GoPackage != nil {
				if service.Package().ProtoName().String() != ipField.Type().Embed().Package().ProtoName().String() {

					go_pkg = m.GetGoPackage(ipField.Type().Embed().File())+"."
				}
			}

			if ttype[0]=='*'{
				ttype= "*"+ go_pkg+ttype[1:len(ttype)]
			}else{
				ttype =  go_pkg+ttype
			}

			tInputServiceSTruct.InputFields = append(tInputServiceSTruct.InputFields, InputField{Name: name, Type: ttype})
		}
		inputServiceStruct = append(inputServiceStruct, tInputServiceSTruct)
	}
	tmp := getServiceStructInputTemplate()
	buf := &bytes.Buffer{}

	if err := tmp.Execute(buf, inputServiceStruct); err != nil {
		return "", err
	}

	return buf.String(), nil
}

type PayloadServiceStruct struct {
	Name       string
	ReturnType string
}

func (m *jaalModule)checkImported(service pgs.Service,message pgs.Message)string{

	if service.Package().ProtoName().String() != message.Package().ProtoName().String(){
		goPkg:= m.GetGoPackage(message.File()) //.Type().Embed().File()
		return goPkg+ "."
	}
	return ""
}

func (m *jaalModule) ServiceStructPayload(service pgs.Service) (string, error) {
	var payloadService []PayloadServiceStruct
	for _, rpc := range service.Methods() {
		opt := rpc.Descriptor().GetOptions()
		if opt == nil {
			continue
		}
		x, err := proto.GetExtension(opt, pbt.E_Schema)
		if err != nil {
			if err == proto.ErrMissingExtension {
				continue
			}
			return "", err
		}
		option := *x.(*pbt.MethodOptions)
		if option.GetMutation() == "" {
			continue
		}
		returnType:= "*"+ m.checkImported(service,rpc.Output())+ rpc.Output().Name().UpperCamelCase().String()
		payloadService = append(payloadService, PayloadServiceStruct{Name: rpc.Name().UpperCamelCase().String(), ReturnType:  returnType})
	}
	tmp := getServiceStructPayloadTemplate()
	buf := &bytes.Buffer{}

	if err := tmp.Execute(buf, payloadService); err != nil {
		return "", err
	}

	return buf.String(), nil
}
func (m *jaalModule) getPossibleReqObjects(service pgs.Service, PossibleReqObjects map[string]bool) error {
	for _, rpc := range service.Methods() {
		opt := rpc.Descriptor().GetOptions()
		if opt == nil {
			continue
		}
		x, err := proto.GetExtension(opt, pbt.E_Schema)

		if err != nil {
			if err == proto.ErrMissingExtension {
				continue
			}
			return err

		}

		option := *x.(*pbt.MethodOptions)

		if option.GetQuery() != "" {
			PossibleReqObjects[rpc.Input().Name().String()] = true
		}
	}
	return nil
}
func (m *jaalModule) InitFunc(initFunctionsName map[string]bool) (string, error) {
	tmp := getInitTemplate()
	buf := &bytes.Buffer{}

	if err := tmp.Execute(buf, initFunctionsName); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (m *jaalModule) GetGoPackage(target pgs.File) string {
	goPackage := "pb"

	if target.Descriptor().GetOptions() != nil && target.Descriptor().GetOptions().GoPackage != nil {
		goPackage = *target.Descriptor().GetOptions().GoPackage
		goPackage = strings.Split(goPackage, ";")[0]
		goPackage = strings.Split(goPackage, "/")[len(strings.Split(goPackage, "/"))-1]
	}
	return goPackage
}
func (m *jaalModule) GetImports(target pgs.File) map[string]string {
	imports := make(map[string]string)
	for _, importFile := range target.Imports() {
		if importFile.Descriptor().Options!=nil && importFile.Descriptor().Options.GoPackage!=nil{
			key := *importFile.Descriptor().Options.GoPackage
			key = strings.Split(key, ";")[0]
			imports[key] = strings.Split(key, "/")[len(strings.Split(key, "/"))-1]
		}
	}
	return imports
}

func (m *jaalModule) ServiceStructInputFunc(service pgs.Service, initFunctionsName map[string]bool) (string, error) {
	var inputServiceStructFunc []InputClass
	for _, rpc := range service.Methods() {
		opt := rpc.Descriptor().GetOptions()
		if opt == nil {
			continue
		}
		x, err := proto.GetExtension(opt, pbt.E_Schema)
		if err != nil {
			if err == proto.ErrMissingExtension {
				continue
			}
			return "", err
		}
		option := *x.(*pbt.MethodOptions)
		if option.GetMutation() == "" {
			continue
		}
		var field []MsgFields
		for _, ipField := range rpc.Input().Fields() {
			tname := ipField.Name().UpperCamelCase().String()
			if ipField.Package().ProtoName() != service.Package().ProtoName() {
				tname = m.GetGoPackage(ipField.File()) + "." + ipField.Name().UpperCamelCase().String()
			}

			fName := ipField.Name().LowerCamelCase().String()
			tval := ""
			if ipField.Descriptor().GetType().String() == "TYPE_MESSAGE" {
				tval = "source"
			} else {
				tval = "*source"
			}
			funcPara := m.RPCFieldType(ipField)
			if funcPara[0] == '*' {
				funcPara = funcPara[1:len(funcPara)]
			}
			field = append(field, MsgFields{TargetName: tname, FieldName: fName, FuncPara: funcPara, TargetVal: tval})
		}
		initFunctionsName["RegisterInput"+rpc.Name().UpperCamelCase().String()+"Input"] = true
		inputServiceStructFunc = append(inputServiceStructFunc, InputClass{Name: rpc.Name().UpperCamelCase().String(), Fields: field})
	}
	tmp := getServiceStructInputFuncTemplate()
	buf := &bytes.Buffer{}

	if err := tmp.Execute(buf, inputServiceStructFunc); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (m *jaalModule) ServiceStructPayloadFunc(service pgs.Service, initFunctionsName map[string]bool) (string, error) {
	var payloadService []PayloadServiceStruct
	for _, rpc := range service.Methods() {
		opt := rpc.Descriptor().GetOptions()
		if opt == nil {
			continue
		}
		x, err := proto.GetExtension(opt, pbt.E_Schema)
		if err != nil {
			if err == proto.ErrMissingExtension {
				continue
			}
			return "", err
		}
		option := *x.(*pbt.MethodOptions)
		if option.GetMutation() == "" {
			continue
		}
		initFunctionsName["RegisterPayload"+rpc.Name().UpperCamelCase().String()+"Payload"] = true
		payloadService = append(payloadService, PayloadServiceStruct{Name: rpc.Name().UpperCamelCase().String(), ReturnType: "*" + rpc.Output().Name().UpperCamelCase().String()})
	}
	tmp := getServiceStructPayloadFuncTemplate()
	buf := &bytes.Buffer{}

	if err := tmp.Execute(buf, payloadService); err != nil {
		return "", err
	}

	return buf.String(), nil
}
