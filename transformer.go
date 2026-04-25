// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"reflect"

	core "dappco.re/go/core"

	"github.com/gin-gonic/gin"
)

// TransformerIn remaps an external request DTO into the handler-facing DTO.
// Implementations may also validate and normalise the inbound payload.
//
// Example:
//
//	type externalCreateUser struct {
//		FullName string `json:"full_name"`
//	}
//	type createUser struct {
//		Name string `json:"name"`
//	}
//	var tx api.TransformerIn[externalCreateUser, createUser]
type TransformerIn[I, O any] interface {
	TransformIn(*gin.Context, I) (O, error)
}

// TransformerOut remaps a handler-facing response DTO into an external
// response DTO inside the standard OK() envelope.
//
// Example:
//
//	var tx api.TransformerOut[user, externalUser]
type TransformerOut[I, O any] interface {
	TransformOut(*gin.Context, I) (O, error)
}

// TransformerInFunc adapts a function into a TransformerIn.
type TransformerInFunc[I, O any] func(*gin.Context, I) (O, error)

// TransformIn runs f.
func (f TransformerInFunc[I, O]) TransformIn(c *gin.Context, in I) (O, error) {
	return f(c, in)
}

// TransformerOutFunc adapts a function into a TransformerOut.
type TransformerOutFunc[I, O any] func(*gin.Context, I) (O, error)

// TransformOut runs f.
func (f TransformerOutFunc[I, O]) TransformOut(c *gin.Context, in I) (O, error) {
	return f(c, in)
}

// FieldRenamer remaps top-level JSON object fields. It implements both
// TransformerIn[map[string]any, map[string]any] and
// TransformerOut[map[string]any, map[string]any].
type FieldRenamer struct {
	Fields map[string]string
}

var _ TransformerIn[map[string]any, map[string]any] = FieldRenamer{}
var _ TransformerOut[map[string]any, map[string]any] = FieldRenamer{}

// RenameFields creates a top-level JSON object field renamer. The map keys are
// source field names and values are destination field names.
func RenameFields(fields map[string]string) FieldRenamer {
	return FieldRenamer{Fields: cloneStringMap(fields)}
}

// TransformIn renames inbound request fields.
func (r FieldRenamer) TransformIn(_ *gin.Context, payload map[string]any) (map[string]any, error) {
	return r.rename(payload), nil
}

// TransformOut renames outbound response fields.
func (r FieldRenamer) TransformOut(_ *gin.Context, payload map[string]any) (map[string]any, error) {
	return r.rename(payload), nil
}

func (r FieldRenamer) rename(payload map[string]any) map[string]any {
	if payload == nil {
		return nil
	}

	out := make(map[string]any, len(payload))
	for key, value := range payload {
		out[key] = value
	}

	for from, to := range r.Fields {
		from = core.Trim(from)
		to = core.Trim(to)
		if from == "" || to == "" || from == to {
			continue
		}
		value, ok := out[from]
		if !ok {
			continue
		}
		delete(out, from)
		out[to] = value
	}

	return out
}

const (
	transformerDirectionIn  = "in"
	transformerDirectionOut = "out"
)

var (
	ginContextReflectType = reflect.TypeOf((*gin.Context)(nil))
	errorReflectType      = reflect.TypeOf((*error)(nil)).Elem()
)

type compiledTransformer struct {
	method      reflect.Value
	inputType   reflect.Type
	withContext bool
}

func compileTransformerPipeline(direction string, raw any) ([]compiledTransformer, error) {
	if isNilValue(raw) {
		return nil, nil
	}

	if hasTransformerMethod(direction, raw) {
		transformer, err := compileTransformer(direction, raw)
		if err != nil {
			return nil, err
		}
		return []compiledTransformer{transformer}, nil
	}

	value := reflect.ValueOf(raw)
	switch value.Kind() {
	case reflect.Array, reflect.Slice:
		out := make([]compiledTransformer, 0, value.Len())
		for i := 0; i < value.Len(); i++ {
			transformers, err := compileTransformerPipeline(direction, value.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			out = append(out, transformers...)
		}
		return out, nil
	default:
		transformer, err := compileTransformer(direction, raw)
		if err != nil {
			return nil, err
		}
		return []compiledTransformer{transformer}, nil
	}
}

func compileTransformer(direction string, raw any) (compiledTransformer, error) {
	methodName := transformerMethodName(direction)
	method := reflect.ValueOf(raw).MethodByName(methodName)
	if !method.IsValid() {
		return compiledTransformer{}, core.E("Transformer", core.Sprintf("missing %s method", methodName), nil)
	}

	methodType := method.Type()
	if methodType.NumOut() != 2 || !methodType.Out(1).Implements(errorReflectType) {
		return compiledTransformer{}, core.E("Transformer", core.Sprintf("%s must return (O, error)", methodName), nil)
	}

	transformer := compiledTransformer{method: method}

	switch methodType.NumIn() {
	case 1:
		transformer.inputType = methodType.In(0)
	case 2:
		if methodType.In(0) != ginContextReflectType {
			return compiledTransformer{}, core.E("Transformer", core.Sprintf("%s first argument must be *gin.Context", methodName), nil)
		}
		transformer.inputType = methodType.In(1)
		transformer.withContext = true
	default:
		return compiledTransformer{}, core.E("Transformer", core.Sprintf("%s must accept I or (*gin.Context, I)", methodName), nil)
	}

	return transformer, nil
}

func hasTransformerMethod(direction string, raw any) bool {
	if isNilValue(raw) {
		return false
	}
	return reflect.ValueOf(raw).MethodByName(transformerMethodName(direction)).IsValid()
}

func transformerMethodName(direction string) string {
	if direction == transformerDirectionOut {
		return "TransformOut"
	}
	return "TransformIn"
}

func (t compiledTransformer) transform(c *gin.Context, payload []byte) ([]byte, error) {
	input, err := decodeTransformerInput(payload, t.inputType)
	if err != nil {
		return nil, err
	}

	args := []reflect.Value{input}
	if t.withContext {
		args = []reflect.Value{reflect.ValueOf(c), input}
	}

	out := t.method.Call(args)
	if !out[1].IsNil() {
		return nil, out[1].Interface().(error)
	}

	return encodeTransformerOutput(out[0])
}

func decodeTransformerInput(payload []byte, inputType reflect.Type) (reflect.Value, error) {
	if inputType.Kind() == reflect.Pointer {
		target := reflect.New(inputType.Elem())
		if err := unmarshalTransformerPayload(payload, target.Interface()); err != nil {
			return reflect.Value{}, err
		}
		return target, nil
	}

	target := reflect.New(inputType)
	if err := unmarshalTransformerPayload(payload, target.Interface()); err != nil {
		return reflect.Value{}, err
	}
	return target.Elem(), nil
}

func unmarshalTransformerPayload(payload []byte, target any) error {
	result := core.JSONUnmarshal(payload, target)
	if result.OK {
		return nil
	}
	if err, ok := result.Value.(error); ok {
		return core.E("Transformer.Decode", "decode payload", err)
	}
	return core.E("Transformer.Decode", "decode payload", nil)
}

func encodeTransformerOutput(value reflect.Value) ([]byte, error) {
	var payload any
	if value.IsValid() {
		payload = value.Interface()
	}

	result := core.JSONMarshal(payload)
	if !result.OK {
		if err, ok := result.Value.(error); ok {
			return nil, core.E("Transformer.Encode", "encode payload", err)
		}
		return nil, core.E("Transformer.Encode", "encode payload", nil)
	}

	data, ok := result.Value.([]byte)
	if !ok {
		return nil, core.E("Transformer.Encode", "encoded payload was not bytes", nil)
	}
	return data, nil
}

func runTransformerPipeline(c *gin.Context, payload []byte, pipeline []compiledTransformer) ([]byte, error) {
	var err error
	for _, transformer := range pipeline {
		payload, err = transformer.transform(c, payload)
		if err != nil {
			return nil, err
		}
	}
	return payload, nil
}

func transformerRouteKey(method, path string) string {
	return core.Upper(core.Trim(method)) + " " + normaliseTransformerPath(path)
}

func joinTransformerRoutePath(basePath, routePath string) string {
	basePath = normaliseTransformerPath(basePath)
	routePath = normaliseTransformerPath(routePath)

	if routePath == "/" {
		return basePath
	}
	if basePath == "/" {
		return routePath
	}

	if core.HasPrefix(routePath, "/") {
		routePath = routePath[1:]
	}
	return trimTrailingSlashes(basePath) + "/" + routePath
}

func normaliseTransformerPath(path string) string {
	path = core.Trim(path)
	if path == "" {
		return "/"
	}

	segments := core.Split(path, "/")
	cleaned := make([]string, 0, len(segments))
	for _, segment := range segments {
		segment = core.Trim(segment)
		if segment == "" {
			continue
		}
		switch {
		case core.HasPrefix(segment, ":") && len(segment) > 1:
			segment = "{" + segment[1:] + "}"
		case core.HasPrefix(segment, "*") && len(segment) > 1:
			segment = "{" + segment[1:] + "}"
		}
		cleaned = append(cleaned, segment)
	}
	if len(cleaned) == 0 {
		return "/"
	}
	b := core.NewBuilder()
	for _, segment := range cleaned {
		b.WriteByte('/')
		b.WriteString(segment)
	}
	return b.String()
}
