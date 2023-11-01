package signal

import (
	"arthas/protobuf"
	"google.golang.org/protobuf/proto"
)

type XmppStanzaElementBuilder struct {
	protobuf.XmppStanzaElement
}

func NewXmppStanzaElementBuilder(name string) *XmppStanzaElementBuilder {
	return &XmppStanzaElementBuilder{
		XmppStanzaElement: protobuf.XmppStanzaElement{
			Name: proto.String(name),
		},
	}
}

func (b *XmppStanzaElementBuilder) WithAttribute(attribute *protobuf.XmppAttribute) *XmppStanzaElementBuilder {
	b.Attributes = append(b.Attributes, attribute)
	return b
}

func (b *XmppStanzaElementBuilder) WithChild(child *protobuf.XmppStanzaElement) *XmppStanzaElementBuilder {
	b.Children = append(b.Children, child)
	return b
}

func (b *XmppStanzaElementBuilder) WithStringValue(value string) *XmppStanzaElementBuilder {
	b.Value = &protobuf.XmppValue{
		Value: &protobuf.XmppValue_String_{String_: value},
	}
	return b
}

func (b *XmppStanzaElementBuilder) WithDataValue(data []byte) *XmppStanzaElementBuilder {
	b.Value = &protobuf.XmppValue{
		Value: &protobuf.XmppValue_Data{Data: data},
	}
	return b
}

func (b *XmppStanzaElementBuilder) Build() *protobuf.XmppStanzaElement {
	return &b.XmppStanzaElement
}

type XmppAttributeBuilder struct {
	protobuf.XmppAttribute
}

func NewXmppAttributeBuilder(key string) *XmppAttributeBuilder {
	return &XmppAttributeBuilder{
		XmppAttribute: protobuf.XmppAttribute{
			Key: proto.String(key),
		},
	}
}

func (b *XmppAttributeBuilder) WithStringValue(value string) *XmppAttributeBuilder {
	b.Value = &protobuf.XmppAttributeValue{
		Value: &protobuf.XmppAttributeValue_String_{String_: value},
	}
	return b
}

func (b *XmppAttributeBuilder) WithDataValue(data []byte) *XmppAttributeBuilder {
	b.Value = &protobuf.XmppAttributeValue{
		Value: &protobuf.XmppAttributeValue_Data{Data: data},
	}
	return b
}

func (b *XmppAttributeBuilder) WithWADomainJID(value string) *XmppAttributeBuilder {
	b.Value = &protobuf.XmppAttributeValue{
		Value: &protobuf.XmppAttributeValue_WADomainJID{WADomainJID: value},
	}
	return b
}

func (b *XmppAttributeBuilder) WithWAPhoneNumberUserJID(value string) *XmppAttributeBuilder {
	b.Value = &protobuf.XmppAttributeValue{
		Value: &protobuf.XmppAttributeValue_WAPhoneNumberUserJID{WAPhoneNumberUserJID: value},
	}
	return b
}

// ... 其他 XmppAttributeValue 字段的构建函数

func (b *XmppAttributeBuilder) Build() *protobuf.XmppAttribute {
	return &b.XmppAttribute
}
