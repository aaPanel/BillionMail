package settings

import (
	"billionmail-core/api/settings/v1"
	"billionmail-core/internal/service/public"
	"context"
	"encoding/json"
	"github.com/gogf/gf/v2/os/gfile"
	"os"
	"path/filepath"
	"sort"
)

func (c *ControllerV1) GetTimeZoneList(ctx context.Context, req *v1.GetTimeZoneListReq) (res *v1.GetTimeZoneListRes, err error) {
	res = &v1.GetTimeZoneListRes{}

	zoneFile := public.AbsPath("../core/data/zoneinfo_list.json")
	group, err := loadTimeZoneGroupFromFile(zoneFile)
	if err != nil || len(group.ZoneList) == 0 {
		group, _ = getTimeZoneGroup()
		err = saveTimeZoneGroupToFile(group, zoneFile)
	}

	res.Data = group.Zones
	res.SetSuccess(public.LangCtx(ctx, "Successful"))
	return res, nil
}

type TimeZoneGroup struct {
	ZoneList []string            `json:"zoneList"`
	Zones    map[string][]string `json:"zones"`
}

func getTimeZoneGroup() (TimeZoneGroup, error) {
	zoneList := []string{"Asia", "Africa", "America", "Antarctica", "Arctic", "Atlantic", "Australia", "Europe", "Indian", "Pacific"}
	zones := make(map[string][]string)
	root := "/usr/share/zoneinfo"

	for _, mainZone := range zoneList {
		dir := filepath.Join(root, mainZone)
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, f := range files {
			if !f.IsDir() {
				zones[mainZone] = append(zones[mainZone], f.Name())
			}
		}
		sort.Strings(zones[mainZone])
	}
	return TimeZoneGroup{
		ZoneList: zoneList,
		Zones:    zones,
	}, nil
}

func saveTimeZoneGroupToFile(group TimeZoneGroup, filePath string) error {
	data, err := json.Marshal(group)
	if err != nil {
		return err
	}
	return gfile.PutContents(filePath, string(data))
}

func loadTimeZoneGroupFromFile(filePath string) (TimeZoneGroup, error) {
	var group TimeZoneGroup
	if !gfile.Exists(filePath) {
		return group, os.ErrNotExist
	}
	data := gfile.GetContents(filePath)
	if err := json.Unmarshal([]byte(data), &group); err != nil {
		return group, err
	}
	return group, nil
}
