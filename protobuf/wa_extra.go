package protobuf

func (x *XmppStanzaElement) GetChildrenByName(name string) *XmppStanzaElement {
	for _, child := range x.Children {
		if *child.Name == name {
			return child
		}
	}
	return nil
}

func (x *XmppStanzaElement) GetAttributeByKey(key string) *XmppAttribute {
	for _, attr := range x.Attributes {
		if *attr.Key == key {
			return attr
		}
	}
	return nil
}
