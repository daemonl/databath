package query

type Context interface {
	GetValueFor(string) interface{}
	GetUserLevel() (isApplication bool, userAccessLevel uint64)
}

type MapContext struct {
	IsApplication   bool
	UserAccessLevel uint64
	Fields          map[string]interface{}
}

func (mc *MapContext) GetUserLevel() (bool, uint64) {
	return mc.IsApplication, mc.UserAccessLevel
}

func (mc *MapContext) GetValueFor(key string) interface{} {
	val, ok := mc.Fields[key]
	if !ok {
		return key
	}
	return val
}
