package mpd

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MPDSuite struct{}

var _ = Suite(&MPDSuite{})

func readFile(c *C, name string) (*MPD, string, string) {
	expected, err := ioutil.ReadFile(name)
	c.Assert(err, IsNil)

	mpd := new(MPD)
	err = mpd.Decode(expected)
	c.Assert(err, IsNil)

	obtained, err := mpd.Encode()
	c.Assert(err, IsNil)
	obtainedName := name + ".ignore"
	err = ioutil.WriteFile(obtainedName, obtained, 0666)
	c.Assert(err, IsNil)

	os.Remove(obtainedName)

	return mpd, string(expected), string(obtained)
}

func checkLineByLine(c *C, obtained string, expected string) {
	obtainedSlice := strings.Split(strings.TrimSpace(obtained), "\n")
	expectedSlice := strings.Split(strings.TrimSpace(expected), "\n")
	c.Assert(obtainedSlice, HasLen, len(expectedSlice))

	for i := range obtainedSlice {
		c.Check(obtainedSlice[i], Equals, expectedSlice[i], Commentf("line %d", i+1))
	}
}

func testUnmarshalMarshalElemental(c *C, name string) {
	_, expected, obtained := readFile(c, name)

	// strip stupid XML rubbish
	expected = strings.Replace(expected, `xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" `, ``, 1)
	expected = strings.Replace(expected, `xsi:schemaLocation="urn:mpeg:dash:schema:mpd:2011 http://standards.iso.org/ittf/PubliclyAvailableStandards/MPEG-DASH_schema_files/DASH-MPD.xsd" `, ``, 1)

	checkLineByLine(c, obtained, expected)
}

func testUnmarshalMarshalAkamai(c *C, name string) {
	_, expected, obtained := readFile(c, name)

	// strip stupid XML rubbish
	expected = strings.Replace(expected, `xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" `, ``, 1)
	expected = strings.Replace(expected, ` xsi:schemaLocation="urn:mpeg:DASH:schema:MPD:2011 DASH-MPD.xsd"`, ``, 1)

	checkLineByLine(c, obtained, expected)
}

func (s *MPDSuite) TestUnmarshalMarshalVod(c *C) {
	testUnmarshalMarshalElemental(c, "fixtures/elemental_delta_vod.mpd")
}

func (s *MPDSuite) TestUnmarshalMarshalLive(c *C) {
	testUnmarshalMarshalElemental(c, "fixtures/elemental_delta_live.mpd")
}

func (s *MPDSuite) TestUnmarshalMarshalLiveDelta161(c *C) {
	testUnmarshalMarshalElemental(c, "fixtures/elemental_delta_1.6.1_live.mpd")
}

func (s *MPDSuite) TestUnmarshalMarshalSegmentTemplate(c *C) {
	testUnmarshalMarshalAkamai(c, "fixtures/akamai_bbb_30fps.mpd")
}
