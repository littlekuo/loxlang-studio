package util

import (
	"fmt"
)

func ErrorMsg(line int, message string) error {
	return errorMsgDetail(line, "", message)
}

func errorMsgDetail(line int, where string, message string) error {
	return fmt.Errorf("[line %d] Error%s: %s", line, where, message)
}
