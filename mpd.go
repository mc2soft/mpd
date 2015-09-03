// Package mpd implements parsing and generating of MPEG-DASH Media Presentation Description (MPD) files.
package mpd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strconv"
)

// http://mpeg.chiariglione.org/standards/mpeg-dash
// https://www.brendanlong.com/the-structure-of-an-mpeg-dash-mpd.html

var emptyElementRE = regexp.MustCompile(`></[A-Za-z]+>`)

// from XSD
// Conditional Unsigned Integer (unsignedInt or boolean)
type ConditionalUintType struct {
	u *uint64
	b *bool
}

func (c ConditionalUintType) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	if c.u != nil {
		return xml.Attr{Name: name, Value: strconv.FormatUint(*c.u, 10)}, nil
	}

	if c.b != nil {
		return xml.Attr{Name: name, Value: strconv.FormatBool(*c.b)}, nil
	}

	// both are nil - no attribute, client will threat it like "false"
	return xml.Attr{}, nil
}

func (c *ConditionalUintType) UnmarshalXMLAttr(attr xml.Attr) error {
	u, err := strconv.ParseUint(attr.Value, 10, 64)
	if err == nil {
		c.u = &u
		return nil
	}

	b, err := strconv.ParseBool(attr.Value)
	if err == nil {
		c.b = &b
		return nil
	}

	return fmt.Errorf("ConditionalUintType: can't UnmarshalXMLAttr %#v", attr)
}

// check interfaces
var (
	_ xml.MarshalerAttr   = ConditionalUintType{}
	_ xml.UnmarshalerAttr = &ConditionalUintType{}
)

type MPD struct {
	XMLNS                      *string `xml:"xmlns,attr"`
	Type                       *string `xml:"type,attr"`
	MinimumUpdatePeriod        *string `xml:"minimumUpdatePeriod,attr"`
	AvailabilityStartTime      *string `xml:"availabilityStartTime,attr"`
	MediaPresentationDuration  *string `xml:"mediaPresentationDuration,attr"`
	MinBufferTime              *string `xml:"minBufferTime,attr"`
	SuggestedPresentationDelay *string `xml:"suggestedPresentationDelay,attr"`
	TimeShiftBufferDepth       *string `xml:"timeShiftBufferDepth,attr"`
	Profiles                   string  `xml:"profiles,attr"`
	Period                     *Period `xml:"Period,omitempty"`
}

// Do not try to use encoding.TextMarshaler and encoding.TextUnmarshaler:
// https://github.com/golang/go/issues/6859#issuecomment-118890463

func (m *MPD) Encode() ([]byte, error) {
	x := new(bytes.Buffer)
	e := xml.NewEncoder(x)
	e.Indent("", "  ")
	err := e.Encode(m)
	if err != nil {
		return nil, err
	}

	res := new(bytes.Buffer)
	res.WriteString(`<?xml version="1.0" encoding="utf-8"?>`)
	res.WriteByte('\n')
	for {
		s, err := x.ReadString('\n')
		if s != "" {
			s = emptyElementRE.ReplaceAllString(s, `/>`)
			res.WriteString(s)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	res.WriteByte('\n')
	return res.Bytes(), err
}

func (m *MPD) Decode(b []byte) error {
	return xml.Unmarshal(b, m)
}

type Period struct {
	Start          *string          `xml:"start,attr"`
	ID             *string          `xml:"id,attr"`
	Duration       *string          `xml:"duration,attr"`
	AdaptationSets []*AdaptationSet `xml:"AdaptationSet,omitempty"`
}

type AdaptationSet struct {
	MimeType                string              `xml:"mimeType,attr"`
	SegmentAlignment        ConditionalUintType `xml:"segmentAlignment,attr"`
	SubsegmentAlignment     ConditionalUintType `xml:"subsegmentAlignment,attr"`
	StartWithSAP            *uint64             `xml:"startWithSAP,attr"`
	SubsegmentStartsWithSAP *uint64             `xml:"subsegmentStartsWithSAP,attr"`
	BitstreamSwitching      *bool               `xml:"bitstreamSwitching,attr"`
	Lang                    *string             `xml:"lang,attr"`
	Representations         []Representation    `xml:"Representation,omitempty"`
}

type Representation struct {
	ID                 *string             `xml:"id,attr"`
	Width              *uint64             `xml:"width,attr"`
	Height             *uint64             `xml:"height,attr"`
	FrameRate          *string             `xml:"frameRate,attr"`
	Bandwidth          *uint64             `xml:"bandwidth,attr"`
	AudioSamplingRate  *string             `xml:"audioSamplingRate,attr"`
	Codecs             *string             `xml:"codecs,attr"`
	ContentProtections []ContentProtection `xml:"ContentProtection,omitempty"`
	SegmentTemplate    *SegmentTemplate    `xml:"SegmentTemplate,omitempty"`
}

type ContentProtection struct {
	SchemeIDURI *string `xml:"schemeIdUri,attr"`
	Value       *string `xml:"value,attr"`
}

type SegmentTemplate struct {
	Timescale              *uint64            `xml:"timescale,attr"`
	Media                  *string            `xml:"media,attr"`
	Initialization         *string            `xml:"initialization,attr"`
	StartNumber            *uint64            `xml:"startNumber,attr"`
	PresentationTimeOffset *uint64            `xml:"presentationTimeOffset,attr"`
	SegmentTimelineS       []SegmentTimelineS `xml:"SegmentTimeline>S,omitempty"`
}

type SegmentTimelineS struct {
	T *uint64 `xml:"t,attr"`
	D uint64  `xml:"d,attr"`
	R *int64  `xml:"r,attr"`
}
