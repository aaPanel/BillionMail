package mail_boxes

import (
	"billionmail-core/internal/consts"
	"billionmail-core/internal/service/public"
	"context"
	"fmt"

	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/mail_boxes/v1"
	"billionmail-core/internal/service/mail_boxes"
)

func (c *ControllerV1) DeleteMailbox(ctx context.Context, req *v1.DeleteMailboxReq) (res *v1.DeleteMailboxRes, err error) {
	res = &v1.DeleteMailboxRes{}

	if len(req.Emails) == 0 {
		return nil, gerror.New("Email addresses cannot be empty")
	}

	// Do NOT filter out empty entries here. `mailbox.username` is the primary key
	// and an empty string is a legal key value, so rows with username = '' exist
	// (an import used to create them). Filtering them out made such a row
	// impossible to delete from the Mailboxes page: the request failed with
	// "No valid email addresses provided" before it reached the database. Deletion
	// matches the primary key exactly, so an empty identifier can only hit that row.
	emails := req.Emails

	affected, err := mail_boxes.DeleteBatch(ctx, emails)
	if err != nil {
		return nil, err
	}

	for _, email := range emails {
		_ = public.WriteLog(ctx, public.LogParams{
			Type: consts.LOGTYPE.Mailboxes,
			Log:  "Deleted mailbox:" + email + " successfully",
		})
	}

	if affected == 0 {
		res.SetSuccess("No mailboxes were deleted (they may not exist)")
	} else {
		res.SetSuccess(fmt.Sprintf("Successfully deleted %d mailbox(es)", affected))
	}

	return
}
