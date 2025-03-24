// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package account

import (
	"context"

	"billionmail-core/api/account/v1"
)

type IAccountV1 interface {
	Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error)
	Logout(ctx context.Context, req *v1.LogoutReq) (res *v1.LogoutRes, err error)
	GetAccounts(ctx context.Context, req *v1.GetAccountsReq) (res *v1.GetAccountsRes, err error)
	UpdateAccount(ctx context.Context, req *v1.UpdateAccountReq) (res *v1.UpdateAccountRes, err error)
	DeleteAccount(ctx context.Context, req *v1.DeleteAccountReq) (res *v1.DeleteAccountRes, err error)
	AddAccount(ctx context.Context, req *v1.AddAccountReq) (res *v1.AddAccountRes, err error)
}
