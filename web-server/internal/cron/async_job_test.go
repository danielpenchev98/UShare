package cron_test

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/danielpenchev98/UShare/web-server/internal/cron"
	"github.com/danielpenchev98/UShare/web-server/internal/db/dao/dao_mocks"
	myerr "github.com/danielpenchev98/UShare/web-server/internal/error"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GroupEraserJobImpl", func() {
	var (
		groupEraser cron.GroupEraserJob
		uamDAO      *dao_mocks.MockUamDAO
		testDir     string
	)

	BeforeEach(func() {
		testDir, _ = os.Getwd()
		controller := gomock.NewController(GinkgoT())
		uamDAO = dao_mocks.NewMockUamDAO(controller)
		groupEraser = cron.NewGroupEraserJobImpl(uamDAO, testDir)
	})

	When("deleting the deactivated groups", func() {
		var (
			groupDirName string
			groupDirPath string
		)
		const testFileName = "test-file"

		BeforeEach(func() {
			groupDirName = "test"
			groupDirPath = path.Join(testDir, "test")
			os.Mkdir(groupDirPath, 0755)
			createFile(path.Join(groupDirPath, testFileName))
		})

		AfterEach(func() {
			os.RemoveAll(groupDirPath)
		})

		Context("and request to fetch deactivated group names fails", func() {
			BeforeEach(func() {
				uamDAO.EXPECT().
					GetDeactivatedGroupNames().
					Return(nil, myerr.NewServerError("test-error"))

				uamDAO.EXPECT().
					EraseDeactivatedGroups(gomock.Any()).
					Times(0)
			})

			It("shoudnt delete resources", func() {
				groupEraser.DeleteGroups()

				_, err := os.Stat(groupDirPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(getCountFiles(groupDirPath)).To(Equal(1))
			})
		})

		Context("and request to fetch deactivated group names succeeds", func() {
			var groupsToDelete []string

			BeforeEach(func() {
				groupsToDelete = []string{groupDirName}
			})

			Context("and request to erase group records in db fails", func() {
				BeforeEach(func() {
					uamDAO.EXPECT().
						GetDeactivatedGroupNames().
						Return(groupsToDelete, nil)

					uamDAO.EXPECT().
						EraseDeactivatedGroups(groupsToDelete).
						Return(myerr.NewServerError("test-error"))
				})

				It("should delete files from FS but not group records in db", func() {
					groupEraser.DeleteGroups()

					_, err := os.Stat(groupDirPath)
					Expect(err).To(HaveOccurred())
					Expect(os.IsNotExist(err)).To(BeTrue())
				})
			})

			Context("and request to erase group records in db succeeds", func() {
				BeforeEach(func() {
					uamDAO.EXPECT().
						GetDeactivatedGroupNames().
						Return(groupsToDelete, nil)

					uamDAO.EXPECT().
						EraseDeactivatedGroups(groupsToDelete).
						Return(nil)
				})

				It("should delete files from FS and group records in db", func() {
					groupEraser.DeleteGroups()

					_, err := os.Stat(groupDirPath)
					Expect(err).To(HaveOccurred())
					Expect(os.IsNotExist(err)).To(BeTrue())
				})
			})
		})
	})
})

func createFile(name string) error {
	file, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	return file.Close()
}

func getCountFiles(dirName string) int {
	files, _ := ioutil.ReadDir(dirName)
	return len(files)
}
