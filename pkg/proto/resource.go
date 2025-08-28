// Copyright 2023 Yusuke Fredrick Tsutsumi
//
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
	"fmt"
	"reflect"
	"slices"
	"strings"

	apipb "buf.build/gen/go/aep/api/protocolbuffers/go/aep/api"
	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/aep-dev/aep-lib-go/pkg/cases"
	"github.com/aep-dev/aep-lib-go/pkg/constants"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/builder"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var operationMd *desc.MessageDescriptor

func init() {
	var err error
	operationMd, err = desc.LoadMessageDescriptor("aep.api.Operation")
	if err != nil {
		panic(fmt.Sprintf("error initializing resource.go: %v", err))
	}
}

// AddResource adds a resource's protos and RPCs to a file and service.
func AddResource(r *api.Resource, a *api.API, fb *builder.FileBuilder, sb *builder.ServiceBuilder, ms *MessageStorage) error {
	// Do not recreate resources if they've already been created.
	resMsg, ok := ms.Messages[fmt.Sprintf("%s/%s", a.Name, r.Singular)]
	if options := resMsg.Options(); options != nil {
		proto.SetExtension(options, annotations.E_Resource, resourceDescriptor(a, r))
	}
	if !ok {
		return fmt.Errorf("%s not found in message storage", r.Singular)
	}

	if r.Methods.Create != nil {
		err := AddCreate(a, r, resMsg, fb, sb)
		if err != nil {
			return err
		}
	}
	if r.Methods.Get != nil {
		err := AddGet(a, r, resMsg, fb, sb)
		if err != nil {
			return err
		}
	}
	if r.Methods.Update != nil {
		err := AddUpdate(a, r, resMsg, fb, sb)
		if err != nil {
			return err
		}
	}
	if r.Methods.Delete != nil {
		err := AddDelete(a, r, resMsg, fb, sb)
		if err != nil {
			return err
		}
	}
	if r.Methods.List != nil {
		err := AddList(r, resMsg, fb, sb)
		if err != nil {
			return err
		}
	}
	// TODO: add global list
	/// if r.GlobalList != nil {
	/// 	err := AddGlobalList(r, resourceMb, fb, sb)
	/// 	if err != nil {
	/// 		return err
	/// 	}
	/// }

	if r.Methods.Apply != nil {
		err := AddApply(a, r, resMsg, fb, sb)
		if err != nil {
			return err
		}
	}

	for _, cm := range r.CustomMethods {
		err := AddCustomMethod(a, r, cm, resMsg, fb, ms, sb)
		if err != nil {
			return err
		}
	}
	return nil
}

// this function should only be called with openapi Schemas that
// map to primitive types.
func protoFieldType(name string, number int, s openapi.Schema, a *api.API, m *MessageStorage, parent *builder.MessageBuilder) (*builder.FieldType, error) {
	switch s.Type {
	case "object":
		typ, err := protoFieldTypeObject(name, &s, a, m, parent)
		if err != nil {
			return nil, fmt.Errorf("error creating proto type object for %s: %w", name, err)
		}
		return typ, nil
	// Ideally we would set the repeated parameter here as well.
	// But "repeated" is a property of the field, not the type.
	case "array":
		typ, err := protoFieldType(name, number, *s.Items, a, m, parent)
		if err != nil {
			return nil, fmt.Errorf("error creating proto type for array item for %s: %w", name, err)
		}
		return typ, nil
	case "string":
		return builder.FieldTypeString(), nil
	case "boolean":
		return builder.FieldTypeBool(), nil
	case "integer":
		if s.Format == "int32" {
			return builder.FieldTypeInt32(), nil
		} else if s.Format == "int64" {
			return builder.FieldTypeInt64(), nil
		}
	case "number":
		if s.Format == "float" {
			return builder.FieldTypeFloat(), nil
		} else if s.Format == "double" {
			return builder.FieldTypeDouble(), nil
		}
	}
	return nil, fmt.Errorf("proto type for %q, format %q not found", s.Type, s.Format)
}

func GenerateMessage(name string, s *openapi.Schema, a *api.API, m *MessageStorage) (*builder.MessageBuilder, error) {
	mb := builder.NewMessage(name)
	options := &descriptorpb.MessageOptions{}
	mb.SetOptions(options)
	mb.SetComments(builder.Comments{
		LeadingComment: fmt.Sprintf("A %v.", name),
	})
	required := map[string]bool{}
	for _, n := range s.Required {
		required[n] = true
	}
	sorted_field_numbers := []int{}
	field_names_by_number := map[int]string{}
	for k, s := range s.Properties {
		if s.XAEPFieldNumber != 0 {
			field_names_by_number[s.XAEPFieldNumber] = k
			sorted_field_numbers = append(sorted_field_numbers, s.XAEPFieldNumber)
		}
	}
	slices.Sort(sorted_field_numbers)

	for _, num := range sorted_field_numbers {
		name := field_names_by_number[num]
		f, err := protoField(name, num, s.Properties[name], a, m, mb)
		if err != nil {
			return nil, fmt.Errorf("error generating message '%s': %v", name, err)
		}
		if required[name] {
			o := &descriptorpb.FieldOptions{}
			proto.SetExtension(o, annotations.E_FieldBehavior, []annotations.FieldBehavior{annotations.FieldBehavior_REQUIRED})
			f.SetOptions(o)
		}
		mb.AddField(f)
	}
	return mb, nil
}

func protoField(name string, number int, s openapi.Schema, a *api.API, m *MessageStorage, parent *builder.MessageBuilder) (*builder.FieldBuilder, error) {
	typ, err := protoFieldType(name, number, s, a, m, parent)
	if err != nil {
		return nil, fmt.Errorf("error creating proto field for '%s', with number %d: %w", name, number, err)
	}
	f := builder.NewField(name, typ).SetNumber(int32(number)).SetComments(
		builder.Comments{
			LeadingComment: fmt.Sprintf("Field for %v.", name),
		},
	)
	if s.Type == "array" {
		f.SetRepeated()
	}

	return f, nil
}

func protoFieldTypeObject(name string, s *openapi.Schema, a *api.API, m *MessageStorage, parent *builder.MessageBuilder) (*builder.FieldType, error) {
	if s.Ref != "" {
		wantedType := s.Ref
		// extract the name from the ref
		wantedType = strings.TrimPrefix(wantedType, "#/components/schemas/")
		wantedType = fmt.Sprintf("%s/%s", a.Name, wantedType)
		_, ok := m.Messages[wantedType]
		if !ok {
			return nil, fmt.Errorf("could not find message %s, referenced by %s", wantedType, name)
		}
		return m.Messages[wantedType].FieldType(), nil
	} else {
		msg, err := GenerateMessage(toMessageName(name), s, a, m)
		if err != nil {
			return nil, err
		}
		parent.AddNestedMessage(msg)
		return builder.FieldTypeMessage(msg), nil
	}
}

// GenerateResourceMesssage adds the resource message.
func GenerateSchemaMessage(name string, s *openapi.Schema, a *api.API, ms *MessageStorage) (Message, error) {
	var m Message
	if s.XAEPProtoMessageName != "" {
		md, err := desc.LoadMessageDescriptor(s.XAEPProtoMessageName)
		if err != nil {
			return nil, fmt.Errorf("error loading message descriptor for %s: %w", s.XAEPProtoMessageName, err)
		}
		m = NewWrappedMessageDescriptor(md)
	} else {
		mb, err := GenerateMessage(toMessageName(name), s, a, ms)
		if err != nil {
			return nil, err
		}
		m = NewWrappedMessageBuilder(mb)
	}
	ms.Messages[fmt.Sprintf("%s/%s", a.Name, name)] = m
	return m, nil
}

func AddCreate(a *api.API, r *api.Resource, resMsg Message, fb *builder.FileBuilder, sb *builder.ServiceBuilder) error {
	// add the resource message
	// create request messages
	mb := builder.NewMessage("Create" + toMessageName(r.Singular) + "Request")
	mb.SetComments(builder.Comments{
		LeadingComment: fmt.Sprintf("A Create request for a  %v resource.", r.Singular),
	})
	addParentField(r, mb)
	if r.Methods.Create.SupportsUserSettableCreate {
		addIdField(r, mb)
	}
	addResourceField(r, resMsg, mb)
	fb.AddMessage(mb)
	method := buildMethod(
		"Create"+toMessageName(r.Singular),
		builder.RpcTypeMessage(mb, false),
		resMsg.RpcType(),
		r.Methods.Create.IsLongRunning,
	)
	method.SetComments(builder.Comments{
		LeadingComment: fmt.Sprintf("An aep-compliant Create method for %v.", r.Singular),
	})
	bodyField := cases.KebabToSnakeCase(r.Singular)
	proto.SetExtension(method.Options, annotations.E_Http, &annotations.HttpRule{
		Pattern: &annotations.HttpRule_Post{
			Post: generateParentHTTPPath(r),
		},
		Body: bodyField,
	})
	method_signature := []string{bodyField}
	if len(r.Parents) > 0 {
		method_signature = []string{constants.FIELD_PARENT_NAME, bodyField}
	}
	proto.SetExtension(method.Options, annotations.E_MethodSignature, []string{strings.Join(method_signature, ",")})
	sb.AddMethod(method)
	return nil
}

// AddGet adds a read method for the resource, along with
// any required messages.
func AddGet(a *api.API, r *api.Resource, resMsg Message, fb *builder.FileBuilder, sb *builder.ServiceBuilder) error {
	mb := builder.NewMessage("Get" + toMessageName(r.Singular) + "Request")
	mb.SetComments(builder.Comments{
		LeadingComment: fmt.Sprintf("Request message for the Get%v method", r.Singular),
	})
	addPathField(a, r, mb)
	fb.AddMessage(mb)
	method := builder.NewMethod("Get"+toMessageName(r.Singular),
		builder.RpcTypeMessage(mb, false),
		resMsg.RpcType(),
	)
	method.SetComments(builder.Comments{
		LeadingComment: fmt.Sprintf("An aep-compliant Get method for %v.", r.Singular),
	})
	options := &descriptorpb.MethodOptions{}
	proto.SetExtension(options, annotations.E_Http, &annotations.HttpRule{
		Pattern: &annotations.HttpRule_Get{
			Get: fmt.Sprintf("/{path=%v}", generateHTTPPath(r)),
		},
	})
	proto.SetExtension(options, annotations.E_MethodSignature, []string{
		strings.Join([]string{constants.FIELD_PATH_NAME}, ","),
	})
	method.SetOptions(options)
	sb.AddMethod(method)
	return nil
}

// AddRead adds a read method for the resource, along with
// any required messages.
func AddUpdate(a *api.API, r *api.Resource, resMsg Message, fb *builder.FileBuilder, sb *builder.ServiceBuilder) error {
	mb := builder.NewMessage("Update" + toMessageName(r.Singular) + "Request")
	mb.SetComments(builder.Comments{
		LeadingComment: fmt.Sprintf("Request message for the Update%v method", toMessageName(r.Singular)),
	})
	addPathField(a, r, mb)
	addResourceField(r, resMsg, mb)
	// TODO: find a way to get the actual field mask proto descriptor type, without
	// querying the global registry.
	fieldMaskDescriptor, _ := desc.LoadMessageDescriptorForType(reflect.TypeOf(fieldmaskpb.FieldMask{}))
	updateMaskField := builder.NewField(constants.FIELD_UPDATE_MASK_NAME, builder.FieldTypeImportedMessage(fieldMaskDescriptor)).
		SetNumber(constants.FIELD_UPDATE_MASK_NUMBER).
		SetComments(builder.Comments{
			LeadingComment: "The update mask for the resource",
		})
	mb.AddField(updateMaskField)
	fb.AddMessage(mb)
	method := buildMethod(
		"Update"+toMessageName(r.Singular),
		builder.RpcTypeMessage(mb, false),
		resMsg.RpcType(),
		r.Methods.Update.IsLongRunning,
	)
	method.SetComments(builder.Comments{
		LeadingComment: fmt.Sprintf("An aep-compliant Update method for %v.", r.Singular),
	})
	body_field := cases.KebabToSnakeCase(r.Singular)
	proto.SetExtension(method.Options, annotations.E_Http, &annotations.HttpRule{
		Pattern: &annotations.HttpRule_Patch{
			Patch: fmt.Sprintf("/{path=%v}", generateHTTPPath(r)),
		},
		Body: body_field,
	})
	proto.SetExtension(method.Options, annotations.E_MethodSignature, []string{
		strings.Join([]string{body_field, constants.FIELD_UPDATE_MASK_NAME}, ","),
	})
	sb.AddMethod(method)
	return nil
}

func AddDelete(a *api.API, r *api.Resource, resMsg Message, fb *builder.FileBuilder, sb *builder.ServiceBuilder) error {
	// add the resource message
	// create request messages
	mb := builder.NewMessage("Delete" + toMessageName(r.Singular) + "Request")
	mb.SetComments(builder.Comments{
		LeadingComment: fmt.Sprintf("Request message for the Delete%v method", toMessageName(r.Singular)),
	})
	addPathField(a, r, mb)
	if len(r.Children) > 0 {
		addForceField(a, r, mb)
	}
	fb.AddMessage(mb)
	emptyMd, err := desc.LoadMessageDescriptor("google.protobuf.Empty")
	if err != nil {
		return err
	}
	method := buildMethod(
		"Delete"+toMessageName(r.Singular),
		builder.RpcTypeMessage(mb, false),
		builder.RpcTypeImportedMessage(emptyMd, false),
		r.Methods.Delete.IsLongRunning,
	)
	method.SetComments(builder.Comments{
		LeadingComment: fmt.Sprintf("An aep-compliant Delete method for %v.", r.Singular),
	})
	proto.SetExtension(method.Options, annotations.E_Http, &annotations.HttpRule{
		Pattern: &annotations.HttpRule_Delete{
			Delete: fmt.Sprintf("/{path=%v}", generateHTTPPath(r)),
		},
	})
	proto.SetExtension(method.Options, annotations.E_MethodSignature, []string{
		strings.Join([]string{constants.FIELD_PATH_NAME}, ","),
	})
	sb.AddMethod(method)
	return nil
}

func AddList(r *api.Resource, resMsg Message, fb *builder.FileBuilder, sb *builder.ServiceBuilder) error {
	// add the resource message
	// create request messages
	reqMb := builder.NewMessage("List" + toMessageName(r.Plural) + "Request")
	reqMb.SetComments(builder.Comments{
		LeadingComment: fmt.Sprintf("Request message for the List%v method", r.Singular),
	})
	addParentField(r, reqMb)
	addPageToken(r, reqMb)
	maxPageSizeField := builder.NewField(constants.FIELD_MAX_PAGE_SIZE_NAME, builder.FieldTypeInt32()).
		SetNumber(constants.FIELD_MAX_PAGE_SIZE_NUMBER).
		SetComments(builder.Comments{
			LeadingComment: "The maximum number of resources to return in a single page.",
		})
	reqMb.AddField(maxPageSizeField)
	if r.Methods.List.SupportsSkip {
		skipField := builder.NewField(constants.FIELD_SKIP_NAME, builder.FieldTypeInt32()).
			SetNumber(constants.FIELD_SKIP_NUMBER).
			SetComments(builder.Comments{
				LeadingComment: "The number of resources to skip before returning the first resource in the page.",
			})
		reqMb.AddField(skipField)
	}
	if r.Methods.List.SupportsFilter {
		filterField := builder.NewField(constants.FIELD_FILTER_NAME, builder.FieldTypeString()).
			SetNumber(constants.FIELD_FILTER_NUMBER).
			SetComments(builder.Comments{
				LeadingComment: "The filter to apply to the list.",
			})
		reqMb.AddField(filterField)
	}
	fb.AddMessage(reqMb)
	respMb := builder.NewMessage("List" + toMessageName(r.Plural) + "Response")
	respMb.SetComments(builder.Comments{
		LeadingComment: fmt.Sprintf("Response message for the List%v method", r.Singular),
	})
	addResourcesField(r, resMsg, respMb)
	addNextPageToken(r, respMb)
	if r.Methods.List.HasUnreachableResources {
		f := builder.NewField(constants.FIELD_UNREACHABLE_NAME, resMsg.FieldType()).SetNumber(constants.FIELD_UNREACHABLE_NUMBER).SetComments(builder.Comments{
			LeadingComment: fmt.Sprintf("A list of %v that were not reachable.", r.Plural),
		}).SetRepeated()
		respMb.AddField(f)
	}
	fb.AddMessage(respMb)
	method := builder.NewMethod("List"+toMessageName(r.Plural),
		builder.RpcTypeMessage(reqMb, false),
		builder.RpcTypeMessage(respMb, false),
	)
	method.SetComments(builder.Comments{
		LeadingComment: fmt.Sprintf("An aep-compliant List method for %v.", r.Plural),
	})
	options := &descriptorpb.MethodOptions{}
	proto.SetExtension(options, annotations.E_Http, &annotations.HttpRule{
		Pattern: &annotations.HttpRule_Get{
			Get: generateParentHTTPPath(r),
		},
	})
	proto.SetExtension(options, annotations.E_MethodSignature, []string{
		strings.Join([]string{constants.FIELD_PARENT_NAME}, ","),
	})
	method.SetOptions(options)
	sb.AddMethod(method)
	return nil
}

func AddGlobalList(r *api.Resource, a *api.API, resMsg Message, fb *builder.FileBuilder, sb *builder.ServiceBuilder) error {
	// add the resource message
	// create request messages
	reqMb := builder.NewMessage("GlobalList" + r.Singular + "Request")
	reqMb.SetComments(builder.Comments{
		LeadingComment: fmt.Sprintf("Request message for the GlobalList%v method", r.Singular),
	})
	addPathField(a, r, reqMb)
	addPageToken(r, reqMb)
	fb.AddMessage(reqMb)
	respMb := builder.NewMessage("GlobalList" + r.Singular + "Response")
	respMb.SetComments(builder.Comments{
		LeadingComment: fmt.Sprintf("Response message for the GlobalList%v method", r.Singular),
	})
	addResourcesField(r, resMsg, respMb)
	addNextPageToken(r, respMb)
	fb.AddMessage(respMb)
	method := builder.NewMethod("GlobalList"+r.Singular,
		builder.RpcTypeMessage(reqMb, false),
		builder.RpcTypeMessage(respMb, false),
	)
	options := &descriptorpb.MethodOptions{}
	proto.SetExtension(options, annotations.E_Http, &annotations.HttpRule{
		Pattern: &annotations.HttpRule_Get{
			Get: fmt.Sprintf("/{path=--/%v}", strings.ToLower(r.Singular)),
		},
	})
	method.SetOptions(options)
	sb.AddMethod(method)
	return nil
}

// AddApply adds a read method for the resource, along with
// any required messages.
func AddApply(a *api.API, r *api.Resource, resMsg Message, fb *builder.FileBuilder, sb *builder.ServiceBuilder) error {
	mb := builder.NewMessage("Apply" + toMessageName(r.Singular) + "Request")
	mb.SetComments(builder.Comments{
		LeadingComment: fmt.Sprintf("Request message for the Apply%v method", r.Singular),
	})
	addPathField(a, r, mb)
	addResourceField(r, resMsg, mb)
	fb.AddMessage(mb)
	method := buildMethod(
		"Apply"+toMessageName(r.Singular),
		builder.RpcTypeMessage(mb, false),
		resMsg.RpcType(),
		r.Methods.Apply.IsLongRunning,
	)
	method.SetComments(builder.Comments{
		LeadingComment: fmt.Sprintf("An aep-compliant Apply method for %v.", r.Plural),
	})
	proto.SetExtension(method.Options, annotations.E_Http, &annotations.HttpRule{
		Pattern: &annotations.HttpRule_Put{
			Put: fmt.Sprintf("/{path=%v}", generateHTTPPath(r)),
		},
		// TODO: do a conversion to underscores instead.
		Body: strings.ToLower(r.Singular),
	})
	sb.AddMethod(method)
	return nil
}

func AddCustomMethod(a *api.API, r *api.Resource, cm *api.CustomMethod, resMsg Message, fb *builder.FileBuilder, m *MessageStorage, sb *builder.ServiceBuilder) error {
	methodName := cases.KebabToCamelCase(cm.Name) + toMessageName(r.Singular)
	request := cm.Request
	if request == nil {
		request = &openapi.Schema{}
	}
	requestMb, err := GenerateMessage(methodName+"Request", request, a, m)
	if err != nil {
		return err
	}
	requestMb.SetComments(builder.Comments{
		LeadingComment: fmt.Sprintf("Request message for the %v method", cm.Name),
	})
	addPathField(a, r, requestMb)
	var responseMsg *builder.RpcType
	if cm.Response != nil {
		responseMb, err := GenerateMessage(methodName+"Response", cm.Response, a, m)
		if err != nil {
			return err
		}
		responseMb.SetComments(builder.Comments{
			LeadingComment: fmt.Sprintf("Response message for the %v method", cm.Name),
		})
		fb.AddMessage(responseMb)
		responseMsg = builder.RpcTypeMessage(responseMb, false)
	} else {
		emptyMsg, err := desc.LoadMessageDescriptor("google.protobuf.Empty")
		if err != nil {
			return fmt.Errorf("error loading message descriptor for google.protobuf.Empty: %v", err)
		}
		responseMsg = builder.RpcTypeImportedMessage(emptyMsg, false)
	}
	fb.AddMessage(requestMb)
	method := buildMethod(
		methodName,
		builder.RpcTypeMessage(requestMb, false),
		responseMsg,
		cm.IsLongRunning,
	)
	method.SetComments(builder.Comments{
		LeadingComment: fmt.Sprintf("%v a %v.", cm.Name, r.Singular),
	})
	http_path := fmt.Sprintf("/{path=%v}:%v", generateHTTPPath(r), cm.Name)
	switch cm.Method {
	case "POST":
		proto.SetExtension(method.Options, annotations.E_Http, &annotations.HttpRule{
			Pattern: &annotations.HttpRule_Post{
				Post: http_path,
			},
			Body: "*",
		})
	case "GET":
		proto.SetExtension(method.Options, annotations.E_Http, &annotations.HttpRule{
			Pattern: &annotations.HttpRule_Get{
				Get: http_path,
			},
		})
	}
	sb.AddMethod(method)
	return nil
}

func generateHTTPPath(r *api.Resource) string {
	elements := []string{api.CollectionName(r)}
	if len(r.Parents) > 0 {
		// TODO: handle multiple parents
		p := r.ParentResources()[0]
		for p != nil {
			elements = append([]string{api.CollectionName(p)}, elements...)
			if len(p.ParentResources()) == 0 {
				break
			}
			p = p.ParentResources()[0]
		}
	}
	return fmt.Sprintf("%v/*", strings.Join(elements, "/*/"))
}

func generateParentHTTPPath(r *api.Resource) string {
	parentPath := ""
	if len(r.Parents) == 0 {
		return fmt.Sprintf("/%v", strings.ToLower(r.Plural))
	}
	if len(r.Parents) > 0 {
		parentPath = fmt.Sprintf("%v", generateHTTPPath(r.ParentResources()[0]))
	}
	return fmt.Sprintf("/{parent=%v}/%v", parentPath, api.CollectionName(r))
}

func addParentField(r *api.Resource, mb *builder.MessageBuilder) {
	o := &descriptorpb.FieldOptions{}
	proto.SetExtension(o, annotations.E_FieldBehavior, []annotations.FieldBehavior{annotations.FieldBehavior_REQUIRED})
	proto.SetExtension(o, annotations.E_ResourceReference, &annotations.ResourceReference{})
	f := builder.
		NewField(constants.FIELD_PARENT_NAME, builder.FieldTypeString()).
		SetNumber(constants.FIELD_PARENT_NUMBER).
		SetComments(builder.Comments{
			LeadingComment: fmt.Sprintf("A field for the parent of %v", r.Singular),
		}).
		SetOptions(o)
	mb.AddField(f)
}

func addIdField(_ *api.Resource, mb *builder.MessageBuilder) {
	f := builder.NewField(constants.FIELD_ID_NAME, builder.FieldTypeString()).SetNumber(constants.FIELD_ID_NUMBER).SetComments(builder.Comments{
		LeadingComment: "An id that uniquely identifies the resource within the collection",
	})
	mb.AddField(f)
}

func addPathField(a *api.API, r *api.Resource, mb *builder.MessageBuilder) {
	o := &descriptorpb.FieldOptions{}
	proto.SetExtension(o, annotations.E_FieldBehavior, []annotations.FieldBehavior{annotations.FieldBehavior_REQUIRED})
	proto.SetExtension(o, annotations.E_ResourceReference, &annotations.ResourceReference{
		Type: fmt.Sprintf("%v/%v", a.Name, r.Singular),
	})
	f := builder.NewField(constants.FIELD_PATH_NAME, builder.FieldTypeString()).
		SetNumber(constants.FIELD_PATH_NUMBER).
		SetComments(builder.Comments{
			LeadingComment: "The globally unique identifier for the resource",
		}).
		SetOptions(o)
	mb.AddField(f)
}

func addResourceField(r *api.Resource, resMsg Message, mb *builder.MessageBuilder) {
	o := &descriptorpb.FieldOptions{}
	proto.SetExtension(o, annotations.E_FieldBehavior, []annotations.FieldBehavior{annotations.FieldBehavior_REQUIRED})
	f := builder.NewField(cases.KebabToSnakeCase(r.Singular), resMsg.FieldType()).
		SetNumber(constants.FIELD_RESOURCE_NUMBER).
		SetComments(builder.Comments{
			LeadingComment: "The resource to perform the operation on.",
		}).
		SetOptions(o)
	mb.AddField(f)
}

func addResourcesField(r *api.Resource, resMsg Message, mb *builder.MessageBuilder) {
	f := builder.NewField(constants.FIELD_RESULTS_NAME, resMsg.FieldType()).SetNumber(constants.FIELD_RESULTS_NUMBER).SetComments(builder.Comments{
		LeadingComment: fmt.Sprintf("A list of %v", r.Plural),
	}).SetRepeated()
	mb.AddField(f)
}

func addPageToken(_ *api.Resource, mb *builder.MessageBuilder) {
	f := builder.NewField(constants.FIELD_PAGE_TOKEN_NAME, builder.FieldTypeString()).SetNumber(constants.FIELD_PAGE_TOKEN_NUMBER).SetComments(builder.Comments{
		LeadingComment: "The page token indicating the starting point of the page",
	})
	mb.AddField(f)
}

func addNextPageToken(_ *api.Resource, mb *builder.MessageBuilder) {
	f := builder.NewField(constants.FIELD_NEXT_PAGE_TOKEN_NAME, builder.FieldTypeString()).SetNumber(constants.FIELD_NEXT_PAGE_TOKEN_NUMBER).SetComments(builder.Comments{
		LeadingComment: "The page token indicating the ending point of this response.",
	})
	mb.AddField(f)
}

func addForceField(_ *api.API, _ *api.Resource, mb *builder.MessageBuilder) {
	o := &descriptorpb.FieldOptions{}
	proto.SetExtension(o, annotations.E_FieldBehavior, []annotations.FieldBehavior{annotations.FieldBehavior_OPTIONAL})
	f := builder.NewField(constants.FIELD_FORCE_NAME, builder.FieldTypeBool()).
		SetNumber(constants.FIELD_FORCE_NUMBER).
		SetComments(builder.Comments{
			LeadingComment: "If true, the resource will be deleted, even if children still exist.",
		}).
		SetOptions(o)
	mb.AddField(f)
}

func resourceDescriptor(a *api.API, r *api.Resource) *annotations.ResourceDescriptor {
	patterns := []string{
		strings.Join(r.PatternElems(), "/"),
	}
	return &annotations.ResourceDescriptor{
		Type:     fmt.Sprintf("%s/%s", a.Name, r.Singular),
		Pattern:  patterns,
		Singular: r.Singular,
		Plural:   r.Plural,
	}
}

func buildMethod(
	name string,
	request *builder.RpcType,
	response *builder.RpcType,
	isLongRunning bool,
) *builder.MethodBuilder {
	finalResponse := response
	options := &descriptorpb.MethodOptions{}
	if isLongRunning {
		finalResponse = builder.RpcTypeImportedMessage(operationMd, false)
		proto.SetExtension(options, apipb.E_OperationInfo, &apipb.OperationInfo{
			ResponseType: response.GetTypeName(),
		})
	}
	method := builder.NewMethod(name, request, finalResponse)
	method.SetOptions(options)
	return method
}
