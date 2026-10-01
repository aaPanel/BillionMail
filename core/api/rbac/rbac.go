// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package rbac

import (
	"context"

	"billionmail-core/api/rbac/v1"
)

type IRbacV1 interface {
	//AccountList(ctx context.Context, req *v1.AccountListReq) (res *v1.AccountListRes, err error)
	//AccountDetail(ctx context.Context, req *v1.AccountDetailReq) (res *v1.AccountDetailRes, err error)
	//AccountCreate(ctx context.Context, req *v1.AccountCreateReq) (res *v1.AccountCreateRes, err error)
	//AccountUpdate(ctx context.Context, req *v1.AccountUpdateReq) (res *v1.AccountUpdateRes, err error)
	//AccountPassword(ctx context.Context, req *v1.AccountPasswordReq) (res *v1.AccountPasswordRes, err error)
	//AccountDelete(ctx context.Context, req *v1.AccountDeleteReq) (res *v1.AccountDeleteRes, err error)
	Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error)
	AaPanelSSO(ctx context.Context, req *v1.AaPanelSSOReq) (res *v1.AaPanelSSORes, err error)
	Logout(ctx context.Context, req *v1.LogoutReq) (res *v1.LogoutRes, err error)
	RefreshToken(ctx context.Context, req *v1.RefreshTokenReq) (res *v1.RefreshTokenRes, err error)
	GetValidateCode(ctx context.Context, req *v1.GetValidateCodeReq) (res *v1.GetValidateCodeRes, err error)
	//CurrentUser(ctx context.Context, req *v1.CurrentUserReq) (res *v1.CurrentUserRes, err error)
	//PermissionList(ctx context.Context, req *v1.PermissionListReq) (res *v1.PermissionListRes, err error)
	//PermissionDetail(ctx context.Context, req *v1.PermissionDetailReq) (res *v1.PermissionDetailRes, err error)
	//PermissionCreate(ctx context.Context, req *v1.PermissionCreateReq) (res *v1.PermissionCreateRes, err error)
	//PermissionUpdate(ctx context.Context, req *v1.PermissionUpdateReq) (res *v1.PermissionUpdateRes, err error)
	//PermissionDelete(ctx context.Context, req *v1.PermissionDeleteReq) (res *v1.PermissionDeleteRes, err error)
	//PermissionCheck(ctx context.Context, req *v1.PermissionCheckReq) (res *v1.PermissionCheckRes, err error)
	//RoleList(ctx context.Context, req *v1.RoleListReq) (res *v1.RoleListRes, err error)
	//RoleDetail(ctx context.Context, req *v1.RoleDetailReq) (res *v1.RoleDetailRes, err error)
	//RoleCreate(ctx context.Context, req *v1.RoleCreateReq) (res *v1.RoleCreateRes, err error)
	//RoleUpdate(ctx context.Context, req *v1.RoleUpdateReq) (res *v1.RoleUpdateRes, err error)
	//RoleDelete(ctx context.Context, req *v1.RoleDeleteReq) (res *v1.RoleDeleteRes, err error)
}
