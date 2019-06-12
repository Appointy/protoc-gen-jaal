package main

import (
	"bytes"
	"os"
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

type InputMap struct {
	FieldName  string
	TargetName string
	TargetVal  string
	Key        string
	Value      string
}

type InputClass struct {
	Name         string
	InputObjName string
	Maps         []InputMap
	Fields       []MsgFields
}

func (m *jaalModule) scalarMap(scalar string) string {
	// Maps protoc scalars to go scalars

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
func (m *jaalModule) fieldElementType(valKey pgs.FieldTypeElem) string {
	// returns Type for a pgs.FieldTypeElem

	switch valKey.ProtoType().Proto().String() {
	case "TYPE_MESSAGE":

		obj := valKey.Embed()
		return obj.Name().String()

	case "TYPE_ENUM":

		enum := valKey.Enum()
		return enum.Name().String()

	default:

		tType := strings.Split(valKey.ProtoType().Proto().String(), "_")
		return m.scalarMap(tType[len(tType)-1])

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
type PayloadMap struct {
	FieldName string
	TargetVal string
}
type Payload struct {
	Name           string
	PayloadObjName string
	UnionObjects   []UnionObjectPayload
	Maps           []PayloadMap
	Fields         []PayloadFields
}

func (m *jaalModule) EnumType(enumData pgs.Enum, imports map[string]string, initFunctionsName map[string]bool) (string, error) {
	// returns generated template in for a enum type

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
	TargetVal                   string
}

func (m *jaalModule) GetSkipOption(message pgs.Message) (bool, error) {
	//returns true if a message have skip flag as true or no skip option
	//returns false if message have skip flag as false

	opt := message.Descriptor().GetOptions()
	x, err := proto.GetExtension(opt, pbt.E_Skip)

	if opt == nil {

		return false, nil

	}

	if err != nil {

		if err == proto.ErrMissingExtension {

			return false, nil

		}

		return false, err

	}

	return *x.((*bool)), nil
}

func (m *jaalModule) OneofInputType(inputData pgs.Message, imports map[string]string, initFunctionsName map[string]bool) (string, error) {
	/*
		returns generated template(Input) in for a Oneof type
		Primitive,Enum and object type are handled inside a oneof
	*/
	var oneOfArr []Oneof

	for _, oneof := range inputData.OneOfs() {

		for _, fields := range oneof.Fields() {

			name := fields.Message().Name().UpperCamelCase().String() + "_" + fields.Name().UpperCamelCase().String()
			initFunctionsName["RegisterInput"+name] = true
			schemaObjectPara := fields.Message().Name().LowerCamelCase().String() + fields.Name().UpperCamelCase().String()
			fieldFuncPara := fields.Name().LowerCamelCase().String()
			targetName := fields.Name().UpperCamelCase().String()
			fieldFuncSecondParaFuncPara := m.RPCFieldType(fields)
			if fieldFuncSecondParaFuncPara[0] == '*' {
				fieldFuncSecondParaFuncPara = fieldFuncSecondParaFuncPara[1:len(fieldFuncSecondParaFuncPara)]
			}
			goPkg := ""
			targetVal := ""
			if fields.Type().IsEnum() {
				goPkg = m.GetGoPackageOfFiles(inputData.File(), fields.Type().Enum().File())
				if goPkg != "" {
					goPkg += "."
				}
				targetVal = "*source"
			} else if fields.Type().IsEmbed() {
				goPkg = m.GetGoPackageOfFiles(inputData.File(), fields.Type().Embed().File())
				if goPkg != "" {
					goPkg += "."
				}
				targetVal = "source"
			} else {
				targetVal = "*source"
			}
			fieldFuncSecondParaFuncPara = goPkg + fieldFuncSecondParaFuncPara
			if fieldFuncSecondParaFuncPara == "field_mask.FieldMask" {
				targetVal = "gtypes.ModifyFieldMask(source)"
			}
			oneOfArr = append(oneOfArr, Oneof{TargetVal: targetVal, Name: name, SchemaObjectPara: schemaObjectPara, FieldFuncPara: fieldFuncPara, TargetName: targetName, FieldFuncSecondParaFuncPara: fieldFuncSecondParaFuncPara})
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
	/*
		returns generated template(Payload) in for all oneOf type
		Primitive,Enum and object type are handled inside a oneof
	*/
	var oneOfArr []OneofPayload

	for _, oneof := range inputData.OneOfs() {

		for _, fields := range oneof.Fields() {

			name := fields.Message().Name().UpperCamelCase().String() + "_" + fields.Name().UpperCamelCase().String()
			initFunctionsName["RegisterPayload"+name] = true
			schemaObjectPara := fields.Message().Name().LowerCamelCase().String() + fields.Name().UpperCamelCase().String()
			fieldFuncPara := fields.Name().LowerCamelCase().String()
			fieldFuncSecondFuncReturn := m.RPCFieldType(fields)
			if fieldFuncSecondFuncReturn[0] == '*' {
				fieldFuncSecondFuncReturn = fieldFuncSecondFuncReturn[1:len(fieldFuncSecondFuncReturn)]
			}
			goPkg := ""
			if fields.Type().IsEnum() {
				goPkg = m.GetGoPackageOfFiles(inputData.File(), fields.Type().Enum().File())
				if goPkg != "" {
					goPkg += "."
				}
			} else if fields.Type().IsEmbed() {
				goPkg = m.GetGoPackageOfFiles(inputData.File(), fields.Type().Embed().File())
				if goPkg != "" {
					goPkg = "*" + goPkg
					goPkg += "."
				} else {
					goPkg = "*"
				}
			}
			fieldFuncSecondFuncReturn = goPkg + fieldFuncSecondFuncReturn
			fieldFuncReturn := fields.Name().UpperCamelCase().String()
			if fieldFuncSecondFuncReturn == "*field_mask.FieldMask" {
				fieldFuncReturn = "gtypes.ModifyFieldMask(in." + fieldFuncReturn + ")"
			} else {
				fieldFuncReturn = "in." + fieldFuncReturn
			}
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
	//returns generated template(Union Struct) in for all one of type

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

func (m *jaalModule) GetMessageName(message pgs.Message) (string, error) {
	opt := message.Descriptor().GetOptions() // checks Name option
	if opt != nil {
		x, err := proto.GetExtension(opt, pbt.E_Name)
		if err != nil && err != proto.ErrMissingExtension {
			return "", err
		}
		if x != nil {
			return *x.(*string), nil;
		}
	}
	return "", nil
}

func (m *jaalModule) InputType(inputData pgs.Message, imports map[string]string, PossibleReqObjects map[string]bool, initFunctionsName map[string]bool) (string, error) {
	// returns generated template(Input) in for a message type

	if skip, err := m.GetSkipOption(inputData); err != nil {
		// checks skip flag
		return "", err

	} else if skip {

		return "", nil
	}

	fullyQualifiedName := inputData.FullyQualifiedName()
	embeddedMessageParent := ""
	if inputData.Parent().Name().String() == strings.Split(fullyQualifiedName, ".")[len(strings.Split(fullyQualifiedName, "."))-2] {
		lenPackage:= len(inputData.Package().ProtoName().String())
		endlength:=len(inputData.Name().String())
		embeddedMessageParent = strings.Replace(fullyQualifiedName[lenPackage+2:len(fullyQualifiedName)-endlength],".","_",-1)
	}

	msg := InputClass{Name: embeddedMessageParent + inputData.Name().UpperCamelCase().String()}

	newName, err := m.GetMessageName(inputData)

	if err != nil {

		return "", err

	} else if newName != "" {

		msg.InputObjName = newName + "Input"

	} else if PossibleReqObjects[inputData.Name().String()] {

		msg.InputObjName = m.InputAppend(inputData.Name().UpperCamelCase().String())

	} else {

		msg.InputObjName = inputData.Name().UpperCamelCase().String() + "Input"

	}

	initFunctionsName["RegisterInput"+msg.Name] = true

	var maps []InputMap

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
		//m.Log(fields.Name(),fields.Type().IsEmbed())

		msgArg := ""
		tVal := ""
		flag := true
		flag2 := true
		flag3 := true
		fieldName := fields.Name().LowerCamelCase().String()
		targetName := fields.Name().UpperCamelCase().String()

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

		} else if fields.Type().IsRepeated() {

			msgArg = "[]"
			tObj := fields.Type().Element()

			if tObj.IsEmbed() {

				msgArg += "*"

			}

			if tObj.IsEmbed(){

				if  tObj.Embed().File().Descriptor().Options != nil && tObj.Embed().File().Descriptor().Options.GoPackage != nil && inputData.Package().ProtoName().String() != tObj.Embed().Package().ProtoName().String() {

					msgArg += m.GetGoPackage(tObj.Embed().File())
					msgArg += "."

				}else{
					if strings.Split(tObj.Embed().FullyQualifiedName(), ".")[len(strings.Split(tObj.Embed().FullyQualifiedName(), "."))-2] == tObj.Embed().Parent().Name().String(){
						tlenPackage:= len(fields.Package().ProtoName().String())
						tendlength:=len(fields.Name().String())
						tembeddedMessageParent := strings.Replace(fields.FullyQualifiedName()[tlenPackage+2:len(fields.FullyQualifiedName())-tendlength],".","_",-1)
						msgArg += tembeddedMessageParent
					}
				}
			}

			ttype := m.fieldElementType(tObj)
			msgArg += ttype

		} else if fields.Type().IsMap() {
			// TODO : Repeated case not handled

			goPkg := ""
			if fields.Type().Element().IsEmbed() {
				goPkg = m.GetGoPackageOfFiles(inputData.File(), fields.Type().Element().Embed().File())
				if goPkg != "" {
					goPkg += "."
				}
			}

			asterik := ""
			if fields.Type().Element().IsEmbed() {
				asterik = "*"
			}

			value := asterik + goPkg + m.fieldElementType(fields.Type().Element())
			maps = append(maps, InputMap{FieldName: fieldName, TargetVal: "*source", TargetName: targetName, Key: m.fieldElementType(fields.Type().Key()), Value: value})
			continue
		} else if fields.Descriptor().GetType().String() == "TYPE_MESSAGE" {

			if fields.Type().IsEmbed() {
				goPkg := m.GetGoPackageOfFiles(inputData.File(), fields.Type().Embed().File())
				if goPkg != "" {
					msgArg += goPkg
					msgArg += "."
				} else {
					if strings.Split(fields.FullyQualifiedName(), ".")[len(strings.Split(fields.FullyQualifiedName(), "."))-2] == fields.Type().Embed().Parent().Name().String() {
						tlenPackage:= len(fields.Package().ProtoName().String())
						tendlength:=len(fields.Name().String())
						tembeddedMessageParent := strings.Replace(fields.FullyQualifiedName()[tlenPackage+2:len(fields.FullyQualifiedName())-tendlength],".","_",-1)
						msgArg += tembeddedMessageParent
					}
				}
			}

			tmsg := strings.Split(fields.Descriptor().GetTypeName(), ".")
			msgArg += tmsg[len(tmsg)-1]

			flag = false

		} else if fields.Descriptor().GetType().String() == "TYPE_ENUM" {
			goPkg := m.GetGoPackageOfFiles(inputData.File(), fields.Type().Enum().File())
			if goPkg != "" {
				goPkg += "."
			}
			flag = false
			tmsg := strings.Split(fields.Descriptor().GetTypeName(), ".")
			msgArg = goPkg + tmsg[len(tmsg)-1]

		} else {

			tmsg := strings.Split(fields.Descriptor().GetType().String(), "_")
			msgArg += m.scalarMap(tmsg[len(tmsg)-1])

		}

		if flag {

			tVal += "*"

		}

		if flag2 {

			tVal += "source"

		}

		if strings.HasSuffix(msgArg, "timestamp.Timestamp") {
			// handles special case of timestamp
			//todo repeated case is not handled

			msgArg = msgArg[:len(msgArg)-19]
			msgArg += "schemabuilder.Timestamp"

			tVal = "(*timestamp.Timestamp)(" + tVal + ")"
		} else if strings.HasSuffix(msgArg, "field_mask.FieldMask") {
			tVal = "gtypes.ModifyFieldMask(" + tVal + ")"
		}
		msg.Fields = append(msg.Fields, MsgFields{TargetName: targetName, FieldName: fieldName, FuncPara: msgArg, TargetVal: tVal})

	}
	// adds all maps
	msg.Maps = maps
	tmp := getInputTemplate()
	buf := &bytes.Buffer{}

	if err := tmp.Execute(buf, msg); err != nil {

		return "", err

	}

	return buf.String(), nil
}

func (m *jaalModule) PayloadType(payloadData pgs.Message, imports map[string]string, initFunctionsName map[string]bool) (string, error) {
	// returns generated template(Payload) in for a message type

	if skip, err := m.GetSkipOption(payloadData); err != nil {
		// checks skip flag
		return "", err

	} else if skip {

		return "", nil
	}

	fullyQualifiedName := payloadData.FullyQualifiedName()
	embeddedMessageParent := ""
	if payloadData.Parent().Name().String() == strings.Split(fullyQualifiedName, ".")[len(strings.Split(fullyQualifiedName, "."))-2] {
		lenPackage:= len(payloadData.Package().ProtoName().String())
		endlength:=len(payloadData.Name().String())
		embeddedMessageParent = strings.Replace(fullyQualifiedName[lenPackage+2:len(fullyQualifiedName)-endlength],".","_",-1)
	}
	msg := Payload{Name: embeddedMessageParent + payloadData.Name().UpperCamelCase().String()}
	newName, err := m.GetMessageName(payloadData)
	if err != nil {
		return "", err
	} else if newName != "" {
		msg.PayloadObjName = newName
	} else {
		msg.PayloadObjName = payloadData.Name().UpperCamelCase().String()
	}
	initFunctionsName["RegisterPayload"+msg.Name] = true
	var maps []PayloadMap
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
		fieldName := fields.Name().LowerCamelCase().String()

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

		} else if fields.Type().IsRepeated() {

			msgArg = "[]"
			tObj := fields.Type().Element()

			if tObj.IsEmbed() {

				msgArg += "*"

			}

			if tObj.IsEmbed(){

				if tObj.Embed().File().Descriptor().Options != nil && tObj.Embed().File().Descriptor().Options.GoPackage != nil  && payloadData.Package().ProtoName().String() != tObj.Embed().Package().ProtoName().String() {

					msgArg += m.GetGoPackage(tObj.Embed().File())
					msgArg += "."
				}else{
					if strings.Split(tObj.Embed().FullyQualifiedName(), ".")[len(strings.Split(tObj.Embed().FullyQualifiedName(), "."))-2] == tObj.Embed().Parent().Name().String(){
						tlenPackage:= len(fields.Package().ProtoName().String())
						tendlength:=len(fields.Name().String())
						tembeddedMessageParent := strings.Replace(fields.FullyQualifiedName()[tlenPackage+2:len(fields.FullyQualifiedName())-tendlength],".","_",-1)
						msgArg += tembeddedMessageParent
					}
				}

			}

			ttype := m.fieldElementType(tObj)
			msgArg += ttype
			tVal += "in." + fields.Name().UpperCamelCase().String()

		} else if fields.Type().IsMap() {
			//TODO : Repeated case not handled

			tVal += "in."
			tVal += fields.Name().UpperCamelCase().String()
			maps = append(maps, PayloadMap{FieldName: fieldName, TargetVal: tVal})
			continue

		} else if fields.Descriptor().GetType().String() == "TYPE_MESSAGE" {

			msgArg += "*"

			if fields.Type().IsEmbed() {
				goPkg := m.GetGoPackageOfFiles(payloadData.File(), fields.Type().Embed().File())
				if goPkg != "" {
					msgArg += goPkg
					msgArg += "."
				} else {
					if strings.Split(fields.FullyQualifiedName(), ".")[len(strings.Split(fields.FullyQualifiedName(), "."))-2] == fields.Type().Embed().Parent().Name().String() {
						tlenPackage:= len(fields.Package().ProtoName().String())
						tendlength:=len(fields.Name().String())
						tembeddedMessageParent := strings.Replace(fields.FullyQualifiedName()[tlenPackage+2:len(fields.FullyQualifiedName())-tendlength],".","_",-1)
						msgArg += tembeddedMessageParent
					}
				}
			}

			tTypeArr := strings.Split(fields.Descriptor().GetTypeName(), ".")
			msgArg += tTypeArr[len(tTypeArr)-1]

			tVal += "in."
			tVal += fields.Name().UpperCamelCase().String()

		} else if fields.Descriptor().GetType().String() == "TYPE_ENUM" {
			goPkg := m.GetGoPackageOfFiles(payloadData.File(), fields.Type().Enum().File())
			if goPkg != "" {
				goPkg += "."
			}

			tTypeArr := strings.Split(fields.Descriptor().GetTypeName(), ".")
			msgArg = goPkg + tTypeArr[len(tTypeArr)-1]
			tVal += "in."
			tVal += fields.Name().UpperCamelCase().String()

		} else {
			goPkg := m.GetGoPackageOfFiles(payloadData.File(), fields.File())
			if goPkg != "" {
				goPkg += "."
			}
			tTypeArr := strings.Split(fields.Descriptor().GetType().String(), "_")
			msgArg = goPkg + m.scalarMap(tTypeArr[len(tTypeArr)-1])
			tVal += "in."
			tVal += fields.Name().UpperCamelCase().String()

		}
		if strings.HasSuffix(msgArg, "timestamp.Timestamp") {
			// handles special case of timestamp
			//todo repeated case is not handled
			msgArg = msgArg[:len(msgArg)-19]
			msgArg += "schemabuilder.Timestamp"

			tVal = "(*schemabuilder.Timestamp)(" + tVal + ")"
		} else if strings.HasSuffix(msgArg, "field_mask.FieldMask") {
			tVal = "gtypes.ModifyFieldMask(" + tVal + ")"
		}

		msg.Fields = append(msg.Fields, PayloadFields{FieldName: fieldName, FuncPara: msgArg, TargetVal: tVal})

	}

	// adds all maps
	msg.Maps = maps

	tmp := getPayloadTemplate()
	buf := &bytes.Buffer{}

	if err := tmp.Execute(buf, msg); err != nil {

		return "", err

	}

	return buf.String(), nil
}

type MapData struct {
	Name       string
	Key        string
	Value      string
	NewVarName string
}

type Fields struct {
	Name string
	Type string
}

type Query struct {
	FieldName          string
	InType             []Fields
	InputName          string
	ReturnType         []Fields
	FirstReturnArgType string
	ReturnFunc         string
	MapsData           []MapData
	Oneofs             []OneOfMutation
}

type Mutation struct {
	FieldName          string
	InputType          string
	FirstReturnArgType string
	RequestType        string
	RequestFields      []string
	ResponseType       string
	ReturnType         string
	OneOfs             []OneOfMutation
}

type OneOfMutation struct {
	Name   string
	Fields []Fields
}
type Service struct {
	Name      string
	Queries   []Query
	Mutations []Mutation
}

func (m *jaalModule) InputAppend(str string) string {
	// returns input object name for input type

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

func (m *jaalModule) GetOption(rpc pgs.Method) (bool, pbt.MethodOptions, error) {
	//returns method option for a rpc method (Used to get query and mutation data)

	opt := rpc.Descriptor().GetOptions()
	x, err := proto.GetExtension(opt, pbt.E_Schema)

	if opt == nil {

		return false, pbt.MethodOptions{}, nil

	}

	if err != nil {

		if err == proto.ErrMissingExtension {

			return false, pbt.MethodOptions{}, nil

		}

		return false, pbt.MethodOptions{}, err

	}

	option := *x.(*pbt.MethodOptions)

	return true, option, nil
}

func (m *jaalModule) ServiceInput(service pgs.Service) (string, error) {
	// returns generated template(Service) in for a service type

	var varQuery []Query
	var varMutation []Mutation

	for _, rpc := range service.Methods() {

		flag, option, err := m.GetOption(rpc)

		if err != nil {

			m.Log("Error", err)
			os.Exit(0)

		}

		if flag == false {
			continue

		}
		//todo case for no query and mutation
		if option.GetMutation() == "" {

			fieldName := option.GetQuery()
			firstReturnArgType := "*"
			if rpc.Output().Package().ProtoName().String() != service.Package().ProtoName().String() {
				firstReturnArgType += m.GetGoPackage(rpc.Output().File())
				firstReturnArgType += "."
			}
			firstReturnArgType += rpc.Output().Name().UpperCamelCase().String()
			returnFunc := rpc.Name().UpperCamelCase().String()
			var inType []Fields
			var returnType []Fields
			var mapsData []MapData
			var oneOfs []OneOfMutation
			for _, oneOf := range rpc.Input().OneOfs() {
				var fields []Fields
				for _, field := range oneOf.Fields() {
					goPkg := m.GetGoPackageOfFiles(service.File(), field.File())
					if goPkg != "" {
						goPkg += "."
					}
					name := field.Name().UpperCamelCase().String()
					ttype := goPkg + rpc.Input().Name().UpperCamelCase().String() + "_" + name
					ttype = "*" + ttype
					fields = append(fields, Fields{Name: name, Type: ttype})
				}
				oneOfs = append(oneOfs, OneOfMutation{Name: oneOf.Name().UpperCamelCase().String(), Fields: fields})
			}
			for _, field := range rpc.Input().Fields() {

				name := field.Name().UpperCamelCase().String()
				tType := ""

				if strings.ToLower(name) == "id" {

					tType = "schemabuilder.ID"

				} else if field.Type().IsRepeated() {

					tType = "[]"
					tObj := field.Type().Element()

					if tObj.IsEmbed() {

						tType += "*"

					}

					if tObj.IsEmbed() && tObj.Embed().File().Descriptor().Options != nil && tObj.Embed().File().Descriptor().Options.GoPackage != nil {

						if service.Package().ProtoName().String() != tObj.Embed().Package().ProtoName().String() {

							tType += m.GetGoPackage(tObj.Embed().File())
							tType += "."

						}else{
							if strings.Split(tObj.Embed().FullyQualifiedName(), ".")[len(strings.Split(tObj.Embed().FullyQualifiedName(), "."))-2] == tObj.Embed().Parent().Name().String(){
								tType+= (tObj.Embed().Parent().Name().String()+"_")
							}
						}

					}

					tType += m.fieldElementType(tObj)

				} else if field.Type().IsMap() {
					tType = "*schemabuilder.Map"
					goPkg := ""
					if field.Type().Element().IsEmbed() {
						goPkg = m.GetGoPackageOfFiles(service.File(), field.Type().Element().Embed().File())
						if goPkg != "" {
							goPkg += "."
						}
					}

					asterik := ""
					if field.Type().Element().IsEmbed() {
						asterik = "*"
					}

					value := asterik + goPkg + m.fieldElementType(field.Type().Element())
					mapsData = append(mapsData, MapData{Name: name, NewVarName: field.Name().LowerCamelCase().String(), Key: m.fieldElementType(field.Type().Key()), Value: value})
					//returnType = "map returns"
				} else if field.Type().IsEmbed() && field.Type().Embed().File().Descriptor().Options != nil && field.Type().Embed().File().Descriptor().Options.GoPackage != nil {
					goPkg := m.GetGoPackageOfFiles(service.File(), field.Type().Embed().File())
					if goPkg != "" {
						tType += goPkg
						tType += "."
					} else {
						if strings.Split(field.FullyQualifiedName(), ".")[len(strings.Split(field.FullyQualifiedName(), "."))-2] == field.Type().Embed().Parent().Name().String() {

							tType += field.Type().Embed().Parent().Name().String() + "_"
						}
					}

					tType = "*" + tType

					funcRType := m.RPCFieldType(field)

					if funcRType[0] == '*' {

						funcRType = funcRType[1:len(funcRType)]

					}

					tType += funcRType

				} else {
					if field.Type().IsEmbed() {

						tType += "*"

					} else if field.Type().IsEnum() {
						goPkg := m.GetGoPackageOfFiles(service.File(), field.Type().Enum().File())
						if goPkg != "" {
							tType += goPkg
							tType += "."
						}

					}

					funcRType := m.RPCFieldType(field)

					if funcRType[0] == '*' {

						funcRType = funcRType[1:len(funcRType)]

					}

					tType += funcRType
				}
				if field.InOneOf() {
					tType = "*"
					if field.Package().ProtoName().String() != service.Package().ProtoName().String() {
						tType += m.GetGoPackage(field.File())
						tType += "."
					}
					tType += rpc.Input().Name().UpperCamelCase().String()
					tType += "_"
					tType += field.Name().UpperCamelCase().String()
				} else if strings.ToLower(name) == "id" {
					returnType = append(returnType, Fields{Name: name, Type: "args.Id.Value"})

				} else if tType == "*timestamp.Timestamp" {
					tType = "*schemabuilder.Timestamp"
					returnType = append(returnType, Fields{Name: name, Type: "(*timestamp.Timestamp)" + "(args." + name + ")"})
				} else if tType == "*field_mask.FieldMask" {
					returnType = append(returnType, Fields{Name: name, Type: "gtypes.ModifyFieldMask" + "(args." + name + ")"})
				} else if field.Type().IsMap() {
					returnType = append(returnType, Fields{Name: name, Type: field.Name().LowerCamelCase().String() + "Map"})
				} else {
					returnType = append(returnType, Fields{Name: name, Type: "args." + name})
				}
				inType = append(inType, Fields{Name: name, Type: tType})
			}
			inputName := ""
			if rpc.Input().Package().ProtoName().String() != service.Package().ProtoName().String() {
				inputName += m.GetGoPackage(rpc.Input().File())
				inputName += "."
			}
			inputName += rpc.Input().Name().UpperCamelCase().String()
			varQuery = append(varQuery, Query{Oneofs: oneOfs, InputName: inputName, MapsData: mapsData, ReturnType: returnType, FieldName: fieldName, InType: inType, FirstReturnArgType: firstReturnArgType, ReturnFunc: returnFunc})

		} else if option.GetQuery() == "" {

			fieldName := option.GetMutation()
			inputType := "*" + rpc.Name().UpperCamelCase().String() + "Input"
			firstReturnArgType := "*" + rpc.Name().UpperCamelCase().String() + "Payload"
			returnType := "&" + rpc.Name().UpperCamelCase().String() + "Payload"
			goPkg := m.GetGoPackageOfFiles(service.File(), rpc.Input().File())
			if goPkg != "" {
				goPkg += "."
			}

			requestType := "&" + goPkg + rpc.Input().Name().UpperCamelCase().String()
			var requestFields []string
			var oneOfMutation []OneOfMutation
			for _, oneOf := range rpc.Input().OneOfs() {
				var fields []Fields
				for _, field := range oneOf.Fields() {
					goPkg := m.GetGoPackageOfFiles(service.File(), field.File())
					if goPkg != "" {
						goPkg += "."
					}
					name := field.Name().UpperCamelCase().String()
					ttype := goPkg + rpc.Input().Name().UpperCamelCase().String() + "_" + name
					ttype = "*" + ttype
					fields = append(fields, Fields{Name: name, Type: ttype})
				}
				oneOfMutation = append(oneOfMutation, OneOfMutation{Name: oneOf.Name().UpperCamelCase().String(), Fields: fields})
			}
			for _, fields := range rpc.Input().NonOneOfFields() {

				requestFields = append(requestFields, fields.Name().UpperCamelCase().String())

			}

			responseType := rpc.Name().UpperCamelCase().String()
			varMutation = append(varMutation, Mutation{OneOfs: oneOfMutation, FieldName: fieldName, InputType: inputType, FirstReturnArgType: firstReturnArgType, RequestType: requestType, RequestFields: requestFields, ResponseType: responseType, ReturnType: returnType})

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
	//returns type of a rpc field(pgs.Field type)

	switch field.Descriptor().GetType().String() {

	case "TYPE_MESSAGE":

		if field.Type().IsRepeated() {

			tTypeArr := strings.Split(*field.Descriptor().TypeName, ".")
			typeName := tTypeArr[len(tTypeArr)-1]

			return "*" + typeName

		} else if field.Type().IsMap() {

			return "*[" + m.fieldElementType(field.Type().Key()) + "]" + m.fieldElementType(field.Type().Element())

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

		return m.scalarMap(scalarType)

	}
}

func (m *jaalModule) ServiceStructInput(service pgs.Service) (string, error) {
	//returns template(Service input struct) for a service

	var inputServiceStruct []InputServiceStruct

	for _, rpc := range service.Methods() {

		flag, option, err := m.GetOption(rpc)

		if err != nil {

			m.Log("Error", err)
			os.Exit(0)

		}

		if flag == false {

			continue

		}

		if option.GetMutation() == "" {

			continue

		}

		tInputServiceSTruct := InputServiceStruct{RpcName: rpc.Name().UpperCamelCase().String()}
		for _, ipField := range rpc.Input().Fields() {

			name := ipField.Name().UpperCamelCase().String()
			ttype := m.RPCFieldType(ipField)

			if ipField.Type().IsRepeated() {

				ttype = "[]"
				tObj := ipField.Type().Element()

				if tObj.IsEmbed() {

					ttype += "*"

				}

				if tObj.IsEmbed() && tObj.Embed().File().Descriptor().Options != nil && tObj.Embed().File().Descriptor().Options.GoPackage != nil {

					if service.Package().ProtoName().String() != tObj.Embed().Package().ProtoName().String() {

						ttype += m.GetGoPackage(tObj.Embed().File())
						ttype += "."
					}else{
						if strings.Split(tObj.Embed().FullyQualifiedName(), ".")[len(strings.Split(tObj.Embed().FullyQualifiedName(), "."))-2] == tObj.Embed().Parent().Name().String(){
							ttype+= (tObj.Embed().Parent().Name().String()+"_")
						}
					}
				}

				ttype += m.fieldElementType(tObj)
			} else if ipField.Type().IsMap() {
				goPkg := ""

				if ipField.Type().Element().IsEmbed() {
					goPkg = m.GetGoPackageOfFiles(service.File(), ipField.Type().Element().Embed().File())
					if goPkg != "" {
						goPkg += "."
					}
				}

				asterik := ""
				if ipField.Type().Element().IsEmbed() {
					asterik = "*"
				}

				value := asterik + goPkg + m.fieldElementType(ipField.Type().Element())
				ttype = "map[" + m.fieldElementType(ipField.Type().Key()) + "]" + value
			} else {

				goPkg := ""

				if ipField.Type().IsEmbed() {
					goPkg = m.GetGoPackageOfFiles(service.File(), ipField.Type().Embed().File())
					if goPkg != "" {
						goPkg += "."
					} else {
						if strings.Split(ipField.FullyQualifiedName(), ".")[len(strings.Split(ipField.FullyQualifiedName(), "."))-2] == ipField.Type().Embed().Parent().Name().String() {
							goPkg += ipField.Type().Embed().Parent().Name().String() + "_"
						}
					}
				} else if ipField.Type().IsEnum() {
					goPkg = m.GetGoPackageOfFiles(service.File(), ipField.Type().Enum().File())
					if goPkg != "" {
						goPkg += "."
					}
				}

				if ttype[0] == '*' {

					ttype = "*" + goPkg + ttype[1:len(ttype)]

				} else {

					ttype = goPkg + ttype

				}
			}
			if ipField.InOneOf() {
				// handles one of fields
				goPkg := m.GetGoPackageOfFiles(service.File(), ipField.OneOf().File())
				if goPkg != "" {
					goPkg += "."
				}
				ttype = goPkg + rpc.Input().Name().UpperCamelCase().String() + "_" + name
				ttype = "*" + ttype
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

func (m *jaalModule) GetGoPackageOfFiles(file1 pgs.File, file2 pgs.File) (string) {
	/*
		If file2 have different package than file1 then package of file2 is  returned
		else empty string is returned
	*/
	if file1.Package().ProtoName().String() != file2.Package().ProtoName().String() {
		return m.GetGoPackage(file2)
	}
	return ""
}

func (m *jaalModule) ServiceStructPayload(service pgs.Service) (string, error) {
	/*
		returns template(Service payload struct) for a service
	*/
	var payloadService []PayloadServiceStruct

	for _, rpc := range service.Methods() {

		flag, option, err := m.GetOption(rpc)

		if err != nil {

			m.Log("Error", err)
			os.Exit(0)

		}

		if flag == false {

			continue

		}

		if option.GetMutation() == "" {

			continue

		}
		goPkg := m.GetGoPackageOfFiles(service.File(), rpc.Output().File())

		if goPkg != "" {
			goPkg += "."
		}

		returnType := "*" + goPkg + rpc.Output().Name().UpperCamelCase().String()
		payloadService = append(payloadService, PayloadServiceStruct{Name: rpc.Name().UpperCamelCase().String(), ReturnType: returnType})
	}

	tmp := getServiceStructPayloadTemplate()
	buf := &bytes.Buffer{}

	if err := tmp.Execute(buf, payloadService); err != nil {

		return "", err

	}

	return buf.String(), nil
}

func (m *jaalModule) getPossibleReqObjects(service pgs.Service, PossibleReqObjects map[string]bool) error {
	//saves all rpc inputs where rpcs are query type (used to append Input and remove Req), in PossibleReqObjects map

	for _, rpc := range service.Methods() {

		flag, option, err := m.GetOption(rpc)

		if err != nil {

			m.Log("Error", err)
			os.Exit(0)

		}

		if flag == false {

			continue

		}

		if option.GetMutation() == "" {

			continue

		}

		if option.GetQuery() != "" {

			PossibleReqObjects[rpc.Input().Name().String()] = true

		}
	}

	return nil
}

func (m *jaalModule) InitFunc(initFunctionsName map[string]bool) (string, error) {
	//returns template of init function

	tmp := getInitTemplate()
	buf := &bytes.Buffer{}

	if err := tmp.Execute(buf, initFunctionsName); err != nil {

		return "", err

	}

	return buf.String(), nil
}

func (m *jaalModule) GetGoPackage(target pgs.File) string {
	//returns go package for a file

	goPackage := "pb"

	if target.Descriptor().GetOptions() != nil && target.Descriptor().GetOptions().GoPackage != nil {
		goPackage = *target.Descriptor().GetOptions().GoPackage
		goPackage = strings.Split(goPackage, ";")[0]
		goPackage = strings.Split(goPackage, "/")[len(strings.Split(goPackage, "/"))-1]

	}

	return goPackage
}
func (m *jaalModule) GetImports(target pgs.File) map[string]string {
	// returns a map of all imports

	imports := make(map[string]string)

	for _, importFile := range target.Imports() {

		if importFile.Descriptor().Options != nil && importFile.Descriptor().Options.GoPackage != nil {

			key := *importFile.Descriptor().Options.GoPackage
			key = strings.Split(key, ";")[0]
			imports[key] = strings.Split(key, "/")[len(strings.Split(key, "/"))-1]
		}
	}

	return imports
}

func (m *jaalModule) ServiceStructInputFunc(service pgs.Service, initFunctionsName map[string]bool) (string, error) {
	//returns template of service input struct registered methods for a service

	var inputServiceStructFunc []InputClass

	for _, rpc := range service.Methods() {
		var maps []InputMap
		flag, option, err := m.GetOption(rpc)

		if err != nil {

			m.Log("Error", err)
			os.Exit(0)

		}

		if flag == false {

			continue

		}

		if option.GetMutation() == "" {

			continue

		}

		var field []MsgFields

		for _, ipField := range rpc.Input().Fields() {

			tname := ipField.Name().UpperCamelCase().String()

			//if ipField.Package().ProtoName() != service.Package().ProtoName() {
			//
			//	tname = m.GetGoPackage(ipField.File()) + "." + ipField.Name().UpperCamelCase().String()
			//
			//}

			fName := ipField.Name().LowerCamelCase().String()
			tval := ""
			funcPara := ""
			if ipField.InOneOf() {
				goPkg := m.GetGoPackageOfFiles(service.File(), ipField.OneOf().File())
				if goPkg != "" {
					goPkg += "."
				}
				funcPara = goPkg + rpc.Input().Name().UpperCamelCase().String() + "_" + ipField.Name().UpperCamelCase().String()
				funcPara = "*" + funcPara
				field = append(field, MsgFields{TargetName: tname, FieldName: fName, FuncPara: funcPara, TargetVal: "source"})
				continue
			} else if ipField.Type().IsRepeated() {
				funcPara = "[]"
				tObj := ipField.Type().Element()

				if tObj.IsEmbed() {

					funcPara += "*"

				}

				if tObj.IsEmbed() && tObj.Embed().File().Descriptor().Options != nil && tObj.Embed().File().Descriptor().Options.GoPackage != nil {

					if service.Package().ProtoName().String() != tObj.Embed().Package().ProtoName().String() {

						funcPara += m.GetGoPackage(tObj.Embed().File())
						funcPara += "."

					}else{
						if strings.Split(tObj.Embed().FullyQualifiedName(), ".")[len(strings.Split(tObj.Embed().FullyQualifiedName(), "."))-2] == tObj.Embed().Parent().Name().String(){
							funcPara+= (tObj.Embed().Parent().Name().String()+"_")
						}
					}

				}
				tval = "source"
				funcPara += m.fieldElementType(tObj)
			} else if ipField.Type().IsMap() {
				// TODO : Repeated case not handled
				goPkg := ""
				if ipField.Type().Element().IsEmbed() {

					goPkg = m.GetGoPackageOfFiles(service.File(), ipField.Type().Element().Embed().File())
					if goPkg != "" {
						goPkg += "."
					}
				}

				asterik := ""
				if ipField.Type().Element().IsEmbed() {
					asterik = "*"
				}

				value := asterik + goPkg + m.fieldElementType(ipField.Type().Element())
				maps = append(maps, InputMap{FieldName: fName, TargetVal: "*source", TargetName: ipField.Name().UpperCamelCase().String(), Key: m.fieldElementType(ipField.Type().Key()), Value: value})
				continue
			} else {
				if ipField.Descriptor().GetType().String() == "TYPE_MESSAGE" {

					tval = "source"

				} else {

					tval = "*source"

				}

				funcPara = m.RPCFieldType(ipField)

				if funcPara[0] == '*' {

					funcPara = funcPara[1:len(funcPara)]

				}

				goPkg := ""
				if ipField.Type().IsEmbed() {

					if ipField.Type().Embed().File().Descriptor().Options != nil && ipField.Type().Embed().File().Descriptor().Options.GoPackage != nil && service.Package().ProtoName().String() != ipField.Type().Embed().Package().ProtoName().String() {

						goPkg = m.GetGoPackage(ipField.Type().Embed().File()) + "."
					} else {
						// message is embedded inside a message then it's gopkg is it's parent message
						if strings.Split(ipField.FullyQualifiedName(), ".")[len(strings.Split(ipField.FullyQualifiedName(), "."))-2] == ipField.Type().Embed().Parent().Name().String() {
							goPkg = ipField.Type().Embed().Parent().Name().String() + "_"
						}
					}
				} else if ipField.Type().IsEnum() {
					goPkg = m.GetGoPackageOfFiles(service.File(), ipField.Type().Enum().File())
					if goPkg != "" {
						goPkg += "."
					}
				}
				funcPara = "*" + goPkg + funcPara
			}
			if strings.HasSuffix(funcPara, "*timestamp.Timestamp") {
				funcPara = funcPara[:len(funcPara)-19] + "schemabuilder.Timestamp"
				tval = "(*timestamp.Timestamp)(source)"
			} else if strings.HasSuffix(funcPara, "*field_mask.FieldMask") {
				tval = "gtypes.ModifyFieldMask(source)"
			}

			field = append(field, MsgFields{TargetName: tname, FieldName: fName, FuncPara: funcPara, TargetVal: tval})
		}

		initFunctionsName["RegisterInput"+rpc.Name().UpperCamelCase().String()+"Input"] = true
		inputServiceStructFunc = append(inputServiceStructFunc, InputClass{Name: rpc.Name().UpperCamelCase().String(), Fields: field, Maps: maps})
	}

	tmp := getServiceStructInputFuncTemplate()
	buf := &bytes.Buffer{}

	if err := tmp.Execute(buf, inputServiceStructFunc); err != nil {

		return "", err

	}

	return buf.String(), nil
}

func (m *jaalModule) ServiceStructPayloadFunc(service pgs.Service, initFunctionsName map[string]bool) (string, error) {
	//returns template of service payload struct registered methods for a service

	var payloadService []PayloadServiceStruct

	for _, rpc := range service.Methods() {

		flag, option, err := m.GetOption(rpc)

		if err != nil {

			m.Log("Error", err)
			os.Exit(0)

		}

		if flag == false {

			continue

		}

		if option.GetMutation() == "" {

			continue

		}

		initFunctionsName["RegisterPayload"+rpc.Name().UpperCamelCase().String()+"Payload"] = true
		goPkg := m.GetGoPackageOfFiles(service.File(), rpc.Output().File())
		if goPkg != "" {
			goPkg += "."
		}
		returnType := "*" + goPkg + rpc.Output().Name().UpperCamelCase().String() // "*" + rpc.Output().Name().UpperCamelCase().String()
		payloadService = append(payloadService, PayloadServiceStruct{Name: rpc.Name().UpperCamelCase().String(), ReturnType: returnType})
	}

	tmp := getServiceStructPayloadFuncTemplate()
	buf := &bytes.Buffer{}

	if err := tmp.Execute(buf, payloadService); err != nil {

		return "", err

	}

	return buf.String(), nil
}
