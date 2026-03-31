package selector

func (s Selector) IsDisplayID() bool {
	return s.Kind == KindDisplayID
}

func (s Selector) IsCSS() bool {
	return s.Kind == KindCSS
}

func (s Selector) IsXPath() bool {
	return s.Kind == KindXPath
}
