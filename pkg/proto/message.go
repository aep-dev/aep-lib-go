package proto

import (
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/builder"
	"google.golang.org/protobuf/types/descriptorpb"
)

// type Message wraps a MessageBuilder or a
// message descriptor. It abstracts those concrete
// implementations to provide a common interface.
type Message interface {
	FieldType() *builder.FieldType
	RpcType() *builder.RpcType
	Options() *descriptorpb.MessageOptions
	AddMessage(fb *builder.FileBuilder)
}

// A variant that wraps MessageBuilder, for messages
// created during generation.
type WrappedMessageBuilder struct {
	mb *builder.MessageBuilder
}

func NewWrappedMessageBuilder(mb *builder.MessageBuilder) WrappedMessageBuilder {
	return WrappedMessageBuilder{mb: mb}
}

func (wrapped WrappedMessageBuilder) FieldType() *builder.FieldType {
	return builder.FieldTypeMessage(wrapped.mb)
}

func (wrapped WrappedMessageBuilder) Options() *descriptorpb.MessageOptions {
	return wrapped.mb.Options
}

func (wrapped WrappedMessageBuilder) RpcType() *builder.RpcType {
	return builder.RpcTypeMessage(wrapped.mb, false)
}

func (wrapped WrappedMessageBuilder) AddMessage(fb *builder.FileBuilder) {
	fb.AddMessage(wrapped.mb)
}

// A variant that wraps MessageDescriptor, for messages
// referenced
type WrappedMessageDescriptor struct {
	md *desc.MessageDescriptor
}

func NewWrappedMessageDescriptor(md *desc.MessageDescriptor) WrappedMessageDescriptor {
	return WrappedMessageDescriptor{md: md}
}

func (wrapped WrappedMessageDescriptor) FieldType() *builder.FieldType {
	return builder.FieldTypeImportedMessage(wrapped.md)
}

func (wrapped WrappedMessageDescriptor) Options() *descriptorpb.MessageOptions {
	return nil
}

func (wrapped WrappedMessageDescriptor) RpcType() *builder.RpcType {
	return builder.RpcTypeImportedMessage(wrapped.md, false)
}

func (wrapped WrappedMessageDescriptor) AddMessage(fb *builder.FileBuilder) {
	// noop
}
