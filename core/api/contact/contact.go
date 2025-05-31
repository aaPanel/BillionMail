// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package contact

import (
	"context"

	"billionmail-core/api/contact/v1"
)

type IContactV1 interface {
	CreateGroup(ctx context.Context, req *v1.CreateGroupReq) (res *v1.CreateGroupRes, err error)
	ImportContacts(ctx context.Context, req *v1.ImportContactsReq) (res *v1.ImportContactsRes, err error)
	ExportContacts(ctx context.Context, req *v1.ExportContactsReq) (res *v1.ExportContactsRes, err error)
	DeleteGroup(ctx context.Context, req *v1.DeleteGroupReq) (res *v1.DeleteGroupRes, err error)
	UpdateGroup(ctx context.Context, req *v1.UpdateGroupReq) (res *v1.UpdateGroupRes, err error)
	ListGroups(ctx context.Context, req *v1.ListGroupsReq) (res *v1.ListGroupsRes, err error)
	ListContacts(ctx context.Context, req *v1.ListContactsReq) (res *v1.ListContactsRes, err error)
	ListContactsGroups(ctx context.Context, req *v1.ListContactsGroupsReq) (res *v1.ListContactsGroupsRes, err error)
	MergeContactsGroups(ctx context.Context, req *v1.MergeContactsGroupsReq) (res *v1.MergeContactsGroupsRes, err error)
	DeleteContacts(ctx context.Context, req *v1.DeleteContactsReq) (res *v1.DeleteContactsRes, err error)
	UpdateContactsGroup(ctx context.Context, req *v1.UpdateContactsGroupReq) (res *v1.UpdateContactsGroupRes, err error)
	GetContactsTrend(ctx context.Context, req *v1.GetContactsTrendReq) (res *v1.GetContactsTrendRes, err error)
	GetGroupContactCount(ctx context.Context, req *v1.GetGroupContactCountReq) (res *v1.GetGroupContactCountRes, err error)
	EditContacts(ctx context.Context, req *v1.EditContactsReq) (res *v1.EditContactsRes, err error)
}
