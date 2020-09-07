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
// http://standards.iso.org/ittf/PubliclyAvailableStandards/MPEG-DASH_schema_files/DASH-MPD.xsd

var emptyElementRE = regexp.MustCompile(`></[A-Za-z]+>`)

// ConditionalUint (ConditionalUintType) defined in XSD as a union of unsignedInt and boolean.
type ConditionalUint struct {
	u *uint64
	b *bool
}

// MarshalXMLAttr encodes ConditionalUint.
func (c ConditionalUint) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	if c.u != nil {
		return xml.Attr{Name: name, Value: strconv.FormatUint(*c.u, 10)}, nil
	}

	if c.b != nil {
		return xml.Attr{Name: name, Value: strconv.FormatBool(*c.b)}, nil
	}

	// both are nil - no attribute, client will threat it like "false"
	return xml.Attr{}, nil
}

// UnmarshalXMLAttr decodes ConditionalUint.
func (c *ConditionalUint) UnmarshalXMLAttr(attr xml.Attr) error {
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

	return fmt.Errorf("ConditionalUint: can't UnmarshalXMLAttr %#v", attr)
}

// check interfaces
var (
	_ xml.MarshalerAttr   = ConditionalUint{}
	_ xml.UnmarshalerAttr = &ConditionalUint{}
)

// MPD represents root XML element for parse.
type MPD struct {
	XMLName                    xml.Name `xml:"MPD"`
	XMLNS                      *string  `xml:"xmlns,attr"`
	Type                       *string  `xml:"type,attr"`
	MinimumUpdatePeriod        *string  `xml:"minimumUpdatePeriod,attr"`
	AvailabilityStartTime      *string  `xml:"availabilityStartTime,attr"`
	MediaPresentationDuration  *string  `xml:"mediaPresentationDuration,attr"`
	MinBufferTime              *string  `xml:"minBufferTime,attr"`
	SuggestedPresentationDelay *string  `xml:"suggestedPresentationDelay,attr"`
	TimeShiftBufferDepth       *string  `xml:"timeShiftBufferDepth,attr"`
	PublishTime                *string  `xml:"publishTime,attr"`
	Profiles                   string   `xml:"profiles,attr"`
	XSI                        *string  `xml:"xsi,attr,omitempty"`
	SCTE35                     *string  `xml:"scte35,attr,omitempty"`
	XSISchemaLocation          *string  `xml:"schemaLocation,attr"`
	ID                         *string  `xml:"id,attr"`
	Period                     *Period  `xml:"Period,omitempty"`
}

// MPD represents root XML element for Marshal.
type mpdMarshal struct {
	XMLName                    xml.Name       `xml:"MPD"`
	XSI                        *string        `xml:"xmlns:xsi,attr,omitempty"`
	XMLNS                      *string        `xml:"xmlns,attr"`
	XSISchemaLocation          *string        `xml:"xsi:schemaLocation,attr"`
	ID                         *string        `xml:"id,attr"`
	Type                       *string        `xml:"type,attr"`
	PublishTime                *string        `xml:"publishTime,attr"`
	MinimumUpdatePeriod        *string        `xml:"minimumUpdatePeriod,attr"`
	AvailabilityStartTime      *string        `xml:"availabilityStartTime,attr"`
	MediaPresentationDuration  *string        `xml:"mediaPresentationDuration,attr"`
	MinBufferTime              *string        `xml:"minBufferTime,attr"`
	SuggestedPresentationDelay *string        `xml:"suggestedPresentationDelay,attr"`
	TimeShiftBufferDepth       *string        `xml:"timeShiftBufferDepth,attr"`
	Profiles                   string         `xml:"profiles,attr"`
	SCTE35                     *string        `xml:"xmlns:scte35,attr,omitempty"`
	Period                     *PeriodMarshal `xml:"Period,omitempty"`
}

// Do not try to use encoding.TextMarshaler and encoding.TextUnmarshaler:
// https://github.com/golang/go/issues/6859#issuecomment-118890463

// Encode generates MPD XML.
func (m *MPD) Encode() ([]byte, error) {
	x := new(bytes.Buffer)
	e := xml.NewEncoder(x)
	e.Indent("", "  ")

	xml := modifyXMLStuct(m)

	err := e.Encode(xml)
	if err != nil {
		return nil, err
	}

	// hacks for self-closing tags
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

// Decode parses MPD XML.
func (m *MPD) Decode(b []byte) error {
	return xml.Unmarshal(b, m)
}

// Period represents XSD's PeriodType.
type Period struct {
	Start          *string          `xml:"start,attr"`
	ID             *string          `xml:"id,attr"`
	Duration       *string          `xml:"duration,attr"`
	AdaptationSets []*AdaptationSet `xml:"AdaptationSet,omitempty"`
}

// Period represents XSD's PeriodType.
type PeriodMarshal struct {
	Start          *string                 `xml:"start,attr"`
	ID             *string                 `xml:"id,attr"`
	Duration       *string                 `xml:"duration,attr"`
	AdaptationSets []*adaptationSetMarshal `xml:"AdaptationSet,omitempty"`
}

// AdaptationSet represents XSD's AdaptationSetType.
type AdaptationSet struct {
	MimeType                string           `xml:"mimeType,attr"`
	SegmentAlignment        ConditionalUint  `xml:"segmentAlignment,attr"`
	StartWithSAP            *uint64          `xml:"startWithSAP,attr"`
	BitstreamSwitching      *bool            `xml:"bitstreamSwitching,attr"`
	SubsegmentAlignment     ConditionalUint  `xml:"subsegmentAlignment,attr"`
	SubsegmentStartsWithSAP *uint64          `xml:"subsegmentStartsWithSAP,attr"`
	Lang                    *string          `xml:"lang,attr"`
	Representations         []Representation `xml:"Representation,omitempty"`
	Codecs                  *string          `xml:"codecs,attr"`
}

type adaptationSetMarshal struct {
	MimeType                string                  `xml:"mimeType,attr"`
	SegmentAlignment        ConditionalUint         `xml:"segmentAlignment,attr"`
	StartWithSAP            *uint64                 `xml:"startWithSAP,attr"`
	BitstreamSwitching      *bool                   `xml:"bitstreamSwitching,attr"`
	SubsegmentAlignment     ConditionalUint         `xml:"subsegmentAlignment,attr"`
	SubsegmentStartsWithSAP *uint64                 `xml:"subsegmentStartsWithSAP,attr"`
	Lang                    *string                 `xml:"lang,attr"`
	Representations         []representationMarshal `xml:"Representation,omitempty"`
	Codecs                  *string                 `xml:"codecs,attr"`
}

// Representation represents XSD's RepresentationType.
type Representation struct {
	ID                 *string          `xml:"id,attr"`
	Width              *uint64          `xml:"width,attr"`
	Height             *uint64          `xml:"height,attr"`
	SAR                *string          `xml:"sar,attr"`
	FrameRate          *string          `xml:"frameRate,attr"`
	Bandwidth          *uint64          `xml:"bandwidth,attr"`
	AudioSamplingRate  *string          `xml:"audioSamplingRate,attr"`
	Codecs             *string          `xml:"codecs,attr"`
	ContentProtections []Descriptor     `xml:"ContentProtection,omitempty"`
	SegmentTemplate    *SegmentTemplate `xml:"SegmentTemplate,omitempty"`
}

type representationMarshal struct {
	ID                 *string             `xml:"id,attr"`
	Width              *uint64             `xml:"width,attr"`
	Height             *uint64             `xml:"height,attr"`
	SAR                *string             `xml:"sar,attr"`
	FrameRate          *string             `xml:"frameRate,attr"`
	Bandwidth          *uint64             `xml:"bandwidth,attr"`
	AudioSamplingRate  *string             `xml:"audioSamplingRate,attr"`
	Codecs             *string             `xml:"codecs,attr"`
	ContentProtections []descriptorMarshal `xml:"ContentProtection,omitempty"`
	SegmentTemplate    *SegmentTemplate    `xml:"SegmentTemplate,omitempty"`
}

// Descriptor represents XSD's DescriptorType.
type Descriptor struct {
	SchemeIDURI    *string `xml:"schemeIdUri,attr"`
	Value          *string `xml:"value,attr,omitempty"`
	CencDefaultKID *string `xml:"default_KID,attr,omitempty"`
	Cenc           *string `xml:"cenc,attr,omitempty"`
	Pssh           *Pssh   `xml:"pssh"`
}

type descriptorMarshal struct {
	XMLName        xml.Name     `xml:"ContentProtection"`
	SchemeIDURI    *string      `xml:"schemeIdUri,attr"`
	Value          *string      `xml:"value,attr,omitempty"`
	CencDefaultKID *string      `xml:"cenc:default_KID,attr,omitempty"`
	Cenc           *string      `xml:"xmlns:cenc,attr,omitempty"`
	Pssh           *psshMarshal `xml:"cenc:pssh"`
}

// CencPssh represents XSD's CencPsshType .
type Pssh struct {
	Cenc  *string `xml:"cenc,attr"`
	Value *string `xml:",chardata"`
}

type psshMarshal struct {
	Cenc  *string `xml:"xmlns:cenc,attr"`
	Value *string `xml:",chardata"`
}

// SegmentTemplate represents XSD's SegmentTemplateType.
type SegmentTemplate struct {
	Timescale              *uint64            `xml:"timescale,attr"`
	Media                  *string            `xml:"media,attr"`
	Initialization         *string            `xml:"initialization,attr"`
	StartNumber            *uint64            `xml:"startNumber,attr"`
	PresentationTimeOffset *uint64            `xml:"presentationTimeOffset,attr"`
	SegmentTimelineS       []SegmentTimelineS `xml:"SegmentTimeline>S,omitempty"`
}

// SegmentTimelineS represents XSD's SegmentTimelineType's inner S elements.
type SegmentTimelineS struct {
	T *uint64 `xml:"t,attr"`
	D uint64  `xml:"d,attr"`
	R *int64  `xml:"r,attr"`
}

// modifyXMLStuct generates true MPD .
func modifyXMLStuct(mpd *MPD) *mpdMarshal {
	mpdMarshal := new(mpdMarshal)

	// MPD
	if mpd.XMLNS != nil {
		mpdMarshal.XMLNS = mpd.XMLNS
	}
	if mpd.MinimumUpdatePeriod != nil {
		mpdMarshal.MinimumUpdatePeriod = mpd.MinimumUpdatePeriod
	}
	if mpd.AvailabilityStartTime != nil {
		mpdMarshal.AvailabilityStartTime = mpd.AvailabilityStartTime
	}
	if mpd.MediaPresentationDuration != nil {
		mpdMarshal.MediaPresentationDuration = mpd.MediaPresentationDuration
	}
	if mpd.MinBufferTime != nil {
		mpdMarshal.MinBufferTime = mpd.MinBufferTime
	}
	if mpd.SuggestedPresentationDelay != nil {
		mpdMarshal.SuggestedPresentationDelay = mpd.SuggestedPresentationDelay
	}
	if mpd.TimeShiftBufferDepth != nil {
		mpdMarshal.TimeShiftBufferDepth = mpd.TimeShiftBufferDepth
	}
	if mpd.PublishTime != nil {
		mpdMarshal.PublishTime = mpd.PublishTime
	}
	if mpd.Type != nil {
		mpdMarshal.Type = mpd.Type
	}
	mpdMarshal.Profiles = mpd.Profiles
	if mpd.XSI != nil {
		mpdMarshal.XSI = mpd.XSI
	}
	if mpd.SCTE35 != nil {
		mpdMarshal.SCTE35 = mpd.SCTE35
	}
	if mpd.XSISchemaLocation != nil {
		mpdMarshal.XSISchemaLocation = mpd.XSISchemaLocation
	}
	if mpd.ID != nil {
		mpdMarshal.ID = mpd.ID
	}

	// Period
	mpdMarshal.Period = &PeriodMarshal{}
	if mpd.Period != nil {
		if mpd.Period.Duration != nil {
			mpdMarshal.Period.Duration = mpd.Period.Duration
		}
		if mpd.Period.ID != nil {
			mpdMarshal.Period.ID = mpd.Period.ID
		}
		if mpd.Period.Start != nil {
			mpdMarshal.Period.Start = mpd.Period.Start
		}

		if mpd.Period.AdaptationSets != nil {
			// AdaptationSets
			for _, as := range mpd.Period.AdaptationSets {
				adaptationSet := &adaptationSetMarshal{}
				adaptationSet.BitstreamSwitching = as.BitstreamSwitching
				adaptationSet.Codecs = as.Codecs
				adaptationSet.Lang = as.Lang
				adaptationSet.MimeType = as.MimeType
				adaptationSet.SegmentAlignment = as.SegmentAlignment
				adaptationSet.StartWithSAP = as.StartWithSAP
				adaptationSet.SubsegmentAlignment = as.SubsegmentAlignment
				adaptationSet.SubsegmentStartsWithSAP = as.SubsegmentStartsWithSAP

				if as.Representations != nil {
					// Representations
					for _, r := range as.Representations {
						representation := representationMarshal{}
						representation.AudioSamplingRate = r.AudioSamplingRate
						representation.Bandwidth = r.Bandwidth
						representation.Codecs = r.Codecs
						representation.FrameRate = r.FrameRate
						representation.Height = r.Height
						representation.ID = r.ID
						representation.Width = r.Width
						representation.SegmentTemplate = r.SegmentTemplate
						representation.SAR = r.SAR

						if r.ContentProtections != nil {
							// ContentProtections
							for _, cp := range r.ContentProtections {
								descriptorMarshal := descriptorMarshal{}
								descriptorMarshal.CencDefaultKID = cp.CencDefaultKID
								descriptorMarshal.SchemeIDURI = cp.SchemeIDURI
								descriptorMarshal.Value = cp.Value
								descriptorMarshal.Cenc = cp.Cenc
								if cp.Pssh != nil {
									pssh := &psshMarshal{}
									pssh.Cenc = cp.Pssh.Cenc
									pssh.Value = cp.Pssh.Value
									descriptorMarshal.Pssh = pssh
								}
								representation.ContentProtections = append(representation.ContentProtections, descriptorMarshal)
							}
						}
						adaptationSet.Representations = append(adaptationSet.Representations, representation)
					}
				}
				mpdMarshal.Period.AdaptationSets = append(mpdMarshal.Period.AdaptationSets, adaptationSet)
			}
		}
	}

	return mpdMarshal
}
