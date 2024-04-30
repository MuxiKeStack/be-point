package events

import (
	"encoding/json"
	"github.com/MuxiKeStack/be-point/pkg/canalx"
)

type MysqlBinlogEventHandler interface {
	Handle(event canalx.Message[any]) error
}

func copyByJSON(src any, dst any) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}
