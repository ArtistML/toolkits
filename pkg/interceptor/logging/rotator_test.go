package logging

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/gogf/gf/frame/g"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFileRotator(t *testing.T) {
	Convey("Testing default", t, func(c C) {
		r := new(FileRotator)
		So(r.fileName(), ShouldNotBeEmpty)
		So(r.maxDuration(), ShouldEqual, defaultMaxDuration)
		So(r.maxSize(), ShouldEqual, defaultMaxSize)
		So(r.timeFormat(), ShouldEqual, defaultTimeFormat)
	})
	Convey("Testing FileRotator", t, func(c C) {
		r := &FileRotator{
			FileName:    filepath.Join(g.Cfg().GetString("logger.Path"), "rotator.log"),
			TimeFormat:  time.RFC3339Nano,
			MaxDuration: 10 * time.Millisecond,
			MaxSize:     32,
		}
		defer func() { _ = r.Close() }()
		short := []byte("[DEBUG] short\n")
		long := []byte("[ERROR] loooooooooooooooooooooooooooooooong\n")
		n, err := r.Write(short)
		So(err, ShouldBeNil)
		So(n, ShouldEqual, len(short))
		_, err = r.Write(long)
		So(err, ShouldNotBeNil)
		Convey("Chaos test", func() {
			r.MaxSize = 1 << 10
			for i := 0; i < 1000; i++ {
				_, err = r.Write([]byte(fmt.Sprintf("[DEBUG] count %d\n", i)))
				So(err, ShouldBeNil)
			}
		})
	})
}
