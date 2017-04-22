package shared

import (
	"fmt"
	"reflect"
	"sync"
	"context"
)

// TODO need to check context, if create, threshold is 0, if put, patch, threshold is 1
func ValidateUniqueness(subj *Resource, sch *Schema, repo Repository, ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case error:
				err = r.(error)
			default:
				err = Error.Text("%v", r)
			}
		}
	}()

	uniquenessValidatorInstance.validateUniquenessWithReflection(reflect.ValueOf(subj.Complex), sch.ToAttribute(), repo, ctx)
	return
}

var (
	oneUniquenessValidator      sync.Once
	uniquenessValidatorInstance *uniquenessValidator
)

func init() {
	oneUniquenessValidator.Do(func() {
		uniquenessValidatorInstance = &uniquenessValidator{}
	})
}

type uniquenessValidator struct{}

func (uv *uniquenessValidator) validateUniquenessWithReflection(v reflect.Value, guide *Attribute, repo Repository, ctx context.Context) {
	for _, attr := range guide.SubAttributes {
		v0 := v.MapIndex(reflect.ValueOf(attr.Name))
		if !attr.Assigned(v0) {
			continue
		}
		if v0.Kind() == reflect.Interface {
			v0 = v0.Elem()
		}

		switch attr.Uniqueness {
		case Server, Global:
			query := fmt.Sprintf("%s eq \"%v\"", attr.Assist.Path, v0.Interface())
			count, err := repo.Count(query)
			if err != nil {
				uv.throw(err, ctx)
			} else if count > 0 {
				uv.throw(Error.Duplicate(attr.Assist.Path, v0.Interface()), ctx)
			}
		}

		if attr.ExpectsComplex() && v0.Kind() == reflect.Map {
			uv.validateUniquenessWithReflection(v0, attr, repo, ctx)
		}
	}
}

func (uv *uniquenessValidator) throw(err error, ctx context.Context) {
	panic(err)
}