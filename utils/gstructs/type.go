package gstructs

// Signature 返回类型的唯一签名
func (t Type) Signature() string {
	if t.Type == nil {
		return ""
	}
	return t.PkgPath() + "/" + t.String()
}

// FieldKeys 返回当前结构体/映射的键
func (t Type) FieldKeys() []string {
	if t.Type == nil {
		return nil
	}

	numField := t.NumField()
	if numField == 0 {
		return nil
	}

	keys := make([]string, numField)
	for i := 0; i < numField; i++ {
		keys[i] = t.Field(i).Name
	}
	return keys
}
