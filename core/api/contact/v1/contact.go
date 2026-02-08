package v1

import (
	"billionmail-core/utility/types/api_v1"

	"github.com/gogf/gf/v2/frame/g"
)

type ContactGroup struct {
	Id          int    `json:"id"          dc:"Group ID"`
	Name        string `json:"name"        dc:"Group Name"`
	Description string `json:"description" dc:"Description"`
	CreateTime  int    `json:"create_time" dc:"Create Time"`
	UpdateTime  int    `json:"update_time" dc:"Update Time"`
	Token       string `json:"token"      dc:"Subscription Token"`
	DoubleOptin int    `json:"double_optin" dc:"Double Opt-in Status(0: Single Opt-in 1: Double Opt-in)"`

	WelcomeHtml string `json:"welcome_mail_html" dc:"Welcome Email Html"`
	WelcomeDrag string `json:"welcome_mail_drag" dc:"Welcome Email Drag"`

	WelcomeSubject   string `json:"welcome_subject" dc:"Welcome Email Subject"`
	SendWelcomeEmail int    `json:"send_welcome_email" dc:"Whether to send a welcome email"`
	ConfirmSubject   string `json:"confirm_subject" dc:"Confirmation Email Subject"`

	ConfirmHtml string `json:"confirm_mail_html" dc:"Confirmation Email Html"`
	ConfirmDrag string `json:"confirm_mail_drag" dc:"Confirmation Email Drag"`

	SuccessUrl    string `json:"success_url" dc:"Success URL"`
	ConfirmUrl    string `json:"confirm_url" dc:"Confirmation URL"`
	AlreadyUrl    string `json:"already_url" dc:"Already Subscribed URL"`
	SubscribeForm string `json:"subscribe_form" dc:"Subscription Form HTML"`
	// New unsubscribe related fields
	UnsubscribeMailHtml    string `json:"unsubscribe_mail_html" dc:"Unsubscribe Email Html"`
	UnsubscribeMailDrag    string `json:"unsubscribe_mail_drag" dc:"Unsubscribe EmailDrag"`
	UnsubscribeSubject     string `json:"unsubscribe_subject" dc:"Unsubscribe Email Subject"`
	UnsubscribeRedirectUrl string `json:"unsubscribe_redirect_url" dc:"Unsubscribe Success RedirectURL"`
	SendUnsubscribeEmail   int    `json:"send_unsubscribe_email" dc:"Whether to send unsubscribe email"`
}

type ContactGroupInfo struct {
	ContactGroup
	TotalCount       int    `json:"total_count" dc:"Total Contacts"`
	ActiveCount      int    `json:"active_count" dc:"Subscribed Contacts"`
	UnsubscribeCount int    `json:"unsubscribe_count" dc:"Unsubscribed Contacts"`
	SubscribeLink    string `json:"subscribe_link" dc:"Subscription Link"`
	Sender           string `json:"sender" dc:"Sender"`
}

type GroupInfo struct {
	Id   int    `json:"id"   dc:"Group ID"`
	Name string `json:"name" dc:"Group Name"`
}

type TagInfo struct {
	Id         int    `json:"id"          dc:"Tag ID"`
	Name       string `json:"name"        dc:"Tag Name"`
	CreateTime int    `json:"create_time" dc:"Create Time"`
}

type Contact struct {
	Id           int               `json:"id"          dc:"Contact ID"`
	Email        string            `json:"email"       dc:"Email Address"`
	GroupId      int               `json:"group_id"    dc:"Group ID"`
	Active       int               `json:"active"      dc:"Status(1:Subscribed 0:Unsubscribed)"`
	TaskId       int               `json:"task_id"     dc:"Bulk Mail Task ID"`
	CreateTime   int               `json:"create_time" dc:"Create Time"`
	Groups       []GroupInfo       `json:"groups"      dc:"Contact Groups"`
	Status       int               `json:"status"      dc:"Status( 1:Confirmed   0:Unconfirmed)"`
	GroupName    string            `json:"group_name"      dc:"Contact Group Name"`
	Attribs      map[string]string `json:"attribs"`
	LastActiveAt int               `json:"last_active_at" dc:"Last Active At"`
	Tags         []TagInfo         `json:"tags"        dc:"Contact Tags"`
}

type CreateGroupReq struct {
	g.Meta        `path:"/contact/group/create" method:"post" tags:"Contact" summary:"Create new contact group"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	GroupId       int    `json:"group_id" dc:"GroupID to import contacts into"`
	Name          string `json:"name"  dc:"name"`
	Description   string `json:"description" dc:"description"`
	FileData      string `json:"file_data" dc:"file data"`
	FileType      string `json:"file_type" v:"in:csv,excel,txt" dc:"file type, default:txt"`
	CreateType    int    `json:"create_type" v:"required|in:1,2,3" dc:"Create type (1: Create group only 2: Create new group and import contacts 3: Import files into existing groups)"`
	DoubleOptin   int    `json:"double_optin" v:"in:0,1" dc:"Data confirmation ( 0: Single Opt-in, 1: Double Opt-in) Default:0"`
}

type CreateGroupRes struct {
	api_v1.StandardRes
	Data struct {
		GroupId int `json:"group_id" dc:"Created Group ID"`
	} `json:"data"`
}

// ImportContactsReq Import contacts request
type ImportContactsReq struct {
	g.Meta        `path:"/contact/group/import" method:"post" tags:"Contact" summary:"Import contacts"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	GroupIds      []int  `json:"group_ids" v:"required" dc:"Group IDs to import contacts into"`
	FileData      string `json:"file_data" dc:"file data"`
	FileType      string `json:"file_type" v:"in:csv,excel,txt" dc:"file type, default:txt"`
	Contacts      string `json:"contacts" dc:"paste content, one email per line, optional with subscription status, format: email[,status]"`
	ImportType    int    `json:"import_type" v:"required|in:1,2" dc:"Import type (1: Upload file content 2: Paste content)"`
	DefaultActive int    `json:"default_active" v:"in:0,1" dc:"Default subscription status (1: Subscribed 0: Unsubscribed) Default:1"`
	Status        int    `json:"status" v:"in:0,1" dc:"Data confirmation (1:Confirmed   0:Unconfirmed) Default:0"`
	Overwrite     int    `json:"overwrite" v:"in:0,1" dc:"Overwrite attribs, subscription status of existing  (1:overwrite   0:unoverwrite) Default:0"`
}

// ImportContactsRes Import contacts response
type ImportContactsRes struct {
	api_v1.StandardRes
	Data struct {
		ImportedCount int `json:"imported_count" dc:"Number of successfully imported contacts"`
	} `json:"data"`
}

type ExportContactsReq struct {
	g.Meta          `path:"/contact/group/export" method:"post" tags:"Contact" summary:"Export contacts"`
	Authorization   string `json:"authorization" dc:"Authorization" in:"header"`
	GroupIds        []int  `json:"group_ids" v:"required" dc:"Group IDs"`
	ExportType      int    `json:"export_type" v:"required|in:1,2" dc:"Export Type(1:Merged 2:Separate)"`
	Format          string `json:"format" v:"required|in:csv,excel,txt" dc:"Export Format"`
	IncludeInactive bool   `json:"include_unsubscribe" dc:"Include Unsubscribed(Default:true)"`
}

type ExportContactsRes struct {
	api_v1.StandardRes
}

type DeleteGroupReq struct {
	g.Meta        `path:"/contact/group/delete" method:"post" tags:"Contact" summary:"Delete contact group"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	GroupIds      []int  `json:"group_ids" v:"required" dc:"Group IDs"`
}

type DeleteGroupRes struct {
	api_v1.StandardRes
	Data struct {
		SuccessCount int `json:"success_count" dc:"Successfully Deleted Groups"`
		FailedCount  int `json:"failed_count" dc:"Failed Deleted Groups"`
	} `json:"data"`
}

type UpdateGroupReq struct {
	g.Meta        `path:"/contact/group/update" method:"post" tags:"Contact" summary:"Update contact group"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	GroupId       int    `json:"group_id" v:"required" dc:"Group ID"`
	Name          string `json:"name" dc:"Group Name"`
	Description   string `json:"description" dc:"Description"`
	DoubleOptin   int    `json:"double_optin" v:"in:0,1" dc:"Double Opt-in Status(0: Single Opt-in 1: Double Opt-in) Default:0"`
	WelcomeHtml   string `json:"welcome_mail_html" dc:"Welcome Email Html"`
	WelcomeDrag   string `json:"welcome_mail_drag" dc:"Welcome Email Drag"`
	ConfirmHtml   string `json:"confirm_mail_html" dc:"Confirmation Email Html"`
	ConfirmDrag   string `json:"confirm_mail_drag" dc:"Confirmation Email Drag"`
	SuccessUrl    string `json:"success_url" dc:"Success URL"`
	ConfirmUrl    string `json:"confirm_url" dc:"Confirmation URL"`
	AlreadyUrl    string `json:"already_url" dc:"Already Subscribed URL"`
	SubscribeForm string `json:"subscribe_form" dc:"Subscription Form HTML"`

	WelcomeSubject   string `json:"welcome_subject" dc:"Welcome Email Subject"`
	SendWelcomeEmail int    `json:"send_welcome_email" dc:"Whether to send a welcome email"`
	ConfirmSubject   string `json:"confirm_subject" dc:"Confirmation Email Subject"`
}

type UpdateGroupRes struct {
	api_v1.StandardRes
}

type UpdateGroupUnsubscribeReq struct {
	g.Meta        `path:"/contact/group/update_unsubscribe" method:"post" tags:"Contact" summary:"Update contact group Unsubscribe settings"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	GroupId       int    `json:"group_id" v:"required" dc:"Group ID"`

	UnsubscribeMailHtml    string `json:"unsubscribe_mail_html" dc:"Unsubscribe EmailHtml"`
	UnsubscribeMailDrag    string `json:"unsubscribe_mail_drag" dc:"Unsubscribe EmailDrag"`
	UnsubscribeSubject     string `json:"unsubscribe_subject" dc:"Unsubscribe Email Subject"`
	UnsubscribeRedirectUrl string `json:"unsubscribe_redirect_url" dc:"Unsubscribe Success RedirectURL"`
	SendUnsubscribeEmail   int    `json:"send_unsubscribe_email" dc:"Whether to send unsubscribe email"`
}

type UpdateGroupUnsubscribeRes struct {
	api_v1.StandardRes
}
type ListGroupsReq struct {
	g.Meta        `path:"/contact/group/list" method:"get" tags:"Contact" summary:"List contact groups"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Page          int    `json:"page" v:"required|min:1" dc:"Page Number"`
	PageSize      int    `json:"page_size" v:"required|min:1" dc:"Page Size"`
	Keyword       string `json:"keyword" dc:"Search Keyword"`
}

type ListGroupsRes struct {
	api_v1.StandardRes
	Data struct {
		Total int                 `json:"total" dc:"Total Count"`
		List  []*ContactGroupInfo `json:"list" dc:"Group List"`
	} `json:"data"`
}

type ListContactsReq struct {
	g.Meta        `path:"/contact/list" method:"get" tags:"Contact" summary:"List all contacts"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Page          int    `json:"page" v:"required|min:1" dc:"Page Number"`
	PageSize      int    `json:"page_size" v:"required|min:1" dc:"Page Size"`
	GroupId       int    `json:"group_id" dc:"Group ID(Optional)"`
	Keyword       string `json:"keyword" dc:"Search Email"`
	Status        int    `json:"status" v:"required|in:0,1" dc:"Status(1:Subscribed 0:Unsubscribed)" default:"1"`
}

type ListContactsRes struct {
	api_v1.StandardRes
	Data struct {
		Total int        `json:"total" dc:"Total Count"`
		List  []*Contact `json:"list" dc:"Contact List"`
	} `json:"data"`
}

type ListContactsGroupsReq struct {
	g.Meta        `path:"/contact/group/all" method:"get" tags:"Contact" summary:"List all contact groups"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Keyword       string `json:"keyword" dc:"Search Keyword"`
}

type ListContactsGroupsRes struct {
	api_v1.StandardRes
	Data struct {
		List []*ContactGroup `json:"list" dc:"Group List"`
	} `json:"data"`
}

type MergeContactsGroupsReq struct {
	g.Meta        `path:"/contact/group/merge_contacts" method:"post" tags:"Contact" summary:"Merge contact groups"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	SourceGroups  []int  `json:"source_groups" v:"required" dc:"Source Group IDs"`
	TargetGroup   int    `json:"target_group" v:"required" dc:"Target Group ID"`
}

type MergeContactsGroupsRes struct {
	api_v1.StandardRes
}

type DeleteContactsReq struct {
	g.Meta        `path:"/contact/delete" method:"post" tags:"Contact" summary:"Delete contacts"`
	Authorization string   `json:"authorization" dc:"Authorization" in:"header"`
	Emails        []string `json:"emails" v:"required" dc:"Email Addresses"`
	Status        int      `json:"status" v:"required|in:0,1" dc:"Status to Delete(1:Subscribed 0:Unsubscribed)"`
}

type DeleteContactsRes struct {
	api_v1.StandardRes
	Data struct {
		DeletedCount int `json:"deleted_count" dc:"Number of Deleted Contacts"`
	} `json:"data"`
}

type UpdateContactsGroupReq struct {
	g.Meta        `path:"/contact/update_group" method:"post" tags:"Contact" summary:"Update contact group"`
	Authorization string   `json:"authorization" dc:"Authorization" in:"header"`
	Emails        []string `json:"emails" v:"required" dc:"Email Addresses"`
	Status        int      `json:"status" v:"required|in:0,1" dc:"Contact Status(1:Subscribed 0:Unsubscribed)"`
	GroupIds      []int    `json:"group_ids" v:"required" dc:"New Group IDs"`
}

type UpdateContactsGroupRes struct {
	api_v1.StandardRes
	Data struct {
		UpdatedCount int `json:"updated_count" dc:"Number of Updated Contacts"`
	} `json:"data"`
}

type GetContactsTrendReq struct {
	g.Meta           `path:"/contact/trend" method:"get" tags:"Contact" summary:"Get contact trend"`
	Authorization    string `json:"authorization" dc:"Authorization" in:"header"`
	GroupId          int    `json:"group_id" dc:"Group ID"`
	Active           int    `json:"active" v:"in:0,1,-1" dc:"Active(1:Subscribed 0:Unsubscribed -1:all)" default:"-1"`
	LastActiveStatus int    `json:"last_active_status" v:"in:0,1,-1" dc:"Status(1:active 0:inactive -1:all)" default:"-1"`
	TimeInterval     int    `json:"time_interval" dc:"TimeInterval(7 : last 7 days, 30: last 30 days, 90: last 90 days, 180: last half year, 365: last year, 0: all)" default:"0"`
	Tags             string `json:"tags" dc:"Tags(-1:all or tag IDs, multiple IDs separated by commas)" default:"-1"`
}

type MonthlyTrend struct {
	Month            string `json:"month"               dc:"Month Format: YYYY-MM"`
	SubscribeCount   int    `json:"subscribe_count"    dc:"Subscribe Count"`
	UnsubscribeCount int    `json:"unsubscribe_count"  dc:"Unsubscribe Count"`
}

type DailyTrend struct {
	Date             string `json:"date"                dc:"Date Format: YYYY-MM-DD"`
	SubscribeCount   int    `json:"subscribe_count"    dc:"Subscribe Count"`
	UnsubscribeCount int    `json:"unsubscribe_count"  dc:"Unsubscribe Count"`
}

type GetContactsTrendRes struct {
	api_v1.StandardRes
	Data struct {
		TimeGranularity string          `json:"time_granularity" dc:"Time Granularity (daily/monthly)"`
		MonthlyData     []*MonthlyCount `json:"monthly_data" dc:"Monthly Trend Data"`
		DailyData       []*DailyCount   `json:"daily_data" dc:"Daily Trend Data"`
	} `json:"data"`
}

type MonthlyCount struct {
	Date  string `json:"date" dc:"Month Format: YYYY-MM"`
	Count int    `json:"count" dc:"Count"`
}

type DailyCount struct {
	Date  string `json:"date" dc:"Date Format: YYYY-MM-DD"`
	Count int    `json:"count" dc:"Count"`
}

type GetGroupContactCountReq struct {
	g.Meta        `path:"/contact/group/contact_count" method:"post" tags:"Contact" summary:"Get group contact count"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	GroupIds      []int  `json:"group_ids" v:"required" dc:"Group IDs"`
}

type GetGroupContactCountRes struct {
	api_v1.StandardRes
	Data struct {
		Total int `json:"total" dc:"Total contact count"`
	} `json:"data" dc:"Data"`
}

type GetSingleGroupTagContactCountReq struct {
	g.Meta        `path:"/contact/group/tag_contact_count" method:"post" tags:"Contact" summary:"Get single group contact count with tag ids"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	GroupId       int    `json:"group_id" v:"required" dc:"Group ID"`
	TagIds        []int  `json:"tag_ids"  dc:"Tag IDs"`
	TagLogic      string `json:"tag_logic" v:"required|in:AND,OR,NOT" dc:"Tag Logic (AND/OR/NOT)"`
}

type GetSingleGroupTagContactCountRes struct {
	api_v1.StandardRes
	Data struct {
		Total int `json:"total" dc:"Total contact count with given tag ids in the group"`
	} `json:"data" dc:"Data"`
}

type GetGroupInfoReq struct {
	g.Meta        `path:"/contact/group/info" method:"get" tags:"Contact" summary:"Get group info"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	GroupId       int    `json:"group_id" v:"required" dc:"Group ID"`
}

type GetGroupInfoRes struct {
	api_v1.StandardRes
	Data *ContactGroupInfo `json:"data" dc:"Group Info"`
}

type EditContactsReq struct {
	g.Meta        `path:"/contact/edit" method:"post" tags:"Contact" summary:"Edit contact"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Emails        string `json:"emails"   v:"required"`
	Active        int    `json:"active"`
	GroupIds      []int  `json:"group_ids"`
	Attribs       string `json:"attribs"`
}

type EditContactsRes struct {
	api_v1.StandardRes
}

type ListContactsNDPReq struct {
	g.Meta           `path:"/contact/list_ndp" method:"get" tags:"Contact" summary:"List all contacts"`
	Authorization    string `json:"authorization" dc:"Authorization" in:"header"`
	Page             int    `json:"page" v:"required|min:1" dc:"Page Number"`
	PageSize         int    `json:"page_size" v:"required|min:1" dc:"Page Size"`
	GroupId          int    `json:"group_id" dc:"Group ID(Optional)"`
	Keyword          string `json:"keyword" dc:"Search Email"`
	Active           int    `json:"active" v:"required|in:0,1,-1" dc:"Active(1:Subscribed 0:Unsubscribed -1:all)" default:"-1"`
	LastActiveStatus int    `json:"last_active_status" v:"in:0,1,-1" dc:"Status(1:active 0:inactive -1:all)" default:"-1"`
	TimeInterval     int    `json:"time_interval" dc:"TimeInterval(7 : last 7 days, 30: last 30 days, 90: last 90 days, 180: last half year, 365: last year, 0: all)" default:"0"`
	Tags             string `json:"tags" dc:"Tags(-1:all or tag IDs, multiple IDs separated by commas)" default:"-1"`
	SortBy           string `json:"sort_by" v:"in:create_time,last_active_at" dc:"Sort field (create_time: creation time, last_active_at: last active time)" default:"create_time"`
	SortOrder        string `json:"sort_order" v:"in:asc,desc" dc:"Sort order (asc: ascending, desc: descending)" default:"desc"`
}

type ListContactsNDPRes struct {
	api_v1.StandardRes
	Data struct {
		Total int        `json:"total" dc:"Total Count"`
		List  []*Contact `json:"list" dc:"Contact List"`
	} `json:"data"`
}

type EditContactsNDPReq struct {
	g.Meta        `path:"/contact/edit_ndp" method:"post" tags:"Contact" summary:"Edit contact"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Id            int    `json:"id"   v:"required"`
	Active        int    `json:"active"`
	Status        int    `json:"status"`
	Attribs       string `json:"attribs"`
}

type EditContactsNDPRes struct {
	api_v1.StandardRes
}

type DeleteContactsNDPReq struct {
	g.Meta        `path:"/contact/delete_ndp" method:"post" tags:"Contact" summary:"Delete contacts"`
	Authorization string   `json:"authorization" dc:"Authorization" in:"header"`
	Ids           []string `json:"ids" v:"required" dc:"contacts id"`
}

type DeleteContactsNDPRes struct {
	api_v1.StandardRes
}

type BatchTagContactsReq struct {
	g.Meta        `path:"/contact/batch_tags_opt" method:"post" tags:"Contact" summary:"Tag contacts in bulk"`
	Authorization string `json:"authorization" dc:"Authorization" in:"header"`
	Ids           []int  `json:"ids" v:"required" dc:"Contact IDs"`
	TagIds        []int  `json:"tag_ids" v:"required" dc:"Tag IDs"`
	Action        int    `json:"action" v:"required|in:1,2" dc:"Action (1: Add Tag, 2: Remove Tag)"`
}
type BatchTagContactsRes struct {
	api_v1.StandardRes
}
