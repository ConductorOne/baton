// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0-devel
// 	protoc        (unknown)
// source: baton/v1/outputs.proto

package v1

import (
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ResourceDiff struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Created  []*v2.Resource `protobuf:"bytes,1,rep,name=created,proto3" json:"created,omitempty"`
	Deleted  []*v2.Resource `protobuf:"bytes,2,rep,name=deleted,proto3" json:"deleted,omitempty"`
	Modified []*v2.Resource `protobuf:"bytes,3,rep,name=modified,proto3" json:"modified,omitempty"`
}

func (x *ResourceDiff) Reset() {
	*x = ResourceDiff{}
	if protoimpl.UnsafeEnabled {
		mi := &file_baton_v1_outputs_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResourceDiff) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceDiff) ProtoMessage() {}

func (x *ResourceDiff) ProtoReflect() protoreflect.Message {
	mi := &file_baton_v1_outputs_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceDiff.ProtoReflect.Descriptor instead.
func (*ResourceDiff) Descriptor() ([]byte, []int) {
	return file_baton_v1_outputs_proto_rawDescGZIP(), []int{0}
}

func (x *ResourceDiff) GetCreated() []*v2.Resource {
	if x != nil {
		return x.Created
	}
	return nil
}

func (x *ResourceDiff) GetDeleted() []*v2.Resource {
	if x != nil {
		return x.Deleted
	}
	return nil
}

func (x *ResourceDiff) GetModified() []*v2.Resource {
	if x != nil {
		return x.Modified
	}
	return nil
}

type EntitlementDiff struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Created  []*v2.Entitlement `protobuf:"bytes,1,rep,name=created,proto3" json:"created,omitempty"`
	Deleted  []*v2.Entitlement `protobuf:"bytes,2,rep,name=deleted,proto3" json:"deleted,omitempty"`
	Modified []*v2.Entitlement `protobuf:"bytes,3,rep,name=modified,proto3" json:"modified,omitempty"`
}

func (x *EntitlementDiff) Reset() {
	*x = EntitlementDiff{}
	if protoimpl.UnsafeEnabled {
		mi := &file_baton_v1_outputs_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EntitlementDiff) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EntitlementDiff) ProtoMessage() {}

func (x *EntitlementDiff) ProtoReflect() protoreflect.Message {
	mi := &file_baton_v1_outputs_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EntitlementDiff.ProtoReflect.Descriptor instead.
func (*EntitlementDiff) Descriptor() ([]byte, []int) {
	return file_baton_v1_outputs_proto_rawDescGZIP(), []int{1}
}

func (x *EntitlementDiff) GetCreated() []*v2.Entitlement {
	if x != nil {
		return x.Created
	}
	return nil
}

func (x *EntitlementDiff) GetDeleted() []*v2.Entitlement {
	if x != nil {
		return x.Deleted
	}
	return nil
}

func (x *EntitlementDiff) GetModified() []*v2.Entitlement {
	if x != nil {
		return x.Modified
	}
	return nil
}

type GrantDiff struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Created  []*v2.Grant `protobuf:"bytes,1,rep,name=created,proto3" json:"created,omitempty"`
	Deleted  []*v2.Grant `protobuf:"bytes,2,rep,name=deleted,proto3" json:"deleted,omitempty"`
	Modified []*v2.Grant `protobuf:"bytes,3,rep,name=modified,proto3" json:"modified,omitempty"`
}

func (x *GrantDiff) Reset() {
	*x = GrantDiff{}
	if protoimpl.UnsafeEnabled {
		mi := &file_baton_v1_outputs_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GrantDiff) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GrantDiff) ProtoMessage() {}

func (x *GrantDiff) ProtoReflect() protoreflect.Message {
	mi := &file_baton_v1_outputs_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GrantDiff.ProtoReflect.Descriptor instead.
func (*GrantDiff) Descriptor() ([]byte, []int) {
	return file_baton_v1_outputs_proto_rawDescGZIP(), []int{2}
}

func (x *GrantDiff) GetCreated() []*v2.Grant {
	if x != nil {
		return x.Created
	}
	return nil
}

func (x *GrantDiff) GetDeleted() []*v2.Grant {
	if x != nil {
		return x.Deleted
	}
	return nil
}

func (x *GrantDiff) GetModified() []*v2.Grant {
	if x != nil {
		return x.Modified
	}
	return nil
}

type C1ZDiffOutput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Resources    *ResourceDiff    `protobuf:"bytes,1,opt,name=resources,proto3" json:"resources,omitempty"`
	Entitlements *EntitlementDiff `protobuf:"bytes,2,opt,name=entitlements,proto3" json:"entitlements,omitempty"`
	Grants       *GrantDiff       `protobuf:"bytes,3,opt,name=grants,proto3" json:"grants,omitempty"`
}

func (x *C1ZDiffOutput) Reset() {
	*x = C1ZDiffOutput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_baton_v1_outputs_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *C1ZDiffOutput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*C1ZDiffOutput) ProtoMessage() {}

func (x *C1ZDiffOutput) ProtoReflect() protoreflect.Message {
	mi := &file_baton_v1_outputs_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use C1ZDiffOutput.ProtoReflect.Descriptor instead.
func (*C1ZDiffOutput) Descriptor() ([]byte, []int) {
	return file_baton_v1_outputs_proto_rawDescGZIP(), []int{3}
}

func (x *C1ZDiffOutput) GetResources() *ResourceDiff {
	if x != nil {
		return x.Resources
	}
	return nil
}

func (x *C1ZDiffOutput) GetEntitlements() *EntitlementDiff {
	if x != nil {
		return x.Entitlements
	}
	return nil
}

func (x *C1ZDiffOutput) GetGrants() *GrantDiff {
	if x != nil {
		return x.Grants
	}
	return nil
}

type ResourceTypeOutput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ResourceType *v2.ResourceType `protobuf:"bytes,1,opt,name=resource_type,json=resourceType,proto3" json:"resource_type,omitempty"`
}

func (x *ResourceTypeOutput) Reset() {
	*x = ResourceTypeOutput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_baton_v1_outputs_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResourceTypeOutput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceTypeOutput) ProtoMessage() {}

func (x *ResourceTypeOutput) ProtoReflect() protoreflect.Message {
	mi := &file_baton_v1_outputs_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceTypeOutput.ProtoReflect.Descriptor instead.
func (*ResourceTypeOutput) Descriptor() ([]byte, []int) {
	return file_baton_v1_outputs_proto_rawDescGZIP(), []int{4}
}

func (x *ResourceTypeOutput) GetResourceType() *v2.ResourceType {
	if x != nil {
		return x.ResourceType
	}
	return nil
}

type ResourceOutput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Resource     *v2.Resource     `protobuf:"bytes,1,opt,name=resource,proto3" json:"resource,omitempty"`
	ResourceType *v2.ResourceType `protobuf:"bytes,2,opt,name=resource_type,json=resourceType,proto3" json:"resource_type,omitempty"`
	Parent       *v2.Resource     `protobuf:"bytes,3,opt,name=parent,proto3" json:"parent,omitempty"`
}

func (x *ResourceOutput) Reset() {
	*x = ResourceOutput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_baton_v1_outputs_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResourceOutput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceOutput) ProtoMessage() {}

func (x *ResourceOutput) ProtoReflect() protoreflect.Message {
	mi := &file_baton_v1_outputs_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceOutput.ProtoReflect.Descriptor instead.
func (*ResourceOutput) Descriptor() ([]byte, []int) {
	return file_baton_v1_outputs_proto_rawDescGZIP(), []int{5}
}

func (x *ResourceOutput) GetResource() *v2.Resource {
	if x != nil {
		return x.Resource
	}
	return nil
}

func (x *ResourceOutput) GetResourceType() *v2.ResourceType {
	if x != nil {
		return x.ResourceType
	}
	return nil
}

func (x *ResourceOutput) GetParent() *v2.Resource {
	if x != nil {
		return x.Parent
	}
	return nil
}

type EntitlementOutput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Entitlement  *v2.Entitlement  `protobuf:"bytes,1,opt,name=entitlement,proto3" json:"entitlement,omitempty"`
	Resource     *v2.Resource     `protobuf:"bytes,2,opt,name=resource,proto3" json:"resource,omitempty"`
	ResourceType *v2.ResourceType `protobuf:"bytes,3,opt,name=resource_type,json=resourceType,proto3" json:"resource_type,omitempty"`
}

func (x *EntitlementOutput) Reset() {
	*x = EntitlementOutput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_baton_v1_outputs_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EntitlementOutput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EntitlementOutput) ProtoMessage() {}

func (x *EntitlementOutput) ProtoReflect() protoreflect.Message {
	mi := &file_baton_v1_outputs_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EntitlementOutput.ProtoReflect.Descriptor instead.
func (*EntitlementOutput) Descriptor() ([]byte, []int) {
	return file_baton_v1_outputs_proto_rawDescGZIP(), []int{6}
}

func (x *EntitlementOutput) GetEntitlement() *v2.Entitlement {
	if x != nil {
		return x.Entitlement
	}
	return nil
}

func (x *EntitlementOutput) GetResource() *v2.Resource {
	if x != nil {
		return x.Resource
	}
	return nil
}

func (x *EntitlementOutput) GetResourceType() *v2.ResourceType {
	if x != nil {
		return x.ResourceType
	}
	return nil
}

type GrantOutput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Grant        *v2.Grant        `protobuf:"bytes,1,opt,name=grant,proto3" json:"grant,omitempty"`
	Entitlement  *v2.Entitlement  `protobuf:"bytes,2,opt,name=entitlement,proto3" json:"entitlement,omitempty"`
	Resource     *v2.Resource     `protobuf:"bytes,3,opt,name=resource,proto3" json:"resource,omitempty"`
	ResourceType *v2.ResourceType `protobuf:"bytes,4,opt,name=resource_type,json=resourceType,proto3" json:"resource_type,omitempty"`
	Principal    *v2.Resource     `protobuf:"bytes,5,opt,name=principal,proto3" json:"principal,omitempty"`
}

func (x *GrantOutput) Reset() {
	*x = GrantOutput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_baton_v1_outputs_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GrantOutput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GrantOutput) ProtoMessage() {}

func (x *GrantOutput) ProtoReflect() protoreflect.Message {
	mi := &file_baton_v1_outputs_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GrantOutput.ProtoReflect.Descriptor instead.
func (*GrantOutput) Descriptor() ([]byte, []int) {
	return file_baton_v1_outputs_proto_rawDescGZIP(), []int{7}
}

func (x *GrantOutput) GetGrant() *v2.Grant {
	if x != nil {
		return x.Grant
	}
	return nil
}

func (x *GrantOutput) GetEntitlement() *v2.Entitlement {
	if x != nil {
		return x.Entitlement
	}
	return nil
}

func (x *GrantOutput) GetResource() *v2.Resource {
	if x != nil {
		return x.Resource
	}
	return nil
}

func (x *GrantOutput) GetResourceType() *v2.ResourceType {
	if x != nil {
		return x.ResourceType
	}
	return nil
}

func (x *GrantOutput) GetPrincipal() *v2.Resource {
	if x != nil {
		return x.Principal
	}
	return nil
}

type ResourceTypeListOutput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ResourceTypes []*ResourceTypeOutput `protobuf:"bytes,1,rep,name=resource_types,json=resourceTypes,proto3" json:"resource_types,omitempty"`
}

func (x *ResourceTypeListOutput) Reset() {
	*x = ResourceTypeListOutput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_baton_v1_outputs_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResourceTypeListOutput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceTypeListOutput) ProtoMessage() {}

func (x *ResourceTypeListOutput) ProtoReflect() protoreflect.Message {
	mi := &file_baton_v1_outputs_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceTypeListOutput.ProtoReflect.Descriptor instead.
func (*ResourceTypeListOutput) Descriptor() ([]byte, []int) {
	return file_baton_v1_outputs_proto_rawDescGZIP(), []int{8}
}

func (x *ResourceTypeListOutput) GetResourceTypes() []*ResourceTypeOutput {
	if x != nil {
		return x.ResourceTypes
	}
	return nil
}

type ResourceListOutput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Resources []*ResourceOutput `protobuf:"bytes,1,rep,name=resources,proto3" json:"resources,omitempty"`
}

func (x *ResourceListOutput) Reset() {
	*x = ResourceListOutput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_baton_v1_outputs_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResourceListOutput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceListOutput) ProtoMessage() {}

func (x *ResourceListOutput) ProtoReflect() protoreflect.Message {
	mi := &file_baton_v1_outputs_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceListOutput.ProtoReflect.Descriptor instead.
func (*ResourceListOutput) Descriptor() ([]byte, []int) {
	return file_baton_v1_outputs_proto_rawDescGZIP(), []int{9}
}

func (x *ResourceListOutput) GetResources() []*ResourceOutput {
	if x != nil {
		return x.Resources
	}
	return nil
}

type EntitlementListOutput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Entitlements []*EntitlementOutput `protobuf:"bytes,1,rep,name=entitlements,proto3" json:"entitlements,omitempty"`
}

func (x *EntitlementListOutput) Reset() {
	*x = EntitlementListOutput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_baton_v1_outputs_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EntitlementListOutput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EntitlementListOutput) ProtoMessage() {}

func (x *EntitlementListOutput) ProtoReflect() protoreflect.Message {
	mi := &file_baton_v1_outputs_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EntitlementListOutput.ProtoReflect.Descriptor instead.
func (*EntitlementListOutput) Descriptor() ([]byte, []int) {
	return file_baton_v1_outputs_proto_rawDescGZIP(), []int{10}
}

func (x *EntitlementListOutput) GetEntitlements() []*EntitlementOutput {
	if x != nil {
		return x.Entitlements
	}
	return nil
}

type GrantListOutput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Grants []*GrantOutput `protobuf:"bytes,1,rep,name=grants,proto3" json:"grants,omitempty"`
}

func (x *GrantListOutput) Reset() {
	*x = GrantListOutput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_baton_v1_outputs_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GrantListOutput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GrantListOutput) ProtoMessage() {}

func (x *GrantListOutput) ProtoReflect() protoreflect.Message {
	mi := &file_baton_v1_outputs_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GrantListOutput.ProtoReflect.Descriptor instead.
func (*GrantListOutput) Descriptor() ([]byte, []int) {
	return file_baton_v1_outputs_proto_rawDescGZIP(), []int{11}
}

func (x *GrantListOutput) GetGrants() []*GrantOutput {
	if x != nil {
		return x.Grants
	}
	return nil
}

var File_baton_v1_outputs_proto protoreflect.FileDescriptor

var file_baton_v1_outputs_proto_rawDesc = []byte{
	0x0a, 0x16, 0x62, 0x61, 0x74, 0x6f, 0x6e, 0x2f, 0x76, 0x31, 0x2f, 0x6f, 0x75, 0x74, 0x70, 0x75,
	0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x62, 0x61, 0x74, 0x6f, 0x6e, 0x2e,
	0x76, 0x31, 0x1a, 0x1e, 0x63, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72,
	0x2f, 0x76, 0x32, 0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x21, 0x63, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72,
	0x2f, 0x76, 0x32, 0x2f, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x63, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x6f, 0x72, 0x2f, 0x76, 0x32, 0x2f, 0x67, 0x72, 0x61, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0xaf, 0x01, 0x0a, 0x0c, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x44,
	0x69, 0x66, 0x66, 0x12, 0x33, 0x0a, 0x07, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x63, 0x31, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x52,
	0x07, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x12, 0x33, 0x0a, 0x07, 0x64, 0x65, 0x6c, 0x65,
	0x74, 0x65, 0x64, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x63, 0x31, 0x2e, 0x63,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x52, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x52, 0x07, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x64, 0x12, 0x35, 0x0a,
	0x08, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x19, 0x2e, 0x63, 0x31, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x76,
	0x32, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x52, 0x08, 0x6d, 0x6f, 0x64, 0x69,
	0x66, 0x69, 0x65, 0x64, 0x22, 0xbb, 0x01, 0x0a, 0x0f, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x6c, 0x65,
	0x6d, 0x65, 0x6e, 0x74, 0x44, 0x69, 0x66, 0x66, 0x12, 0x36, 0x0a, 0x07, 0x63, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x64, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x63, 0x31, 0x2e, 0x63,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x45, 0x6e, 0x74, 0x69,
	0x74, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x07, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64,
	0x12, 0x36, 0x0a, 0x07, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x64, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x1c, 0x2e, 0x63, 0x31, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72,
	0x2e, 0x76, 0x32, 0x2e, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x52,
	0x07, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x64, 0x12, 0x38, 0x0a, 0x08, 0x6d, 0x6f, 0x64, 0x69,
	0x66, 0x69, 0x65, 0x64, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x63, 0x31, 0x2e,
	0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x45, 0x6e, 0x74,
	0x69, 0x74, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x08, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69,
	0x65, 0x64, 0x22, 0xa3, 0x01, 0x0a, 0x09, 0x47, 0x72, 0x61, 0x6e, 0x74, 0x44, 0x69, 0x66, 0x66,
	0x12, 0x30, 0x0a, 0x07, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x16, 0x2e, 0x63, 0x31, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72,
	0x2e, 0x76, 0x32, 0x2e, 0x47, 0x72, 0x61, 0x6e, 0x74, 0x52, 0x07, 0x63, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x64, 0x12, 0x30, 0x0a, 0x07, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x64, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x31, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74,
	0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x47, 0x72, 0x61, 0x6e, 0x74, 0x52, 0x07, 0x64, 0x65, 0x6c,
	0x65, 0x74, 0x65, 0x64, 0x12, 0x32, 0x0a, 0x08, 0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64,
	0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x31, 0x2e, 0x63, 0x6f, 0x6e, 0x6e,
	0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x47, 0x72, 0x61, 0x6e, 0x74, 0x52, 0x08,
	0x6d, 0x6f, 0x64, 0x69, 0x66, 0x69, 0x65, 0x64, 0x22, 0xb1, 0x01, 0x0a, 0x0d, 0x43, 0x31, 0x5a,
	0x44, 0x69, 0x66, 0x66, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x34, 0x0a, 0x09, 0x72, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e,
	0x62, 0x61, 0x74, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x44, 0x69, 0x66, 0x66, 0x52, 0x09, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73,
	0x12, 0x3d, 0x0a, 0x0c, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x73,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x62, 0x61, 0x74, 0x6f, 0x6e, 0x2e, 0x76,
	0x31, 0x2e, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x44, 0x69, 0x66,
	0x66, 0x52, 0x0c, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x12,
	0x2b, 0x0a, 0x06, 0x67, 0x72, 0x61, 0x6e, 0x74, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x13, 0x2e, 0x62, 0x61, 0x74, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x72, 0x61, 0x6e, 0x74,
	0x44, 0x69, 0x66, 0x66, 0x52, 0x06, 0x67, 0x72, 0x61, 0x6e, 0x74, 0x73, 0x22, 0x58, 0x0a, 0x12,
	0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x4f, 0x75, 0x74, 0x70,
	0x75, 0x74, 0x12, 0x42, 0x0a, 0x0d, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x74,
	0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x63, 0x31, 0x2e, 0x63,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x52, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0c, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x22, 0xbe, 0x01, 0x0a, 0x0e, 0x52, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x35, 0x0a, 0x08, 0x72, 0x65, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x63, 0x31,
	0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x52, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x52, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x12, 0x42, 0x0a, 0x0d, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x74, 0x79, 0x70,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x63, 0x31, 0x2e, 0x63, 0x6f, 0x6e,
	0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0c, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x54, 0x79, 0x70, 0x65, 0x12, 0x31, 0x0a, 0x06, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x63, 0x31, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x52,
	0x06, 0x70, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x22, 0xce, 0x01, 0x0a, 0x11, 0x45, 0x6e, 0x74, 0x69,
	0x74, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x3e, 0x0a,
	0x0b, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x63, 0x31, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f,
	0x72, 0x2e, 0x76, 0x32, 0x2e, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74,
	0x52, 0x0b, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x35, 0x0a,
	0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x19, 0x2e, 0x63, 0x31, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x76,
	0x32, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x52, 0x08, 0x72, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x12, 0x42, 0x0a, 0x0d, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x63, 0x31,
	0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x52, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0c, 0x72, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x22, 0xaf, 0x02, 0x0a, 0x0b, 0x47, 0x72, 0x61,
	0x6e, 0x74, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x2c, 0x0a, 0x05, 0x67, 0x72, 0x61, 0x6e,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x31, 0x2e, 0x63, 0x6f, 0x6e,
	0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x47, 0x72, 0x61, 0x6e, 0x74, 0x52,
	0x05, 0x67, 0x72, 0x61, 0x6e, 0x74, 0x12, 0x3e, 0x0a, 0x0b, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x6c,
	0x65, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x63, 0x31,
	0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x45, 0x6e,
	0x74, 0x69, 0x74, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x0b, 0x65, 0x6e, 0x74, 0x69, 0x74,
	0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x35, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x63, 0x31, 0x2e, 0x63, 0x6f,
	0x6e, 0x6e, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x52, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x42, 0x0a,
	0x0d, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x63, 0x31, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x54,
	0x79, 0x70, 0x65, 0x52, 0x0c, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x54, 0x79, 0x70,
	0x65, 0x12, 0x37, 0x0a, 0x09, 0x70, 0x72, 0x69, 0x6e, 0x63, 0x69, 0x70, 0x61, 0x6c, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x63, 0x31, 0x2e, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x6f, 0x72, 0x2e, 0x76, 0x32, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x52,
	0x09, 0x70, 0x72, 0x69, 0x6e, 0x63, 0x69, 0x70, 0x61, 0x6c, 0x22, 0x5d, 0x0a, 0x16, 0x52, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x4c, 0x69, 0x73, 0x74, 0x4f, 0x75,
	0x74, 0x70, 0x75, 0x74, 0x12, 0x43, 0x0a, 0x0e, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x5f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x62,
	0x61, 0x74, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x54, 0x79, 0x70, 0x65, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x52, 0x0d, 0x72, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x54, 0x79, 0x70, 0x65, 0x73, 0x22, 0x4c, 0x0a, 0x12, 0x52, 0x65, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x4c, 0x69, 0x73, 0x74, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12,
	0x36, 0x0a, 0x09, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x18, 0x2e, 0x62, 0x61, 0x74, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x52, 0x09, 0x72, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x22, 0x58, 0x0a, 0x15, 0x45, 0x6e, 0x74, 0x69, 0x74,
	0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74,
	0x12, 0x3f, 0x0a, 0x0c, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x62, 0x61, 0x74, 0x6f, 0x6e, 0x2e, 0x76,
	0x31, 0x2e, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x4f, 0x75, 0x74,
	0x70, 0x75, 0x74, 0x52, 0x0c, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74,
	0x73, 0x22, 0x40, 0x0a, 0x0f, 0x47, 0x72, 0x61, 0x6e, 0x74, 0x4c, 0x69, 0x73, 0x74, 0x4f, 0x75,
	0x74, 0x70, 0x75, 0x74, 0x12, 0x2d, 0x0a, 0x06, 0x67, 0x72, 0x61, 0x6e, 0x74, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x62, 0x61, 0x74, 0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x2e,
	0x47, 0x72, 0x61, 0x6e, 0x74, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x52, 0x06, 0x67, 0x72, 0x61,
	0x6e, 0x74, 0x73, 0x42, 0x2f, 0x5a, 0x2d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x63, 0x6f, 0x6e, 0x64, 0x75, 0x63, 0x74, 0x6f, 0x72, 0x6f, 0x6e, 0x65, 0x2f, 0x62,
	0x61, 0x74, 0x6f, 0x6e, 0x2f, 0x70, 0x62, 0x2f, 0x62, 0x61, 0x74, 0x6f, 0x6e, 0x5f, 0x63, 0x6c,
	0x69, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_baton_v1_outputs_proto_rawDescOnce sync.Once
	file_baton_v1_outputs_proto_rawDescData = file_baton_v1_outputs_proto_rawDesc
)

func file_baton_v1_outputs_proto_rawDescGZIP() []byte {
	file_baton_v1_outputs_proto_rawDescOnce.Do(func() {
		file_baton_v1_outputs_proto_rawDescData = protoimpl.X.CompressGZIP(file_baton_v1_outputs_proto_rawDescData)
	})
	return file_baton_v1_outputs_proto_rawDescData
}

var file_baton_v1_outputs_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_baton_v1_outputs_proto_goTypes = []interface{}{
	(*ResourceDiff)(nil),           // 0: baton.v1.ResourceDiff
	(*EntitlementDiff)(nil),        // 1: baton.v1.EntitlementDiff
	(*GrantDiff)(nil),              // 2: baton.v1.GrantDiff
	(*C1ZDiffOutput)(nil),          // 3: baton.v1.C1ZDiffOutput
	(*ResourceTypeOutput)(nil),     // 4: baton.v1.ResourceTypeOutput
	(*ResourceOutput)(nil),         // 5: baton.v1.ResourceOutput
	(*EntitlementOutput)(nil),      // 6: baton.v1.EntitlementOutput
	(*GrantOutput)(nil),            // 7: baton.v1.GrantOutput
	(*ResourceTypeListOutput)(nil), // 8: baton.v1.ResourceTypeListOutput
	(*ResourceListOutput)(nil),     // 9: baton.v1.ResourceListOutput
	(*EntitlementListOutput)(nil),  // 10: baton.v1.EntitlementListOutput
	(*GrantListOutput)(nil),        // 11: baton.v1.GrantListOutput
	(*v2.Resource)(nil),            // 12: c1.connector.v2.Resource
	(*v2.Entitlement)(nil),         // 13: c1.connector.v2.Entitlement
	(*v2.Grant)(nil),               // 14: c1.connector.v2.Grant
	(*v2.ResourceType)(nil),        // 15: c1.connector.v2.ResourceType
}
var file_baton_v1_outputs_proto_depIdxs = []int32{
	12, // 0: baton.v1.ResourceDiff.created:type_name -> c1.connector.v2.Resource
	12, // 1: baton.v1.ResourceDiff.deleted:type_name -> c1.connector.v2.Resource
	12, // 2: baton.v1.ResourceDiff.modified:type_name -> c1.connector.v2.Resource
	13, // 3: baton.v1.EntitlementDiff.created:type_name -> c1.connector.v2.Entitlement
	13, // 4: baton.v1.EntitlementDiff.deleted:type_name -> c1.connector.v2.Entitlement
	13, // 5: baton.v1.EntitlementDiff.modified:type_name -> c1.connector.v2.Entitlement
	14, // 6: baton.v1.GrantDiff.created:type_name -> c1.connector.v2.Grant
	14, // 7: baton.v1.GrantDiff.deleted:type_name -> c1.connector.v2.Grant
	14, // 8: baton.v1.GrantDiff.modified:type_name -> c1.connector.v2.Grant
	0,  // 9: baton.v1.C1ZDiffOutput.resources:type_name -> baton.v1.ResourceDiff
	1,  // 10: baton.v1.C1ZDiffOutput.entitlements:type_name -> baton.v1.EntitlementDiff
	2,  // 11: baton.v1.C1ZDiffOutput.grants:type_name -> baton.v1.GrantDiff
	15, // 12: baton.v1.ResourceTypeOutput.resource_type:type_name -> c1.connector.v2.ResourceType
	12, // 13: baton.v1.ResourceOutput.resource:type_name -> c1.connector.v2.Resource
	15, // 14: baton.v1.ResourceOutput.resource_type:type_name -> c1.connector.v2.ResourceType
	12, // 15: baton.v1.ResourceOutput.parent:type_name -> c1.connector.v2.Resource
	13, // 16: baton.v1.EntitlementOutput.entitlement:type_name -> c1.connector.v2.Entitlement
	12, // 17: baton.v1.EntitlementOutput.resource:type_name -> c1.connector.v2.Resource
	15, // 18: baton.v1.EntitlementOutput.resource_type:type_name -> c1.connector.v2.ResourceType
	14, // 19: baton.v1.GrantOutput.grant:type_name -> c1.connector.v2.Grant
	13, // 20: baton.v1.GrantOutput.entitlement:type_name -> c1.connector.v2.Entitlement
	12, // 21: baton.v1.GrantOutput.resource:type_name -> c1.connector.v2.Resource
	15, // 22: baton.v1.GrantOutput.resource_type:type_name -> c1.connector.v2.ResourceType
	12, // 23: baton.v1.GrantOutput.principal:type_name -> c1.connector.v2.Resource
	4,  // 24: baton.v1.ResourceTypeListOutput.resource_types:type_name -> baton.v1.ResourceTypeOutput
	5,  // 25: baton.v1.ResourceListOutput.resources:type_name -> baton.v1.ResourceOutput
	6,  // 26: baton.v1.EntitlementListOutput.entitlements:type_name -> baton.v1.EntitlementOutput
	7,  // 27: baton.v1.GrantListOutput.grants:type_name -> baton.v1.GrantOutput
	28, // [28:28] is the sub-list for method output_type
	28, // [28:28] is the sub-list for method input_type
	28, // [28:28] is the sub-list for extension type_name
	28, // [28:28] is the sub-list for extension extendee
	0,  // [0:28] is the sub-list for field type_name
}

func init() { file_baton_v1_outputs_proto_init() }
func file_baton_v1_outputs_proto_init() {
	if File_baton_v1_outputs_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_baton_v1_outputs_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResourceDiff); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_baton_v1_outputs_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EntitlementDiff); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_baton_v1_outputs_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GrantDiff); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_baton_v1_outputs_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*C1ZDiffOutput); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_baton_v1_outputs_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResourceTypeOutput); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_baton_v1_outputs_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResourceOutput); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_baton_v1_outputs_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EntitlementOutput); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_baton_v1_outputs_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GrantOutput); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_baton_v1_outputs_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResourceTypeListOutput); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_baton_v1_outputs_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResourceListOutput); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_baton_v1_outputs_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EntitlementListOutput); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_baton_v1_outputs_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GrantListOutput); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_baton_v1_outputs_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_baton_v1_outputs_proto_goTypes,
		DependencyIndexes: file_baton_v1_outputs_proto_depIdxs,
		MessageInfos:      file_baton_v1_outputs_proto_msgTypes,
	}.Build()
	File_baton_v1_outputs_proto = out.File
	file_baton_v1_outputs_proto_rawDesc = nil
	file_baton_v1_outputs_proto_goTypes = nil
	file_baton_v1_outputs_proto_depIdxs = nil
}