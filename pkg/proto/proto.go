// Copyright 2023 Yusuke Fredrick Tsutsumi //
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package proto

import (
	"bytes"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/aep-dev/aep-lib-go/pkg/cases"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/builder"
	"github.com/jhump/protoreflect/desc/protoprint"
	"google.golang.org/protobuf/types/descriptorpb"
)

type MessageStorage struct {
	Messages map[string]Message
}

func APIToProtoString(a *api.API, outputDir string) ([]byte, error) {
	fd, err := APIToProto(a, outputDir)
	if err != nil {
		return []byte{}, err
	}
	printer := protoprint.Printer{
		CustomSortFunction: compareProtoElements,
	}
	var output bytes.Buffer
	err = printer.PrintProtoFile(fd, &output)
	if err != nil {
		return []byte{}, err
	}
	return output.Bytes(), nil
}

func APIToProto(a *api.API, outputDir string) (*desc.FileDescriptor, error) {
	m := &MessageStorage{Messages: map[string]Message{}}
	dir, file := filepath.Split(outputDir)
	packageParts := []string{file}
	for dir != "." {
		dir = filepath.Clean(dir)
		dir, file = filepath.Split(dir)
		dir = filepath.Clean(dir)
		packageParts = append(packageParts, file)
	}
	slices.Reverse(packageParts)
	packageName := strings.Join(packageParts, ".")
	println(packageName)

	fb := builder.NewFile("test.proto")
	fb.Package = packageName
	fb.IsProto3 = true
	// As a file comment is not printed by protoprint,
	// use package comments instead.
	fb.SetPackageComments(builder.Comments{
		LeadingComment: "this file is generated.",
	})
	pServiceName := toProtoServiceName(a.Name)
	serviceNameAsLower := fmt.Sprintf("/%s", strings.ToLower(pServiceName))
	fo := &descriptorpb.FileOptions{
		GoPackage: &serviceNameAsLower,
	}
	fb.SetOptions(fo)
	sb := builder.NewService(pServiceName)
	sb.SetComments(builder.Comments{
		LeadingComment: "A service.",
	})

	// Add resources to MessageStorage.
	err := GenerateSchemaMessages(a, m, fb)
	if err != nil {
		return nil, err
	}

	for _, r := range getSortedResources(a) {
		err := AddResource(r, a, fb, sb, m)
		if err != nil {
			return nil, fmt.Errorf("adding resource %v failed: %w", r.Singular, err)
		}
	}
	fb.AddService(sb)
	fd, err := fb.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to build service file %v: %w", fb.GetName(), err)
	}

	// protoreflect sometimes adds "import {generated-file-0001}.proto" unnecessarily.
	d := []string{}
	for _, v := range fd.AsFileDescriptorProto().Dependency {
		if !strings.Contains(v, "generated-file") {
			d = append(d, v)
		}
	}
	fd.AsFileDescriptorProto().Dependency = d

	return fd, nil
}

func GenerateSchemaMessages(a *api.API, m *MessageStorage, fb *builder.FileBuilder) error {
	// Generate Resource messages on combined map
	schemaByName := make(map[string]*openapi.Schema)
	for name, s := range a.Schemas {
		schemaByName[name] = s
	}
	for name, r := range a.Resources {
		schemaByName[name] = r.Schema
	}
	schemaNames := []string{}
	for name := range schemaByName {
		schemaNames = append(schemaNames, name)
	}
	sort.Strings(schemaNames)
	for _, name := range schemaNames {
		s := schemaByName[name]
		mb, err := GenerateSchemaMessage(name, s, a, m)
		if err != nil {
			return err
		}
		mb.AddMessage(fb)
	}
	return nil
}

func toProtoServiceName(serviceName string) string {
	parts := strings.SplitN(serviceName, ".", 2)
	return cases.Capitalize(parts[0])
}

func toMessageName(resource string) string {
	return cases.KebabToCamelCase(resource)
}

func getSortedResources(a *api.API) []*api.Resource {
	keys := []string{}
	for k := range a.Resources {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	resources := make([]*api.Resource, 0, len(keys))
	for _, k := range keys {
		resources = append(resources, a.Resources[k])
	}
	return resources
}

// compareProtoElements compares two protoprint.Element instances
// and returns true if the first element should come before the second element.
// customize to adhere to the AEPs.
func compareProtoElements(a, b protoprint.Element) bool {
	return protoPrintKindToElement(a.Kind()) < protoPrintKindToElement(b.Kind())
}

func protoPrintKindToElement(ek protoprint.ElementKind) int {
	switch ek {
	case protoprint.KindPackage:
		return 0
	case protoprint.KindImport:
		return 1
	case protoprint.KindOption:
		return 2
	case protoprint.KindService:
		return 4
	case protoprint.KindEnum:
		return 5
	case protoprint.KindMessage:
		return 6
	case protoprint.KindField:
		return 7
	case protoprint.KindExtensionRange:
		return 8
	case protoprint.KindExtension:
		return 9
	case protoprint.KindReservedRange:
		return 10
	case protoprint.KindReservedName:
		return 11
	case protoprint.KindEnumValue:
		return 12
	case protoprint.KindMethod:
		return 13
	default:
		return 99
	}
}
