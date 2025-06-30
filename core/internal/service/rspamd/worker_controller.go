package rspamd

import (
	"billionmail-core/internal/consts"
	docker "billionmail-core/internal/service/dockerapi"
	"billionmail-core/internal/service/public"
	"context"
	"errors"
	"github.com/gogf/gf/util/grand"
	"github.com/gogf/gf/v2/frame/g"
	"path/filepath"
	"strings"
)

// InitWorkerController initializes the worker controller
func InitWorkerController() (err error) {
	wci := filepath.Join(public.AbsPath(consts.RSPAMD_LOCAL_D_PATH), "worker-controller.inc")

	// Check if the file exists
	if public.FileExists(wci) {
		// Check if the password already exists
		ex, _ := g.DB().Model("bm_options").Where("name", "rspamd_worker_controller_password").Exist()
		if ex {
			return
		}
	}

	// Generate password
	passwordPlain := grand.S(16)

	// Generate the worker controller configuration
	dk, err := docker.NewDockerAPI()

	if err != nil {
		return
	}

	defer dk.Close()

	res, err := dk.ExecCommandByName(context.Background(), "billionmail-rspamd-billionmail-1", []string{"rspamadm", "pw", "-p", passwordPlain}, "root")

	if err != nil {
		return
	}

	if res == nil {
		err = errors.New("rspamdadm not found")
		return
	}

	_, err = public.WriteFile(wci, `count = 1;
password = "`+strings.TrimSpace(strings.SplitN(strings.TrimSpace(res.Output), "\n", 2)[0])+`";
secure_ip = "127.0.0.1";
secure_ip = "::1";
static_dir = "${WWWDIR}";`)

	if err != nil {
		return
	}

	// Save the password
	err = public.OptionsMgrInstance.SetOption(context.Background(), "rspamd_worker_controller_password", passwordPlain)

	if err != nil {
		return
	}

	// Restart the RSPAMD container
	err = dk.RestartContainer(context.Background(), "billionmail-rspamd-billionmail-1")

	return
}
