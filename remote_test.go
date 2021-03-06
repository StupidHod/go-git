package git

import (
	"io/ioutil"
	"os"

	"github.com/StupidHod/go-git/config"
	"github.com/StupidHod/go-git/core"
	"github.com/StupidHod/go-git/storage/filesystem"
	"github.com/StupidHod/go-git/storage/memory"
	osfs "github.com/StupidHod/go-git/utils/fs/os"

	. "gopkg.in/check.v1"
)

const FixRefSpec = config.RefSpec("+refs/heads/*:refs/remotes/origin/*")

type RemoteSuite struct {
	BaseSuite
}

var _ = Suite(&RemoteSuite{})

func (s *RemoteSuite) TestConnect(c *C) {
	r := newRemote(nil, &config.RemoteConfig{Name: "foo", URL: RepositoryFixture})

	err := r.Connect()
	c.Assert(err, IsNil)
}

func (s *RemoteSuite) TestnewRemoteInvalidEndpoint(c *C) {
	r := newRemote(nil, &config.RemoteConfig{Name: "foo", URL: "qux"})

	err := r.Connect()
	c.Assert(err, NotNil)
}

func (s *RemoteSuite) TestnewRemoteInvalidSchemaEndpoint(c *C) {
	r := newRemote(nil, &config.RemoteConfig{Name: "foo", URL: "qux://foo"})

	err := r.Connect()
	c.Assert(err, NotNil)
}

func (s *RemoteSuite) TestInfo(c *C) {
	r := newRemote(nil, &config.RemoteConfig{Name: "foo", URL: RepositoryFixture})
	r.upSrv = &MockGitUploadPackService{}

	c.Assert(r.Info(), IsNil)
	c.Assert(r.Connect(), IsNil)
	c.Assert(r.Info(), NotNil)
	c.Assert(r.Info().Capabilities.Get("ofs-delta"), NotNil)
}

func (s *RemoteSuite) TestDefaultBranch(c *C) {
	r := newRemote(nil, &config.RemoteConfig{Name: "foo", URL: RepositoryFixture})
	r.upSrv = &MockGitUploadPackService{}

	c.Assert(r.Connect(), IsNil)
	c.Assert(r.Head().Name(), Equals, core.ReferenceName("refs/heads/master"))
}

func (s *RemoteSuite) TestCapabilities(c *C) {
	r := newRemote(nil, &config.RemoteConfig{Name: "foo", URL: RepositoryFixture})
	r.upSrv = &MockGitUploadPackService{}

	c.Assert(r.Connect(), IsNil)
	c.Assert(r.Capabilities().Get("agent").Values, HasLen, 1)
}

func (s *RemoteSuite) TestFetch(c *C) {
	sto := memory.NewStorage()
	r := newRemote(sto, &config.RemoteConfig{Name: "foo", URL: RepositoryFixture})
	r.upSrv = &MockGitUploadPackService{}

	c.Assert(r.Connect(), IsNil)

	err := r.Fetch(&FetchOptions{
		RefSpecs: []config.RefSpec{FixRefSpec},
	})

	c.Assert(err, IsNil)
	c.Assert(sto.ObjectStorage().(*memory.ObjectStorage).Objects, HasLen, 31)

	expectedRefs := []*core.Reference{
		core.NewReferenceFromStrings("refs/remotes/origin/master", "6ecf0ef2c2dffb796033e5a02219af86ec6584e5"),
		core.NewReferenceFromStrings("refs/remotes/origin/branch", "e8d3ffab552895c19b9fcf7aa264d277cde33881"),
		core.NewReferenceFromStrings("refs/tags/v1.0.0", "6ecf0ef2c2dffb796033e5a02219af86ec6584e5"),
	}

	for _, exp := range expectedRefs {
		r, _ := sto.ReferenceStorage().Get(exp.Name())
		c.Assert(exp.String(), Equals, r.String())
	}
}

func (s *RemoteSuite) TestFetchObjectStorageWriter(c *C) {
	dir, err := ioutil.TempDir("", "fetch")
	c.Assert(err, IsNil)

	defer os.RemoveAll(dir) // clean up

	var sto Storage
	sto, err = filesystem.NewStorage(osfs.NewOS(dir))
	c.Assert(err, IsNil)

	r := newRemote(sto, &config.RemoteConfig{Name: "foo", URL: RepositoryFixture})
	r.upSrv = &MockGitUploadPackService{}

	c.Assert(r.Connect(), IsNil)

	err = r.Fetch(&FetchOptions{
		RefSpecs: []config.RefSpec{FixRefSpec},
	})

	c.Assert(err, IsNil)

	var count int
	iter, err := sto.ObjectStorage().Iter(core.AnyObject)
	c.Assert(err, IsNil)

	iter.ForEach(func(core.Object) error {
		count++
		return nil
	})
	c.Assert(count, Equals, 31)
}

func (s *RemoteSuite) TestFetchNoErrAlreadyUpToDate(c *C) {
	sto := memory.NewStorage()
	r := newRemote(sto, &config.RemoteConfig{Name: "foo", URL: RepositoryFixture})
	r.upSrv = &MockGitUploadPackService{}

	c.Assert(r.Connect(), IsNil)

	o := &FetchOptions{
		RefSpecs: []config.RefSpec{FixRefSpec},
	}

	err := r.Fetch(o)
	c.Assert(err, IsNil)
	err = r.Fetch(o)
	c.Assert(err, Equals, NoErrAlreadyUpToDate)
}

func (s *RemoteSuite) TestHead(c *C) {
	r := newRemote(nil, &config.RemoteConfig{Name: "foo", URL: RepositoryFixture})
	r.upSrv = &MockGitUploadPackService{}

	err := r.Connect()
	c.Assert(err, IsNil)
	c.Assert(r.Head().Hash(), Equals, core.NewHash("6ecf0ef2c2dffb796033e5a02219af86ec6584e5"))
}

func (s *RemoteSuite) TestRef(c *C) {
	r := newRemote(nil, &config.RemoteConfig{Name: "foo", URL: RepositoryFixture})
	r.upSrv = &MockGitUploadPackService{}

	err := r.Connect()
	c.Assert(err, IsNil)

	ref, err := r.Ref(core.HEAD, false)
	c.Assert(err, IsNil)
	c.Assert(ref.Name(), Equals, core.HEAD)

	ref, err = r.Ref(core.HEAD, true)
	c.Assert(err, IsNil)
	c.Assert(ref.Name(), Equals, core.ReferenceName("refs/heads/master"))
}

func (s *RemoteSuite) TestRefs(c *C) {
	r := newRemote(nil, &config.RemoteConfig{Name: "foo", URL: RepositoryFixture})
	r.upSrv = &MockGitUploadPackService{}

	err := r.Connect()
	c.Assert(err, IsNil)

	iter, err := r.Refs()
	c.Assert(err, IsNil)
	c.Assert(iter, NotNil)
}

func (s *RemoteSuite) TestString(c *C) {
	r := newRemote(nil, &config.RemoteConfig{Name: "foo", URL: RepositoryFixture})
	c.Assert(r.String(), Equals, ""+
		"foo\thttps://github.com/git-fixtures/basic.git (fetch)\n"+
		"foo\thttps://github.com/git-fixtures/basic.git (push)",
	)
}
