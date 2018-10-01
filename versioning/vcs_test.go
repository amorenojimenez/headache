/*
 * Copyright 2018 Florent Biville (@fbiville)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package versioning

import (
	"github.com/fbiville/header/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

var (
	vcs     Vcs
	vcsMock *mocks.Vcs
)

func TestCommittedChanges(t *testing.T) {
	I := NewGomegaWithT(t)
	controller := gomock.NewController(t)
	defer controller.Finish()
	vcs = new(mocks.Vcs)
	vcsMock = vcs.(*mocks.Vcs)
	defer vcsMock.AssertExpectations(t)
	vcsMock.On("Diff", []string{"--name-status", "origin/master..HEAD"}).Return(`M	.gitignore
M	configuration.go
D	header.go
D	header_test.go
R099	line_comment.go	core/line_comment.go
A	license-header.txt
`, nil)

	changes, err := getCommittedChanges(vcs, "origin", "master")

	I.Expect(err).To(BeNil())
	I.Expect(changes).To(Equal([]FileChange{
		{Path: ".gitignore"},
		{Path: "configuration.go"},
		{Path: "core/line_comment.go"},
		{Path: "license-header.txt"},
	}))

}

func TestUncommittedFiles(t *testing.T) {
	I := NewGomegaWithT(t)
	controller := gomock.NewController(t)
	defer controller.Finish()
	vcs = new(mocks.Vcs)
	vcsMock = vcs.(*mocks.Vcs)
	defer vcsMock.AssertExpectations(t)
	vcsMock.On("Status", []string{"--porcelain"}).Return(` M Gopkg.lock
 D main.go
?? build.sh
?? git.go
`, nil)

	changes, err := getUncommittedChanges(vcs)

	I.Expect(err).To(BeNil())
	I.Expect(changes).To(Equal([]FileChange{
		{Path: "Gopkg.lock"},
		{Path: "build.sh"},
		{Path: "git.go"},
	}))
}

func TestNoUncommittedFiles(t *testing.T) {
	I := NewGomegaWithT(t)
	controller := gomock.NewController(t)
	defer controller.Finish()
	vcs = new(mocks.Vcs)
	vcsMock = vcs.(*mocks.Vcs)
	defer vcsMock.AssertExpectations(t)
	vcsMock.On("Status", []string{"--porcelain"}).Return("", nil)

	changes, err := getUncommittedChanges(vcs)

	I.Expect(err).To(BeNil())
	I.Expect(changes).To(Equal([]FileChange{}))
}

func TestChangedFiles(t *testing.T) {
	I := NewGomegaWithT(t)
	controller := gomock.NewController(t)
	defer controller.Finish()
	vcs = new(mocks.Vcs)
	vcsMock = vcs.(*mocks.Vcs)
	defer vcsMock.AssertExpectations(t)
	vcsMock.On("Diff", []string{"--name-status", "origin/master..HEAD"}).Return(`D	main.go
D	header_test.go
A	license-header.txt`, nil)
	vcsMock.On("Status", []string{"--porcelain"}).Return(` A main.go
M  license-header.txt
?? git.go`, nil)
	vcsMock.On("Log", []string{"--format=%at", "--", "license-header.txt"}).
		Return("1483228800\n1483228900\n", nil)
	vcsMock.On("Log", []string{"--format=%at", "--", "main.go"}).
		Return("1514764800\n1514764900\n", nil)
	vcsMock.On("Log", []string{"--format=%at", "--", "git.go"}).
		Return("1546300800\n1546300900\n", nil)

	changes, err := GetVcsChanges(vcs, "origin", "master", false)

	I.Expect(err).To(BeNil())
	I.Expect(changes).To(ContainElement(FileChange{
		Path: "license-header.txt", CreationYear: 2017, LastEditionYear: 2017}))
	I.Expect(changes).To(ContainElement(FileChange{
		Path: "main.go", CreationYear: 2018, LastEditionYear: 2018}))
	I.Expect(changes).To(ContainElement(FileChange{
		Path: "git.go", CreationYear: 2019, LastEditionYear: 2019}))
}

type FakeTime struct {
	timestamp int64
}

func (t FakeTime) Now() time.Time {
	return time.Unix(t.timestamp, 0)
}

func TestGetFileDates(t *testing.T) {
	I := NewGomegaWithT(t)
	controller := gomock.NewController(t)
	defer controller.Finish()
	vcs = new(mocks.Vcs)
	vcsMock = vcs.(*mocks.Vcs)
	defer vcsMock.AssertExpectations(t)
	vcsMock.On("Log", []string{"--format=%at", "--", "somefile.go"}).Return(`1537974554
1537973963
1537970000
1537846444
1537844925
1499817600
`, nil)

	history, err := getFileHistory(vcs, "somefile.go", FakeTime{})

	I.Expect(err).To(BeNil())
	I.Expect(history.CreationYear).To(Equal(2017))
	I.Expect(history.LastEditionYear).To(Equal(2018))
}

const fakeNow = 510278400 // 4th of March, 1986

func TestGetFileDatesWithoutDates(t *testing.T) {
	I := NewGomegaWithT(t)
	controller := gomock.NewController(t)
	defer controller.Finish()
	vcs = new(mocks.Vcs)
	vcsMock = vcs.(*mocks.Vcs)
	defer vcsMock.AssertExpectations(t)
	vcsMock.On("Log", []string{"--format=%at", "--", "somefile.go"}).Return(``, nil)

	history, err := getFileHistory(vcs, "somefile.go", FakeTime{timestamp: fakeNow})

	I.Expect(err).To(BeNil())
	I.Expect(history.CreationYear).To(Equal(1986))
	I.Expect(history.LastEditionYear).To(Equal(1986))
}

func TestGetFileDatesWithOnlyOneDate(t *testing.T) {
	I := NewGomegaWithT(t)
	controller := gomock.NewController(t)
	defer controller.Finish()
	vcs = new(mocks.Vcs)
	vcsMock = vcs.(*mocks.Vcs)
	defer vcsMock.AssertExpectations(t)
	vcsMock.On("Log", []string{"--format=%at", "--", "somefile.go"}).Return(`405561600
`, nil)

	history, err := getFileHistory(vcs, "somefile.go", FakeTime{timestamp: fakeNow})

	I.Expect(err).To(BeNil())
	I.Expect(history.CreationYear).To(Equal(1982))
	I.Expect(history.LastEditionYear).To(Equal(1986))
}
