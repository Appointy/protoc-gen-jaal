package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	pgs "github.com/lyft/protoc-gen-star"
	pgsgo "github.com/lyft/protoc-gen-star/lang/go"
	pbt "go.appointy.com/protoc-gen-jaal/schema"
)

type jaalModule struct {
	*pgs.ModuleBase
	pgsgo.Context
}

func (m *jaalModule) InitContext(c pgs.BuildContext) {
	m.ModuleBase.InitContext(c)
	m.Context = pgsgo.InitContext(c.Parameters())
}

func (m *jaalModule) Name() string { return "graphql" }
func (m *jaalModule) CheckSkipFile(target pgs.File) (bool, error) {
	// checks file_skip option
	opt := target.Descriptor().GetOptions()
	if opt != nil {
		x, err := proto.GetExtension(opt, pbt.E_FileSkip)
		if err != nil && proto.ErrMissingExtension != err {
			fmt.Println("Error", err)
			return false, err
		}

		if x != nil && *x.(*bool) == true { // skips only when file_skip is explicitly true
			return true, nil
		}
	}
	return false, nil
}
func (m *jaalModule) Execute(targets map[string]pgs.File, pkgs map[string]pgs.Package) []pgs.Artifact {
	for _, target := range targets { // loop over files

		if ok, err := m.CheckSkipFile(target); err != nil { // checks file_skip option
			fmt.Println("Error :", err)
			continue
		} else if ok == true {
			continue
		}

		fname := target.Name().String()
		fname = fname[:len(fname)-6]

		name := m.BuildContext.OutputPath() + "/" + fname + ".pb.gq.go"

		str, err := m.generateFileData(target)
		if err != nil {
			m.Log("Error : ", err)
		}
		m.AddGeneratorFile(name, str)
	}
	return m.Artifacts()
}
