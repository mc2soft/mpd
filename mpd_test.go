package mpd

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MPDSuite struct{}

var _ = Suite(&MPDSuite{})

func testUnmarshalMarshal(c *C, name string) {
	fmt.Println(name)
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

	obtainedSlice := strings.Split(strings.TrimSpace(string(obtained)), "\n")
	expectedSlice := strings.Split(strings.TrimSpace(expectedS), "\n")
	c.Check(obtainedSlice, HasLen, len(expectedSlice))
	for i := range obtainedSlice {
		c.Check(obtainedSlice[i], Equals, expectedSlice[i], Commentf("line %d", i+1))
	}
}

func (s *MPDSuite) TestUnmarshalMarshalVod(c *C) {
	testUnmarshalMarshal(c, "fixture_elemental_delta_vod.mpd")
}

func (s *MPDSuite) TestUnmarshalMarshalLive(c *C) {
	testUnmarshalMarshal(c, "fixture_elemental_delta_live.mpd")
}

func (s *MPDSuite) TestUnmarshalMarshalLiveDelta161(c *C) {
	testUnmarshalMarshal(c, "fixture_elemental_delta_vod_multi_drm.mpd")
}

func (s *MPDSuite) TestUnmarshalMarshalLiveSimVod(c *C) {
	testUnmarshalMarshal(c, "fixture_livesim_vod.mpd")
}

func TestMPDEqual(t *testing.T) {
	mpd := &MPD{}
	mpdM := &mpdMarshal{}
	require.Equal(t, 18, reflect.ValueOf(mpd).Elem().NumField(),
		"model was updated, need to update this test and function modifyMPD")
	require.Equal(t, reflect.ValueOf(mpd).Elem().NumField(), reflect.ValueOf(mpdM).Elem().NumField(),
		"MPD element count not equal mpdMarshal")
}

func TestPeriodEqual(t *testing.T) {
	mpd := &Period{}
	mpdM := &periodMarshal{}
	require.Equal(t, 4, reflect.ValueOf(mpd).Elem().NumField(),
		"model was updated, need to update this test and function modifyPeriod")
	require.Equal(t, reflect.ValueOf(mpd).Elem().NumField(), reflect.ValueOf(mpdM).Elem().NumField(),
		"Period element count not equal periodMarshal")
}

func TestAdaptationSetEqual(t *testing.T) {
	mpd := &AdaptationSet{}
	mpdM := &adaptationSetMarshal{}
	require.Equal(t, 18, reflect.ValueOf(mpd).Elem().NumField(),
		"model was updated, need to update this test and function modifyAdaptationSets")
	require.Equal(t, reflect.ValueOf(mpd).Elem().NumField(), reflect.ValueOf(mpdM).Elem().NumField(),
		"AdaptationSet element count not equal adaptationSetMarshal")
}

func TestRepresentationEqual(t *testing.T) {
	a := &Representation{}
	b := &Representation{}
	require.Equal(t, 11, reflect.ValueOf(a).Elem().NumField(),
		"model was updated, need to update this test and function modifyRepresentations")
	require.Equal(t, reflect.ValueOf(a).Elem().NumField(), reflect.ValueOf(b).Elem().NumField(),
		"Representation element count not equal Representation")
}

func TestSegmentTemplateEqual(t *testing.T) {
	a := &SegmentTemplate{}
	require.Equal(t, 7, reflect.ValueOf(a).Elem().NumField(),
		"model was updated, need to update this test and function copySegmentTemplate")
}

func TestSegmentTimelineSEqual(t *testing.T) {
	a := &SegmentTimelineS{}
	require.Equal(t, 3, reflect.ValueOf(a).Elem().NumField(),
		"model was updated, need to update this test and function copySegmentTimelineS")
}

func TestDescriptorEqual(t *testing.T) {
	a := &Descriptor{}
	b := &descriptorMarshal{}
	require.Equal(t, 5, reflect.ValueOf(a).Elem().NumField(),
		"model was updated, need to update this test and function modifyContentProtections")
	require.Equal(t, reflect.ValueOf(a).Elem().NumField(), reflect.ValueOf(b).Elem().NumField(),
		"Descriptor element count not equal descriptorMarshal")
}

func TestPsshEqual(t *testing.T) {
	a := &Pssh{}
	b := &psshMarshal{}
	require.Equal(t, 2, reflect.ValueOf(a).Elem().NumField(),
		"model was updated, need to update this test and function modifyPssh")
	require.Equal(t, reflect.ValueOf(a).Elem().NumField(), reflect.ValueOf(b).Elem().NumField(),
		"Pssh element count not equal psshMarshal")
}
