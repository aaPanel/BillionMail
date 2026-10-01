package testutil

import (
	"context"
	"time"

	"billionmail-core/internal/model/entity"
)

// --- Test Context ---

// TestCtx returns a context with a 5-second timeout for tests.
func TestCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// --- Fixture Builders (functional options pattern) ---

// ContactOption configures a test Contact.
type ContactOption func(*entity.Contact)

// NewContact creates a Contact with sensible defaults. Override with options.
func NewContact(opts ...ContactOption) *entity.Contact {
	now := int(time.Now().Unix())
	c := &entity.Contact{
		Id:           1,
		Email:        "test@example.com",
		GroupId:      1,
		Active:       1,
		TaskId:       0,
		CreateTime:   now,
		Status:       1,
		Attribs:      map[string]string{},
		LastActiveAt: now,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

func WithContactID(id int) ContactOption         { return func(c *entity.Contact) { c.Id = id } }
func WithContactEmail(e string) ContactOption     { return func(c *entity.Contact) { c.Email = e } }
func WithContactGroupID(gid int) ContactOption    { return func(c *entity.Contact) { c.GroupId = gid } }
func WithContactActive(a int) ContactOption       { return func(c *entity.Contact) { c.Active = a } }
func WithContactStatus(s int) ContactOption       { return func(c *entity.Contact) { c.Status = s } }
func WithContactAttribs(a map[string]string) ContactOption {
	return func(c *entity.Contact) { c.Attribs = a }
}

// --- EmailTask ---

// EmailTaskOption configures a test EmailTask.
type EmailTaskOption func(*entity.EmailTask)

// NewEmailTask creates an EmailTask with sensible defaults.
func NewEmailTask(opts ...EmailTaskOption) *entity.EmailTask {
	now := int(time.Now().Unix())
	t := &entity.EmailTask{
		Id:             1,
		TaskName:       "Test Campaign",
		Addresser:      "sender@example.com",
		Subject:        "Test Subject",
		FullName:       "Test Sender",
		RecipientCount: 100,
		TaskProcess:    0,
		Pause:          0,
		TemplateId:     1,
		IsRecord:       1,
		Unsubscribe:    1,
		Threads:        5,
		TrackOpen:      1,
		TrackClick:     1,
		StartTime:      now,
		CreateTime:     now,
		UpdateTime:     now,
		Active:         1,
		GroupId:        1,
		TagIds:         []int{},
		TagLogic:       "AND",
		UseTagFilter:   0,
	}
	for _, o := range opts {
		o(t)
	}
	return t
}

func WithTaskID(id int) EmailTaskOption           { return func(t *entity.EmailTask) { t.Id = id } }
func WithTaskName(n string) EmailTaskOption        { return func(t *entity.EmailTask) { t.TaskName = n } }
func WithTaskSubject(s string) EmailTaskOption     { return func(t *entity.EmailTask) { t.Subject = s } }
func WithTaskAddresser(a string) EmailTaskOption   { return func(t *entity.EmailTask) { t.Addresser = a } }
func WithTaskTemplateID(id int) EmailTaskOption    { return func(t *entity.EmailTask) { t.TemplateId = id } }
func WithTaskGroupID(gid int) EmailTaskOption      { return func(t *entity.EmailTask) { t.GroupId = gid } }
func WithTaskPause(p int) EmailTaskOption          { return func(t *entity.EmailTask) { t.Pause = p } }
func WithTaskActive(a int) EmailTaskOption         { return func(t *entity.EmailTask) { t.Active = a } }
func WithTaskThreads(n int) EmailTaskOption        { return func(t *entity.EmailTask) { t.Threads = n } }
func WithTaskTagIds(ids []int) EmailTaskOption     { return func(t *entity.EmailTask) { t.TagIds = ids } }

// --- EmailTemplate ---

// EmailTemplateOption configures a test EmailTemplate.
type EmailTemplateOption func(*entity.EmailTemplate)

// NewEmailTemplate creates an EmailTemplate with sensible defaults.
func NewEmailTemplate(opts ...EmailTemplateOption) *entity.EmailTemplate {
	now := int(time.Now().Unix())
	t := &entity.EmailTemplate{
		Id:         1,
		TempName:   "Test Template",
		AddType:    1,
		Content:    "<html><body>Hello {{.Name}}</body></html>",
		Render:     "",
		CreateTime: now,
		UpdateTime: now,
	}
	for _, o := range opts {
		o(t)
	}
	return t
}

func WithTemplateID(id int) EmailTemplateOption        { return func(t *entity.EmailTemplate) { t.Id = id } }
func WithTemplateName(n string) EmailTemplateOption     { return func(t *entity.EmailTemplate) { t.TempName = n } }
func WithTemplateContent(c string) EmailTemplateOption  { return func(t *entity.EmailTemplate) { t.Content = c } }
func WithTemplateAddType(a int) EmailTemplateOption     { return func(t *entity.EmailTemplate) { t.AddType = a } }

// --- ContactGroup ---

// ContactGroupOption configures a test ContactGroup.
type ContactGroupOption func(*entity.ContactGroup)

// NewContactGroup creates a ContactGroup with sensible defaults.
func NewContactGroup(opts ...ContactGroupOption) *entity.ContactGroup {
	now := int(time.Now().Unix())
	g := &entity.ContactGroup{
		Id:                   1,
		Name:                 "Test Group",
		Description:          "A test contact group",
		CreateTime:           now,
		UpdateTime:           now,
		Token:                "test-token-abc123",
		DoubleOptin:          0,
		SendWelcomeEmail:     0,
		SendUnsubscribeEmail: 0,
	}
	for _, o := range opts {
		o(g)
	}
	return g
}

func WithGroupID(id int) ContactGroupOption          { return func(g *entity.ContactGroup) { g.Id = id } }
func WithGroupName(n string) ContactGroupOption      { return func(g *entity.ContactGroup) { g.Name = n } }
func WithGroupToken(t string) ContactGroupOption     { return func(g *entity.ContactGroup) { g.Token = t } }
func WithGroupDoubleOptin(d int) ContactGroupOption  { return func(g *entity.ContactGroup) { g.DoubleOptin = d } }
func WithGroupDescription(d string) ContactGroupOption {
	return func(g *entity.ContactGroup) { g.Description = d }
}

// --- Mock Interface Pattern ---

// ContactRepository defines the interface for contact data access.
// Implement this in tests with mock implementations.
type ContactRepository interface {
	GetByID(ctx context.Context, id int) (*entity.Contact, error)
	GetByEmail(ctx context.Context, email string) (*entity.Contact, error)
	List(ctx context.Context, groupID int, page, pageSize int) ([]*entity.Contact, int, error)
	Create(ctx context.Context, contact *entity.Contact) (int, error)
	Update(ctx context.Context, contact *entity.Contact) error
	Delete(ctx context.Context, id int) error
}

// EmailTaskRepository defines the interface for email task data access.
type EmailTaskRepository interface {
	GetByID(ctx context.Context, id int) (*entity.EmailTask, error)
	List(ctx context.Context, page, pageSize int) ([]*entity.EmailTask, int, error)
	Create(ctx context.Context, task *entity.EmailTask) (int, error)
	Update(ctx context.Context, task *entity.EmailTask) error
	Delete(ctx context.Context, id int) error
}

// EmailTemplateRepository defines the interface for template data access.
type EmailTemplateRepository interface {
	GetByID(ctx context.Context, id int) (*entity.EmailTemplate, error)
	List(ctx context.Context, page, pageSize int) ([]*entity.EmailTemplate, int, error)
	Create(ctx context.Context, tpl *entity.EmailTemplate) (int, error)
	Update(ctx context.Context, tpl *entity.EmailTemplate) error
	Delete(ctx context.Context, id int) error
}

// ContactGroupRepository defines the interface for group data access.
type ContactGroupRepository interface {
	GetByID(ctx context.Context, id int) (*entity.ContactGroup, error)
	List(ctx context.Context, page, pageSize int) ([]*entity.ContactGroup, int, error)
	Create(ctx context.Context, group *entity.ContactGroup) (int, error)
	Update(ctx context.Context, group *entity.ContactGroup) error
	Delete(ctx context.Context, id int) error
}
