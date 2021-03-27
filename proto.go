package zapp

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func Reflect(key string, v protoreflect.Message) zap.Field {
	return zap.Object(key, reflect(v))
}

func Struct(key string, v *structpb.Struct) zap.Field {
	return zap.Object(key, structMarshaler(v))
}

func Any(key string, v *anypb.Any) zap.Field {
	return zap.Object(key, anyMarshaler(key, v))
}

func Timestamp(key string, v *timestamppb.Timestamp) zap.Field {
	return zap.Time(key, v.AsTime())
}

func Duration(key string, v *durationpb.Duration) zap.Field {
	return zap.Duration(key, v.AsDuration())
}

func FieldMask(key string, v *fieldmaskpb.FieldMask) zap.Field {
	return zap.String(key, strings.Join(v.GetPaths(), ","))
}

func BoolValue(key string, v *wrapperspb.BoolValue) zap.Field {
	if v == nil {
		return zap.Reflect(key, nil)
	}
	return zap.Bool(key, v.Value)
}

func BytesValue(key string, v *wrapperspb.BytesValue) zap.Field {
	if v == nil {
		return zap.Reflect(key, nil)
	}
	return zap.Binary(key, v.Value)
}

func DoubleValue(key string, v *wrapperspb.DoubleValue) zap.Field {
	if v == nil {
		return zap.Reflect(key, nil)
	}
	return zap.Float64(key, v.Value)
}

func FloatValue(key string, v *wrapperspb.FloatValue) zap.Field {
	if v == nil {
		return zap.Reflect(key, nil)
	}
	return zap.Float32(key, v.Value)
}

func StringValue(key string, v *wrapperspb.StringValue) zap.Field {
	if v == nil {
		return zap.Reflect(key, nil)
	}
	return zap.String(key, v.Value)
}

func Int32Value(key string, v *wrapperspb.Int32Value) zap.Field {
	if v == nil {
		return zap.Reflect(key, nil)
	}
	return zap.Int32(key, v.Value)
}

func Int64Value(key string, v *wrapperspb.Int64Value) zap.Field {
	if v == nil {
		return zap.Reflect(key, nil)
	}
	return zap.Int64(key, v.Value)
}

func Uint32Value(key string, v *wrapperspb.UInt32Value) zap.Field {
	if v == nil {
		return zap.Reflect(key, nil)
	}
	return zap.Uint32(key, v.Value)
}

func Uint64Value(key string, v *wrapperspb.UInt64Value) zap.Field {
	if v == nil {
		return zap.Reflect(key, nil)
	}
	return zap.Uint64(key, v.Value)
}

func P(key string, m proto.Message) zap.Field {
	switch v := m.(type) {
	case zapcore.ObjectMarshaler:
		return zap.Object(key, v)
	case *anypb.Any:
		return Any(key, v)
	case *timestamppb.Timestamp:
		return Timestamp(key, v)
	case *durationpb.Duration:
		return Duration(key, v)
	case *fieldmaskpb.FieldMask:
		return zap.String(key, strings.Join(v.GetPaths(), ","))
	case *wrapperspb.BoolValue:
		return BoolValue(key, v)
	case *wrapperspb.BytesValue:
		return zap.Binary(key, v.GetValue())
	case *wrapperspb.DoubleValue:
		return zap.Float64(key, v.GetValue())
	case *wrapperspb.FloatValue:
		return zap.Float32(key, v.GetValue())
	case *wrapperspb.StringValue:
		return zap.String(key, v.GetValue())
	case *wrapperspb.Int32Value:
		return zap.Int32(key, v.GetValue())
	case *wrapperspb.Int64Value:
		return zap.Int64(key, v.GetValue())
	case *wrapperspb.UInt32Value:
		return zap.Uint32(key, v.GetValue())
	case *wrapperspb.UInt64Value:
		return zap.Uint64(key, v.GetValue())
	default:
		return zap.Object(key, reflect(m.ProtoReflect()))

	}
}

func reflect(m protoreflect.Message) zapcore.ObjectMarshalerFunc {
	return func(enc zapcore.ObjectEncoder) error {
		m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
			if fd.IsExtension() || fd.IsPlaceholder() {
				return true
			}

			k := fd.JSONName()
			if of := fd.ContainingOneof(); of != nil {
				enc.AddObject(string(of.Name()), zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
					writePrimitive(k, v, enc, fd.IsList(), fd.Kind())
					return nil
				}))

				return true
			}

			if fd.IsMap() {
				enc.AddObject(k, zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
					mp := v.Map()
					mp.Range(func(mk protoreflect.MapKey, v protoreflect.Value) bool {
						writePrimitive(mk.String(), v, enc, fd.MapValue().IsList(), fd.MapValue().Kind())
						return true
					})
					return nil
				}))

				return true
			}

			writePrimitive(k, v, enc, fd.IsList(), fd.Kind())

			return true
		})
		return nil
	}
}

func writePrimitive(k string, v protoreflect.Value, enc zapcore.ObjectEncoder, repeated bool, kind protoreflect.Kind) {
	switch kind {
	case protoreflect.BoolKind:
		if repeated {
			enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
				l := v.List()
				for i := 0; i < l.Len(); i++ {
					ae.AppendBool(l.Get(i).Bool())
				}
				return nil
			}))
		} else {
			enc.AddBool(k, v.Bool())
		}
	case protoreflect.StringKind, protoreflect.EnumKind:
		if repeated {
			enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
				l := v.List()
				for i := 0; i < l.Len(); i++ {
					ae.AppendString(l.Get(i).String())
				}
				return nil
			}))
		} else {
			enc.AddString(k, v.String())
		}
	case protoreflect.Int32Kind,
		protoreflect.Sint32Kind,
		protoreflect.Sint64Kind,
		protoreflect.Int64Kind,
		protoreflect.Sfixed32Kind,
		protoreflect.Sfixed64Kind:
		if repeated {
			enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
				l := v.List()
				for i := 0; i < l.Len(); i++ {
					ae.AppendInt64(l.Get(i).Int())
				}
				return nil
			}))
		} else {
			enc.AddInt64(k, v.Int())
		}
	case protoreflect.Uint32Kind,
		protoreflect.Uint64Kind,
		protoreflect.Fixed32Kind,
		protoreflect.Fixed64Kind:
		if repeated {
			enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
				l := v.List()
				for i := 0; i < l.Len(); i++ {
					ae.AppendUint64(l.Get(i).Uint())
				}
				return nil
			}))
		} else {
			enc.AddUint64(k, v.Uint())
		}
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		if repeated {
			enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
				l := v.List()
				for i := 0; i < l.Len(); i++ {
					ae.AppendFloat64(l.Get(i).Float())
				}
				return nil
			}))
		} else {
			enc.AddFloat64(k, v.Float())
		}
	case protoreflect.BytesKind:
		if repeated {
			enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
				l := v.List()
				for i := 0; i < l.Len(); i++ {
					ae.AppendByteString(l.Get(i).Bytes())
				}
				return nil
			}))
		} else {
			enc.AddByteString(k, v.Bytes())
		}
	case protoreflect.MessageKind:
		if !repeated {
			switch m := v.Message().Interface().(type) {
			case *timestamppb.Timestamp:
				enc.AddTime(k, m.AsTime())
			case *durationpb.Duration:
				enc.AddDuration(k, m.AsDuration())
			case *anypb.Any:
				enc.AddObject(k, anyMarshaler(k, m))
			case *structpb.Struct:
				enc.AddObject(k, structMarshaler(m))
			case *fieldmaskpb.FieldMask:
				enc.AddString(k, strings.Join(m.GetPaths(), ","))
			case *wrapperspb.BoolValue:
				enc.AddBool(k, m.Value)
			case *wrapperspb.BytesValue:
				enc.AddBinary(k, m.Value)
			case *wrapperspb.DoubleValue:
				enc.AddFloat64(k, m.Value)
			case *wrapperspb.FloatValue:
				enc.AddFloat32(k, m.Value)
			case *wrapperspb.StringValue:
				enc.AddString(k, m.Value)
			case *wrapperspb.Int32Value:
				enc.AddInt32(k, m.Value)
			case *wrapperspb.Int64Value:
				enc.AddInt64(k, m.Value)
			case *wrapperspb.UInt32Value:
				enc.AddUint32(k, m.Value)
			case *wrapperspb.UInt64Value:
				enc.AddUint64(k, m.Value)
			default:
				enc.AddObject(k, reflect(v.Message()))
			}
		} else if l := v.List(); l.Len() > 0 {
			switch l.Get(0).Message().Interface().(type) {
			case *timestamppb.Timestamp:
				enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
					for i := 0; i < l.Len(); i++ {
						ae.AppendTime(l.Get(i).Message().Interface().(*timestamppb.Timestamp).AsTime())
					}
					return nil
				}))
			case *durationpb.Duration:
				enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
					for i := 0; i < l.Len(); i++ {
						ae.AppendDuration(l.Get(i).Message().Interface().(*durationpb.Duration).AsDuration())
					}
					return nil
				}))
			case *anypb.Any:
				enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
					for i := 0; i < l.Len(); i++ {
						m := l.Get(i).Message().Interface().(*anypb.Any)
						ae.AppendObject(anyMarshaler(k, m))
					}
					return nil
				}))
			case *structpb.Struct:
				enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
					for i := 0; i < l.Len(); i++ {
						m := l.Get(i).Message().Interface().(*structpb.Struct)
						ae.AppendObject(structMarshaler(m))
					}
					return nil
				}))
			case *fieldmaskpb.FieldMask:
				enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
					for i := 0; i < l.Len(); i++ {
						ae.AppendString(strings.Join(l.Get(i).Message().Interface().(*fieldmaskpb.FieldMask).GetPaths(), ","))
					}
					return nil
				}))
			case *wrapperspb.BoolValue:
				enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
					for i := 0; i < l.Len(); i++ {
						ae.AppendBool(l.Get(i).Message().Interface().(*wrapperspb.BoolValue).Value)
					}
					return nil
				}))
			case *wrapperspb.BytesValue:
				enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
					for i := 0; i < l.Len(); i++ {
						ae.AppendByteString(l.Get(i).Message().Interface().(*wrapperspb.BytesValue).Value)
					}
					return nil
				}))
			case *wrapperspb.DoubleValue:
				enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
					for i := 0; i < l.Len(); i++ {
						ae.AppendFloat64(l.Get(i).Message().Interface().(*wrapperspb.DoubleValue).Value)
					}
					return nil
				}))
			case *wrapperspb.FloatValue:
				enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
					for i := 0; i < l.Len(); i++ {
						ae.AppendFloat32(l.Get(i).Message().Interface().(*wrapperspb.FloatValue).Value)
					}
					return nil
				}))
			case *wrapperspb.StringValue:
				enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
					for i := 0; i < l.Len(); i++ {
						ae.AppendString(l.Get(i).Message().Interface().(*wrapperspb.StringValue).Value)
					}
					return nil
				}))
			case *wrapperspb.Int32Value:
				enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
					for i := 0; i < l.Len(); i++ {
						ae.AppendInt32(l.Get(i).Message().Interface().(*wrapperspb.Int32Value).Value)
					}
					return nil
				}))
			case *wrapperspb.Int64Value:
				enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
					for i := 0; i < l.Len(); i++ {
						ae.AppendInt64(l.Get(i).Message().Interface().(*wrapperspb.Int64Value).Value)
					}
					return nil
				}))
			case *wrapperspb.UInt32Value:
				enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
					for i := 0; i < l.Len(); i++ {
						ae.AppendUint32(l.Get(i).Message().Interface().(*wrapperspb.UInt32Value).Value)
					}
					return nil
				}))
			case *wrapperspb.UInt64Value:
				enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
					for i := 0; i < l.Len(); i++ {
						ae.AppendUint64(l.Get(i).Message().Interface().(*wrapperspb.UInt64Value).Value)
					}
					return nil
				}))
			default:
				enc.AddArray(k, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
					for i := 0; i < l.Len(); i++ {
						ae.AppendObject(reflect(l.Get(i).Message()))
					}
					return nil
				}))
			}
		}
	case protoreflect.GroupKind:
		return
	}
}

func anyMarshaler(k string, m *anypb.Any) zapcore.ObjectMarshalerFunc {
	return func(oe zapcore.ObjectEncoder) error {
		oe.AddString("@typeUrl", m.TypeUrl)
		pm, err := m.UnmarshalNew()
		if err != nil {
			return fmt.Errorf("error logging %s of type any: %w", k, err)
		}
		return reflect(pm.ProtoReflect()).MarshalLogObject(oe)
	}
}

func structMarshaler(m *structpb.Struct) zapcore.ObjectMarshalerFunc {
	return func(oe zapcore.ObjectEncoder) error {
		for k, v := range m.GetFields() {
			switch v := v.Kind.(type) {
			case *structpb.Value_StructValue:
				oe.AddObject(k, structMarshaler(v.StructValue))
			case *structpb.Value_BoolValue:
				oe.AddBool(k, v.BoolValue)
			case *structpb.Value_ListValue:
				oe.AddArray(k, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
					for _, lv := range v.ListValue.GetValues() {
						structValueMarshaler(lv, ae)
					}
					return nil
				}))
			case *structpb.Value_NullValue:
				oe.AddReflected(k, nil)
			case *structpb.Value_NumberValue:
				oe.AddFloat64(k, v.NumberValue)
			case *structpb.Value_StringValue:
				oe.AddString(k, v.StringValue)
			}
		}
		return nil
	}
}

func structValueMarshaler(v *structpb.Value, ae zapcore.ArrayEncoder) {
	switch v := v.Kind.(type) {
	case *structpb.Value_StructValue:
		ae.AppendObject(structMarshaler(v.StructValue))
	case *structpb.Value_BoolValue:
		ae.AppendBool(v.BoolValue)
	case *structpb.Value_ListValue:
		ae.AppendArray(zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
			for _, lv := range v.ListValue.GetValues() {
				structValueMarshaler(lv, ae)
			}
			return nil
		}))
	case *structpb.Value_NullValue:
		ae.AppendReflected(nil)
	case *structpb.Value_NumberValue:
		ae.AppendFloat64(v.NumberValue)
	case *structpb.Value_StringValue:
		ae.AppendString(v.StringValue)
	}
}
