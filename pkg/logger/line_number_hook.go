package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

type LineNumberHook struct {
}

func (hook LineNumberHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook LineNumberHook) Fire(entry *logrus.Entry) error {
	if entry.Data["f"] == nil {
		return nil
	}
	s := strings.Split(entry.Data["f"].(string), ":")
	if entry.Data["root"] != nil { // called from one of root level functions such as Errorf, Infof etc.
		entry.Data["f"] = fmt.Sprintf("%s:%s", s[0], s[1])
	} else { // called after logger was instantiated using WithFields function
		entry.Data["f"] = fmt.Sprintf("%s:%d", s[0], entry.Caller.Line)
	}
	delete(entry.Data, "root") // delete the transitive property, if not deleted [root:true] will be printed in log
	return nil
}
