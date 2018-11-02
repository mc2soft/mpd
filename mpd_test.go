package mpd

import (
	"io/ioutil"
	"strings"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MPDSuite struct{}

var _ = Suite(&MPDSuite{})

func testUnmarshalMarshal(c *C, name string) {
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

	// strip stupid XML rubish
	expectedS := string(expected)
	expectedS = strings.Replace(expectedS, `xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" `, ``, 1)
	expectedS = strings.Replace(expectedS, `xsi:schemaLocation="urn:mpeg:dash:schema:mpd:2011 http://standards.iso.org/ittf/PubliclyAvailableStandards/MPEG-DASH_schema_files/DASH-MPD.xsd" `, ``, 1)

	obtainedSlice := strings.Split(strings.TrimSpace(string(obtained)), "\n")
	expectedSlice := strings.Split(strings.TrimSpace(expectedS), "\n")
	c.Check(obtainedSlice, HasLen, len(expectedSlice))
	for i := range obtainedSlice {
		c.Check(obtainedSlice[i], Equals, expectedSlice[i], Commentf("line %d", i+1))
	}
}

func (s *MPDSuite) TestUnmarshalMarshalVod(c *C) {
	testUnmarshalMarshal(c, "fixtures/elemental_delta_vod.mpd")
}

func (s *MPDSuite) TestUnmarshalMarshalLive(c *C) {
	testUnmarshalMarshal(c, "fixtures/elemental_delta_live.mpd")
}

func (s *MPDSuite) TestUnmarshalMarshalLiveDelta161(c *C) {
	testUnmarshalMarshal(c, "fixtures/elemental_delta_1.6.1_live.mpd")
}
