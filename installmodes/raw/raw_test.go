package raw

import (
	"fmt"
	"testing"

	"bitbucket.org/ossystems/agent/installmodes"
	"bitbucket.org/ossystems/agent/libarchive"
	"bitbucket.org/ossystems/agent/testsmocks"
	"bitbucket.org/ossystems/agent/utils"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestRawInit(t *testing.T) {
	val, err := installmodes.GetObject("raw")
	assert.NoError(t, err)

	r1, ok := val.(*RawObject)
	if !ok {
		t.Error("Failed to cast return value of \"installmodes.GetObject()\" to RawObject")
	}

	osFs := afero.NewOsFs()
	r2 := &RawObject{
		LibArchiveBackend: &libarchive.LibArchive{},
		FileSystemBackend: osFs,
		Copier:            &utils.ExtendedIO{},
		ChunkSize:         128 * 1024,
		Skip:              0,
		Seek:              0,
		Count:             -1,
		Truncate:          true,
	}

	assert.Equal(t, r2, r1)
}

func TestRawSetupWithSuccess(t *testing.T) {
	r := RawObject{}
	r.TargetType = "device"
	err := r.Setup()
	assert.NoError(t, err)
}

func TestRawSetupWithNotSupportedTargetTypes(t *testing.T) {
	r := RawObject{}

	r.TargetType = "ubivolume"
	err := r.Setup()
	assert.EqualError(t, err, "target-type 'ubivolume' is not supported for the 'raw' handler. Its value must be 'device'")

	r.TargetType = "mtdname"
	err = r.Setup()
	assert.EqualError(t, err, "target-type 'mtdname' is not supported for the 'raw' handler. Its value must be 'device'")

	r.TargetType = "someother"
	err = r.Setup()
	assert.EqualError(t, err, "target-type 'someother' is not supported for the 'raw' handler. Its value must be 'device'")
}

func TestRawInstallWithCopyFileError(t *testing.T) {
	fsbm := &testsmocks.FileSystemBackendMock{}

	lam := &testsmocks.LibArchiveMock{}

	targetDevice := "/dev/xx1"
	sha256sum := "5bdbf286cb4adcff26befa2183f3167c053bc565036736eaa2ae429fe910d93c"
	compressed := false

	cm := &testsmocks.CopierMock{}
	cm.On("CopyFile", fsbm, lam, sha256sum, targetDevice, 128*1024, 0, 0, -1, true, compressed).Return(fmt.Errorf("copy file error"))

	r := RawObject{
		Copier:            cm,
		FileSystemBackend: fsbm,
		LibArchiveBackend: lam,
		ChunkSize:         128 * 1024,
		Count:             -1,
		Truncate:          true,
	}
	r.Target = targetDevice
	r.Sha256sum = sha256sum
	r.Compressed = compressed

	err := r.Install()

	assert.EqualError(t, err, "copy file error")
	cm.AssertExpectations(t)
	lam.AssertExpectations(t)
	fsbm.AssertExpectations(t)
}

func TestRawInstallWithSuccess(t *testing.T) {
	testCases := []struct {
		Name              string
		Sha256sum         string
		Target            string
		TargetType        string
		ChunkSize         int
		Skip              int
		Seek              int
		Count             int
		Truncate          bool
		ExpectedChunkSize int
		Compressed        bool
	}{
		{
			"WithAllFields",
			"5bdbf286cb4adcff26befa2183f3167c053bc565036736eaa2ae429fe910d93c",
			"/dev/xx1",
			"device",
			2048,
			2,
			3,
			-1,
			true,
			2048,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			fsbm := &testsmocks.FileSystemBackendMock{}

			lam := &testsmocks.LibArchiveMock{}

			cm := &testsmocks.CopierMock{}
			cm.On("CopyFile", fsbm, lam, tc.Sha256sum, tc.Target, tc.ExpectedChunkSize, tc.Skip, tc.Seek, tc.Count, tc.Truncate, tc.Compressed).Return(nil)

			r := RawObject{Copier: cm, FileSystemBackend: fsbm, LibArchiveBackend: lam}
			r.Target = tc.Target
			r.TargetType = tc.TargetType
			r.Sha256sum = tc.Sha256sum
			r.ChunkSize = tc.ChunkSize
			r.Skip = tc.Skip
			r.Seek = tc.Seek
			r.Count = tc.Count
			r.Truncate = tc.Truncate
			r.Compressed = tc.Compressed

			err := r.Install()

			assert.NoError(t, err)
			cm.AssertExpectations(t)
			lam.AssertExpectations(t)
			fsbm.AssertExpectations(t)
		})
	}
}

func TestRawCleanupNil(t *testing.T) {
	r := RawObject{}
	assert.Nil(t, r.Cleanup())
}