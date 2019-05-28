package main

import (
	pgs "github.com/lyft/protoc-gen-star"
	pgsgo "github.com/lyft/protoc-gen-star/lang/go"
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

func (m *jaalModule) Execute(targets map[string]pgs.File, pkgs map[string]pgs.Package) []pgs.Artifact {
	for _, target := range targets { // loop over files
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
