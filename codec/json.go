package codec

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	cjsonpb "github.com/cosmos/gogoproto/jsonpb"
	cproto "github.com/cosmos/gogoproto/proto"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

var defaultJM = &jsonpb.Marshaler{OrigName: true, EmitDefaults: true, AnyResolver: nil}

// ProtoMarshalJSON provides an auxiliary function to return Proto3 JSON encoded
// bytes of a message.
func ProtoMarshalJSON(msg proto.Message, resolver jsonpb.AnyResolver) ([]byte, error) {
	// We use the OrigName because camel casing fields just doesn't make sense.
	// EmitDefaults is also often the more expected behavior for CLI users
	jm := defaultJM
	if resolver != nil {
		jm = &jsonpb.Marshaler{OrigName: true, EmitDefaults: true, AnyResolver: resolver}
	}
	err := types.UnpackInterfaces(msg, types.ProtoJSONPacker{JSONPBMarshaler: jm})
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err := jm.Marshal(buf, msg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ProtoMarshalJSON provides an auxiliary function to return Proto3 JSON encoded
// bytes of a message.
func CustomProtoMarshalJSON(msg cproto.Message) ([]byte, error) {
	// We use the OrigName because camel casing fields just doesn't make sense.
	// EmitDefaults is also often the more expected behavior for CLI users
	jm := &cjsonpb.Marshaler{OrigName: true, EmitDefaults: true, AnyResolver: CustomResolver{}}
	err := types.UnpackInterfaces(msg, types.CustomProtoJSONPacker{JSONPBMarshaler: jm})
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err := jm.Marshal(buf, msg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

var _ cjsonpb.AnyResolver = &CustomResolver{}

type CustomResolver struct{}

// Resolve is the AnyResolver.Resolve method.
func (d CustomResolver) Resolve(typeURL string) (cproto.Message, error) {
	// Only the part of typeURL after the last slash is relevant.
	mname := typeURL
	if slash := strings.LastIndex(mname, "/"); slash >= 0 {
		mname = mname[slash+1:]
	}
	mt := proto.MessageType(mname)
	if mt == nil {
		return nil, fmt.Errorf("unknown message type %q", mname)
	}
	return reflect.New(mt.Elem()).Interface().(proto.Message), nil
}
