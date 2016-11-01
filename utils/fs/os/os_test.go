package os_test

import (
	"io/ioutil"
	stdos "os"
	"testing"

	. "gopkg.in/check.v1"
	"github.com/StupidHod/go-git/utils/fs/os"
	"github.com/StupidHod/go-git/utils/fs/test"
)

func Test(t *testing.T) { TestingT(t) }

type OSSuite struct {
	test.FilesystemSuite
	path string
}

var _ = Suite(&OSSuite{})

func (s *OSSuite) SetUpTest(c *C) {
	s.path, _ = ioutil.TempDir(stdos.TempDir(), "go-git-os-fs-test")
	s.FilesystemSuite.Fs = os.NewOS(s.path)
}
func (s *OSSuite) TearDownTest(c *C) {
	err := stdos.RemoveAll(s.path)
	c.Assert(err, IsNil)
}
