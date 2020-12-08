// Code generated by protoc-gen-go. DO NOT EDIT.
// source: vtctldata.proto

package vtctldata

import (
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	logutil "vitess.io/vitess/go/vt/proto/logutil"
	topodata "vitess.io/vitess/go/vt/proto/topodata"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// ExecuteVtctlCommandRequest is the payload for ExecuteVtctlCommand.
// timeouts are in nanoseconds.
type ExecuteVtctlCommandRequest struct {
	Args                 []string `protobuf:"bytes,1,rep,name=args,proto3" json:"args,omitempty"`
	ActionTimeout        int64    `protobuf:"varint,2,opt,name=action_timeout,json=actionTimeout,proto3" json:"action_timeout,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ExecuteVtctlCommandRequest) Reset()         { *m = ExecuteVtctlCommandRequest{} }
func (m *ExecuteVtctlCommandRequest) String() string { return proto.CompactTextString(m) }
func (*ExecuteVtctlCommandRequest) ProtoMessage()    {}
func (*ExecuteVtctlCommandRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_f41247b323a1ab2e, []int{0}
}

func (m *ExecuteVtctlCommandRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ExecuteVtctlCommandRequest.Unmarshal(m, b)
}
func (m *ExecuteVtctlCommandRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ExecuteVtctlCommandRequest.Marshal(b, m, deterministic)
}
func (m *ExecuteVtctlCommandRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ExecuteVtctlCommandRequest.Merge(m, src)
}
func (m *ExecuteVtctlCommandRequest) XXX_Size() int {
	return xxx_messageInfo_ExecuteVtctlCommandRequest.Size(m)
}
func (m *ExecuteVtctlCommandRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ExecuteVtctlCommandRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ExecuteVtctlCommandRequest proto.InternalMessageInfo

func (m *ExecuteVtctlCommandRequest) GetArgs() []string {
	if m != nil {
		return m.Args
	}
	return nil
}

func (m *ExecuteVtctlCommandRequest) GetActionTimeout() int64 {
	if m != nil {
		return m.ActionTimeout
	}
	return 0
}

// ExecuteVtctlCommandResponse is streamed back by ExecuteVtctlCommand.
type ExecuteVtctlCommandResponse struct {
	Event                *logutil.Event `protobuf:"bytes,1,opt,name=event,proto3" json:"event,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *ExecuteVtctlCommandResponse) Reset()         { *m = ExecuteVtctlCommandResponse{} }
func (m *ExecuteVtctlCommandResponse) String() string { return proto.CompactTextString(m) }
func (*ExecuteVtctlCommandResponse) ProtoMessage()    {}
func (*ExecuteVtctlCommandResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_f41247b323a1ab2e, []int{1}
}

func (m *ExecuteVtctlCommandResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ExecuteVtctlCommandResponse.Unmarshal(m, b)
}
func (m *ExecuteVtctlCommandResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ExecuteVtctlCommandResponse.Marshal(b, m, deterministic)
}
func (m *ExecuteVtctlCommandResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ExecuteVtctlCommandResponse.Merge(m, src)
}
func (m *ExecuteVtctlCommandResponse) XXX_Size() int {
	return xxx_messageInfo_ExecuteVtctlCommandResponse.Size(m)
}
func (m *ExecuteVtctlCommandResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ExecuteVtctlCommandResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ExecuteVtctlCommandResponse proto.InternalMessageInfo

func (m *ExecuteVtctlCommandResponse) GetEvent() *logutil.Event {
	if m != nil {
		return m.Event
	}
	return nil
}

type GetKeyspacesRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetKeyspacesRequest) Reset()         { *m = GetKeyspacesRequest{} }
func (m *GetKeyspacesRequest) String() string { return proto.CompactTextString(m) }
func (*GetKeyspacesRequest) ProtoMessage()    {}
func (*GetKeyspacesRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_f41247b323a1ab2e, []int{2}
}

func (m *GetKeyspacesRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetKeyspacesRequest.Unmarshal(m, b)
}
func (m *GetKeyspacesRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetKeyspacesRequest.Marshal(b, m, deterministic)
}
func (m *GetKeyspacesRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetKeyspacesRequest.Merge(m, src)
}
func (m *GetKeyspacesRequest) XXX_Size() int {
	return xxx_messageInfo_GetKeyspacesRequest.Size(m)
}
func (m *GetKeyspacesRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetKeyspacesRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetKeyspacesRequest proto.InternalMessageInfo

type GetKeyspacesResponse struct {
	Keyspaces            []*Keyspace `protobuf:"bytes,1,rep,name=keyspaces,proto3" json:"keyspaces,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *GetKeyspacesResponse) Reset()         { *m = GetKeyspacesResponse{} }
func (m *GetKeyspacesResponse) String() string { return proto.CompactTextString(m) }
func (*GetKeyspacesResponse) ProtoMessage()    {}
func (*GetKeyspacesResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_f41247b323a1ab2e, []int{3}
}

func (m *GetKeyspacesResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetKeyspacesResponse.Unmarshal(m, b)
}
func (m *GetKeyspacesResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetKeyspacesResponse.Marshal(b, m, deterministic)
}
func (m *GetKeyspacesResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetKeyspacesResponse.Merge(m, src)
}
func (m *GetKeyspacesResponse) XXX_Size() int {
	return xxx_messageInfo_GetKeyspacesResponse.Size(m)
}
func (m *GetKeyspacesResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetKeyspacesResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetKeyspacesResponse proto.InternalMessageInfo

func (m *GetKeyspacesResponse) GetKeyspaces() []*Keyspace {
	if m != nil {
		return m.Keyspaces
	}
	return nil
}

type GetKeyspaceRequest struct {
	Keyspace             string   `protobuf:"bytes,1,opt,name=keyspace,proto3" json:"keyspace,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetKeyspaceRequest) Reset()         { *m = GetKeyspaceRequest{} }
func (m *GetKeyspaceRequest) String() string { return proto.CompactTextString(m) }
func (*GetKeyspaceRequest) ProtoMessage()    {}
func (*GetKeyspaceRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_f41247b323a1ab2e, []int{4}
}

func (m *GetKeyspaceRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetKeyspaceRequest.Unmarshal(m, b)
}
func (m *GetKeyspaceRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetKeyspaceRequest.Marshal(b, m, deterministic)
}
func (m *GetKeyspaceRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetKeyspaceRequest.Merge(m, src)
}
func (m *GetKeyspaceRequest) XXX_Size() int {
	return xxx_messageInfo_GetKeyspaceRequest.Size(m)
}
func (m *GetKeyspaceRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetKeyspaceRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetKeyspaceRequest proto.InternalMessageInfo

func (m *GetKeyspaceRequest) GetKeyspace() string {
	if m != nil {
		return m.Keyspace
	}
	return ""
}

type Keyspace struct {
	Name                 string             `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Keyspace             *topodata.Keyspace `protobuf:"bytes,2,opt,name=keyspace,proto3" json:"keyspace,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *Keyspace) Reset()         { *m = Keyspace{} }
func (m *Keyspace) String() string { return proto.CompactTextString(m) }
func (*Keyspace) ProtoMessage()    {}
func (*Keyspace) Descriptor() ([]byte, []int) {
	return fileDescriptor_f41247b323a1ab2e, []int{5}
}

func (m *Keyspace) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Keyspace.Unmarshal(m, b)
}
func (m *Keyspace) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Keyspace.Marshal(b, m, deterministic)
}
func (m *Keyspace) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Keyspace.Merge(m, src)
}
func (m *Keyspace) XXX_Size() int {
	return xxx_messageInfo_Keyspace.Size(m)
}
func (m *Keyspace) XXX_DiscardUnknown() {
	xxx_messageInfo_Keyspace.DiscardUnknown(m)
}

var xxx_messageInfo_Keyspace proto.InternalMessageInfo

func (m *Keyspace) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Keyspace) GetKeyspace() *topodata.Keyspace {
	if m != nil {
		return m.Keyspace
	}
	return nil
}

// TableMaterializeSttings contains the settings for one table.
type TableMaterializeSettings struct {
	TargetTable string `protobuf:"bytes,1,opt,name=target_table,json=targetTable,proto3" json:"target_table,omitempty"`
	// source_expression is a select statement.
	SourceExpression string `protobuf:"bytes,2,opt,name=source_expression,json=sourceExpression,proto3" json:"source_expression,omitempty"`
	// create_ddl contains the DDL to create the target table.
	// If empty, the target table must already exist.
	// if "copy", the target table DDL is the same as the source table.
	CreateDdl            string   `protobuf:"bytes,3,opt,name=create_ddl,json=createDdl,proto3" json:"create_ddl,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TableMaterializeSettings) Reset()         { *m = TableMaterializeSettings{} }
func (m *TableMaterializeSettings) String() string { return proto.CompactTextString(m) }
func (*TableMaterializeSettings) ProtoMessage()    {}
func (*TableMaterializeSettings) Descriptor() ([]byte, []int) {
	return fileDescriptor_f41247b323a1ab2e, []int{6}
}

func (m *TableMaterializeSettings) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TableMaterializeSettings.Unmarshal(m, b)
}
func (m *TableMaterializeSettings) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TableMaterializeSettings.Marshal(b, m, deterministic)
}
func (m *TableMaterializeSettings) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TableMaterializeSettings.Merge(m, src)
}
func (m *TableMaterializeSettings) XXX_Size() int {
	return xxx_messageInfo_TableMaterializeSettings.Size(m)
}
func (m *TableMaterializeSettings) XXX_DiscardUnknown() {
	xxx_messageInfo_TableMaterializeSettings.DiscardUnknown(m)
}

var xxx_messageInfo_TableMaterializeSettings proto.InternalMessageInfo

func (m *TableMaterializeSettings) GetTargetTable() string {
	if m != nil {
		return m.TargetTable
	}
	return ""
}

func (m *TableMaterializeSettings) GetSourceExpression() string {
	if m != nil {
		return m.SourceExpression
	}
	return ""
}

func (m *TableMaterializeSettings) GetCreateDdl() string {
	if m != nil {
		return m.CreateDdl
	}
	return ""
}

// MaterializeSettings contains the settings for the Materialize command.
type MaterializeSettings struct {
	// workflow is the name of the workflow.
	Workflow       string `protobuf:"bytes,1,opt,name=workflow,proto3" json:"workflow,omitempty"`
	SourceKeyspace string `protobuf:"bytes,2,opt,name=source_keyspace,json=sourceKeyspace,proto3" json:"source_keyspace,omitempty"`
	TargetKeyspace string `protobuf:"bytes,3,opt,name=target_keyspace,json=targetKeyspace,proto3" json:"target_keyspace,omitempty"`
	// stop_after_copy specifies if vreplication should be stopped after copying.
	StopAfterCopy bool                        `protobuf:"varint,4,opt,name=stop_after_copy,json=stopAfterCopy,proto3" json:"stop_after_copy,omitempty"`
	TableSettings []*TableMaterializeSettings `protobuf:"bytes,5,rep,name=table_settings,json=tableSettings,proto3" json:"table_settings,omitempty"`
	// optional parameters.
	Cell                 string   `protobuf:"bytes,6,opt,name=cell,proto3" json:"cell,omitempty"`
	TabletTypes          string   `protobuf:"bytes,7,opt,name=tablet_types,json=tabletTypes,proto3" json:"tablet_types,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MaterializeSettings) Reset()         { *m = MaterializeSettings{} }
func (m *MaterializeSettings) String() string { return proto.CompactTextString(m) }
func (*MaterializeSettings) ProtoMessage()    {}
func (*MaterializeSettings) Descriptor() ([]byte, []int) {
	return fileDescriptor_f41247b323a1ab2e, []int{7}
}

func (m *MaterializeSettings) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MaterializeSettings.Unmarshal(m, b)
}
func (m *MaterializeSettings) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MaterializeSettings.Marshal(b, m, deterministic)
}
func (m *MaterializeSettings) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MaterializeSettings.Merge(m, src)
}
func (m *MaterializeSettings) XXX_Size() int {
	return xxx_messageInfo_MaterializeSettings.Size(m)
}
func (m *MaterializeSettings) XXX_DiscardUnknown() {
	xxx_messageInfo_MaterializeSettings.DiscardUnknown(m)
}

var xxx_messageInfo_MaterializeSettings proto.InternalMessageInfo

func (m *MaterializeSettings) GetWorkflow() string {
	if m != nil {
		return m.Workflow
	}
	return ""
}

func (m *MaterializeSettings) GetSourceKeyspace() string {
	if m != nil {
		return m.SourceKeyspace
	}
	return ""
}

func (m *MaterializeSettings) GetTargetKeyspace() string {
	if m != nil {
		return m.TargetKeyspace
	}
	return ""
}

func (m *MaterializeSettings) GetStopAfterCopy() bool {
	if m != nil {
		return m.StopAfterCopy
	}
	return false
}

func (m *MaterializeSettings) GetTableSettings() []*TableMaterializeSettings {
	if m != nil {
		return m.TableSettings
	}
	return nil
}

func (m *MaterializeSettings) GetCell() string {
	if m != nil {
		return m.Cell
	}
	return ""
}

func (m *MaterializeSettings) GetTabletTypes() string {
	if m != nil {
		return m.TabletTypes
	}
	return ""
}

func init() {
	proto.RegisterType((*ExecuteVtctlCommandRequest)(nil), "vtctldata.ExecuteVtctlCommandRequest")
	proto.RegisterType((*ExecuteVtctlCommandResponse)(nil), "vtctldata.ExecuteVtctlCommandResponse")
	proto.RegisterType((*GetKeyspacesRequest)(nil), "vtctldata.GetKeyspacesRequest")
	proto.RegisterType((*GetKeyspacesResponse)(nil), "vtctldata.GetKeyspacesResponse")
	proto.RegisterType((*GetKeyspaceRequest)(nil), "vtctldata.GetKeyspaceRequest")
	proto.RegisterType((*Keyspace)(nil), "vtctldata.Keyspace")
	proto.RegisterType((*TableMaterializeSettings)(nil), "vtctldata.TableMaterializeSettings")
	proto.RegisterType((*MaterializeSettings)(nil), "vtctldata.MaterializeSettings")
}

func init() { proto.RegisterFile("vtctldata.proto", fileDescriptor_f41247b323a1ab2e) }

var fileDescriptor_f41247b323a1ab2e = []byte{
	// 505 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x93, 0xd1, 0x8e, 0xd3, 0x3c,
	0x10, 0x85, 0x95, 0xb6, 0xbb, 0x7f, 0x33, 0xfd, 0x9b, 0x82, 0x0b, 0x52, 0x54, 0x84, 0x54, 0x02,
	0xec, 0x56, 0x42, 0x4a, 0x61, 0x79, 0x02, 0x28, 0x15, 0x02, 0x04, 0x17, 0xa1, 0x02, 0x89, 0x9b,
	0xc8, 0x4d, 0x67, 0xab, 0x68, 0xdd, 0x38, 0xc4, 0xd3, 0xee, 0x96, 0x37, 0xe0, 0x65, 0x78, 0x46,
	0x14, 0x3b, 0x76, 0xbb, 0xd2, 0x72, 0xe7, 0x7c, 0x3e, 0x33, 0x73, 0xe6, 0x58, 0x81, 0xc1, 0x8e,
	0x32, 0x12, 0x2b, 0x4e, 0x3c, 0x2e, 0x2b, 0x49, 0x92, 0xf9, 0x0e, 0x8c, 0xfa, 0x42, 0xae, 0xb7,
	0x94, 0x0b, 0x73, 0x33, 0x0a, 0x48, 0x96, 0xf2, 0xa0, 0x8c, 0xbe, 0xc3, 0x68, 0x7e, 0x83, 0xd9,
	0x96, 0xf0, 0x5b, 0x5d, 0x32, 0x93, 0x9b, 0x0d, 0x2f, 0x56, 0x09, 0xfe, 0xdc, 0xa2, 0x22, 0xc6,
	0xa0, 0xc3, 0xab, 0xb5, 0x0a, 0xbd, 0x71, 0x7b, 0xe2, 0x27, 0xfa, 0xcc, 0x9e, 0x43, 0xc0, 0x33,
	0xca, 0x65, 0x91, 0x52, 0xbe, 0x41, 0xb9, 0xa5, 0xb0, 0x35, 0xf6, 0x26, 0xed, 0xa4, 0x6f, 0xe8,
	0xc2, 0xc0, 0x68, 0x06, 0x8f, 0xee, 0x6c, 0xac, 0x4a, 0x59, 0x28, 0x64, 0xcf, 0xe0, 0x04, 0x77,
	0x58, 0x50, 0xe8, 0x8d, 0xbd, 0x49, 0xef, 0x22, 0x88, 0xad, 0xcd, 0x79, 0x4d, 0x13, 0x73, 0x19,
	0x3d, 0x84, 0xe1, 0x7b, 0xa4, 0x4f, 0xb8, 0x57, 0x25, 0xcf, 0x50, 0x35, 0xb6, 0xa2, 0x0f, 0xf0,
	0xe0, 0x36, 0x6e, 0x9a, 0xbe, 0x02, 0xff, 0xca, 0x42, 0xed, 0xb9, 0x77, 0x31, 0x8c, 0x0f, 0xd9,
	0xd8, 0x82, 0xe4, 0xa0, 0x8a, 0x5e, 0x02, 0x3b, 0x6a, 0x65, 0xf7, 0x1e, 0x41, 0xd7, 0x4a, 0xb4,
	0x41, 0x3f, 0x71, 0xdf, 0xd1, 0x17, 0xe8, 0x5a, 0x79, 0x9d, 0x4f, 0xc1, 0x37, 0x56, 0xa3, 0xcf,
	0x2c, 0x3e, 0xaa, 0x6d, 0xe9, 0xe5, 0x58, 0xec, 0x42, 0x77, 0x83, 0x0e, 0xfd, 0x7e, 0x7b, 0x10,
	0x2e, 0xf8, 0x52, 0xe0, 0x67, 0x4e, 0x58, 0xe5, 0x5c, 0xe4, 0xbf, 0xf0, 0x2b, 0x12, 0xe5, 0xc5,
	0x5a, 0xb1, 0x27, 0xf0, 0x3f, 0xf1, 0x6a, 0x8d, 0x94, 0x52, 0x2d, 0x69, 0x06, 0xf5, 0x0c, 0xd3,
	0x55, 0xec, 0x05, 0xdc, 0x57, 0x72, 0x5b, 0x65, 0x98, 0xe2, 0x4d, 0x59, 0xa1, 0x52, 0xb9, 0x2c,
	0xf4, 0x60, 0x3f, 0xb9, 0x67, 0x2e, 0xe6, 0x8e, 0xb3, 0xc7, 0x00, 0x59, 0x85, 0x9c, 0x30, 0x5d,
	0xad, 0x44, 0xd8, 0xd6, 0x2a, 0xdf, 0x90, 0x77, 0x2b, 0x11, 0xfd, 0x69, 0xc1, 0xf0, 0x2e, 0x1b,
	0x23, 0xe8, 0x5e, 0xcb, 0xea, 0xea, 0x52, 0xc8, 0x6b, 0x9b, 0x87, 0xfd, 0x66, 0xe7, 0x30, 0x68,
	0xe6, 0xdf, 0x5a, 0xdb, 0x4f, 0x02, 0x83, 0x5d, 0x58, 0xe7, 0x30, 0x68, 0x76, 0x71, 0x42, 0x63,
	0x20, 0x30, 0xd8, 0x09, 0xcf, 0x60, 0xa0, 0x48, 0x96, 0x29, 0xbf, 0x24, 0xac, 0xd2, 0x4c, 0x96,
	0xfb, 0xb0, 0x33, 0xf6, 0x26, 0xdd, 0xa4, 0x5f, 0xe3, 0x37, 0x35, 0x9d, 0xc9, 0x72, 0xcf, 0x3e,
	0x42, 0xa0, 0x53, 0x49, 0x55, 0xe3, 0x33, 0x3c, 0xd1, 0x6f, 0xfe, 0xf4, 0xe8, 0xcd, 0xff, 0x95,
	0x6c, 0xd2, 0xd7, 0xa5, 0x6e, 0x43, 0x06, 0x9d, 0x0c, 0x85, 0x08, 0x4f, 0xcd, 0x4b, 0xd6, 0x67,
	0x13, 0xfe, 0x52, 0xd4, 0xe1, 0xef, 0x4b, 0x54, 0xe1, 0x7f, 0x36, 0xfc, 0x9a, 0x2d, 0x6a, 0xf4,
	0x76, 0xf2, 0xe3, 0x6c, 0x97, 0x13, 0x2a, 0x15, 0xe7, 0x72, 0x6a, 0x4e, 0xd3, 0xb5, 0x9c, 0xee,
	0x68, 0xaa, 0x7f, 0xaf, 0xa9, 0x33, 0xb2, 0x3c, 0xd5, 0xe0, 0xf5, 0xdf, 0x00, 0x00, 0x00, 0xff,
	0xff, 0xd5, 0x06, 0x15, 0x84, 0xac, 0x03, 0x00, 0x00,
}
