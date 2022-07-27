// Package mpd implements parsing and generating of MPEG-DASH Media Presentation Description (MPD) files.
package mpd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strconv"

	copyobj "github.com/RamanPndy/go-dash-mpd/utils"
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
	Cenc                       *string  `xml:"cenc,attr,omitempty"`
	SCTE35                     *string  `xml:"scte35,attr,omitempty"`
	XSISchemaLocation          *string  `xml:"schemaLocation,attr"`
	ID                         *string  `xml:"id,attr"`
	BaseURL                    *string  `xml:"BaseURL,omitempty"`
	Period                     []Period `xml:"Period,omitempty"`
}

// MPD represents root XML element for Marshal.
type mpdMarshal struct {
	XMLName                    xml.Name        `xml:"MPD"`
	XSI                        *string         `xml:"xmlns:xsi,attr,omitempty"`
	Cenc                       *string         `xml:"xmlns:cenc,attr,omitempty"`
	XMLNS                      *string         `xml:"xmlns,attr"`
	XSISchemaLocation          *string         `xml:"xsi:schemaLocation,attr"`
	ID                         *string         `xml:"id,attr"`
	Type                       *string         `xml:"type,attr"`
	PublishTime                *string         `xml:"publishTime,attr"`
	MinimumUpdatePeriod        *string         `xml:"minimumUpdatePeriod,attr"`
	AvailabilityStartTime      *string         `xml:"availabilityStartTime,attr"`
	MediaPresentationDuration  *string         `xml:"mediaPresentationDuration,attr"`
	MinBufferTime              *string         `xml:"minBufferTime,attr"`
	SuggestedPresentationDelay *string         `xml:"suggestedPresentationDelay,attr"`
	TimeShiftBufferDepth       *string         `xml:"timeShiftBufferDepth,attr"`
	Profiles                   string          `xml:"profiles,attr"`
	SCTE35                     *string         `xml:"xmlns:scte35,attr,omitempty"`
	BaseURL                    *string         `xml:"BaseURL,omitempty"`
	Period                     []periodMarshal `xml:"Period,omitempty"`
}

// Do not try to use encoding.TextMarshaler and encoding.TextUnmarshaler:
// https://github.com/golang/go/issues/6859#issuecomment-118890463

// Encode generates MPD XML.
func (m *MPD) Encode() ([]byte, error) {
	x := new(bytes.Buffer)
	e := xml.NewEncoder(x)
	e.Indent("", "  ")

	xml := modifyMPD(m)

	err := e.Encode(xml)
	if err != nil {
		return nil, err
	}

	// hacks for self-closing tags
	res := new(bytes.Buffer)
	res.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
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
type periodMarshal struct {
	Start          *string                 `xml:"start,attr"`
	ID             *string                 `xml:"id,attr"`
	Duration       *string                 `xml:"duration,attr"`
	AdaptationSets []*adaptationSetMarshal `xml:"AdaptationSet,omitempty"`
}

// AdaptationSet represents XSD's AdaptationSetType.
type AdaptationSet struct {
	Role                    *Role                 `xml:"Role,omitempty"`
	MimeType                string                `xml:"mimeType,attr,omitempty"`
	SegmentAlignment        ConditionalUint       `xml:"segmentAlignment,attr"`
	StartWithSAP            *uint64               `xml:"startWithSAP,attr"`
	BitstreamSwitching      *bool                 `xml:"bitstreamSwitching,attr"`
	SubsegmentAlignment     ConditionalUint       `xml:"subsegmentAlignment,attr"`
	SubsegmentStartsWithSAP *uint64               `xml:"subsegmentStartsWithSAP,attr"`
	Lang                    *string               `xml:"lang,attr"`
	ContentProtections      []DRMDescriptor       `xml:"ContentProtection,omitempty"`
	Representations         []Representation      `xml:"Representation,omitempty"`
	Codecs                  *string               `xml:"codecs,attr"`
	ContentType             string                `xml:"contentType,attr,omitempty"`
	ID                      *string               `xml:"id,attr"`
	Width                   *string               `xml:"width,attr,omitempty"`
	Height                  *string               `xml:"height,attr,omitempty"`
	MaxWidth                *string               `xml:"maxWidth,attr,omitempty"`
	MaxHeight               *string               `xml:"maxHeight,attr,omitempty"`
	FrameRate               *string               `xml:"frameRate,attr,omitempty"`
	Par                     *string               `xml:"par,attr,omitempty"`
	SupplementalProperty    *SupplementalProperty `xml:"SupplementalProperty,omitempty"`
}

type adaptationSetMarshal struct {
	Role                    *roleMarshal                 `xml:"Role,omitempty"`
	MimeType                string                       `xml:"mimeType,attr,omitempty"`
	SegmentAlignment        ConditionalUint              `xml:"segmentAlignment,attr"`
	StartWithSAP            *uint64                      `xml:"startWithSAP,attr"`
	BitstreamSwitching      *bool                        `xml:"bitstreamSwitching,attr"`
	SubsegmentAlignment     ConditionalUint              `xml:"subsegmentAlignment,attr"`
	SubsegmentStartsWithSAP *uint64                      `xml:"subsegmentStartsWithSAP,attr"`
	Lang                    *string                      `xml:"lang,attr"`
	ContentProtections      []drmDescriptorMarshal       `xml:"ContentProtection,omitempty"`
	Representations         []representationMarshal      `xml:"Representation,omitempty"`
	Codecs                  *string                      `xml:"codecs,attr"`
	ContentType             string                       `xml:"contentType,attr,omitempty"`
	ID                      *string                      `xml:"id,attr"`
	Width                   *string                      `xml:"width,attr,omitempty"`
	Height                  *string                      `xml:"height,attr,omitempty"`
	MaxWidth                *string                      `xml:"maxWidth,attr,omitempty"`
	MaxHeight               *string                      `xml:"maxHeight,attr,omitempty"`
	FrameRate               *string                      `xml:"frameRate,attr,omitempty"`
	Par                     *string                      `xml:"par,attr,omitempty"`
	SupplementalProperty    *supplementalPropertyMarshal `xml:"SupplementalProperty,omitempty"`
}

type Role struct {
	SchemeIdUri *string `xml:"schemeIdUri,attr,omitempty"`
	Value       *string `xml:"value,attr,omitempty"`
}

type roleMarshal struct {
	SchemeIdUri *string `xml:"schemeIdUri,attr,omitempty"`
	Value       *string `xml:"value,attr,omitempty"`
}

type SupplementalProperty struct {
	SchemeIdUri *string `xml:"schemeIdUri,attr,omitempty"`
	Value       *string `xml:"value,attr,omitempty"`
}

type supplementalPropertyMarshal struct {
	SchemeIdUri *string `xml:"schemeIdUri,attr,omitempty"`
	Value       *string `xml:"value,attr,omitempty"`
}

// Representation represents XSD's RepresentationType.
type Representation struct {
	ID                        *string                    `xml:"id,attr"`
	Width                     *uint64                    `xml:"width,attr"`
	Height                    *uint64                    `xml:"height,attr"`
	SAR                       *string                    `xml:"sar,attr"`
	FrameRate                 *string                    `xml:"frameRate,attr"`
	Bandwidth                 *uint64                    `xml:"bandwidth,attr"`
	AudioSamplingRate         *string                    `xml:"audioSamplingRate,attr"`
	Codecs                    *string                    `xml:"codecs,attr"`
	BaseURL                   *string                    `xml:"BaseURL,omitempty"`
	ContentProtections        []DRMDescriptor            `xml:"ContentProtection,omitempty"`
	SegmentTemplate           *SegmentTemplate           `xml:"SegmentTemplate,omitempty"`
	MimeType                  string                     `xml:"mimeType,attr,omitempty"`
	AudioChannelConfiguration *AudioChannelConfiguration `xml:"AudioChannelConfiguration,omitempty"`
}

type representationMarshal struct {
	ID                        *string                           `xml:"id,attr"`
	Width                     *uint64                           `xml:"width,attr"`
	Height                    *uint64                           `xml:"height,attr"`
	SAR                       *string                           `xml:"sar,attr"`
	FrameRate                 *string                           `xml:"frameRate,attr"`
	Bandwidth                 *uint64                           `xml:"bandwidth,attr"`
	AudioSamplingRate         *string                           `xml:"audioSamplingRate,attr"`
	Codecs                    *string                           `xml:"codecs,attr"`
	BaseURL                   *string                           `xml:"BaseURL,omitempty"`
	ContentProtections        []drmDescriptorMarshal            `xml:"ContentProtection,omitempty"`
	SegmentTemplate           *SegmentTemplate                  `xml:"SegmentTemplate,omitempty"`
	MimeType                  string                            `xml:"mimeType,attr,omitempty"`
	AudioChannelConfiguration *audioChannelConfigurationMarshal `xml:"AudioChannelConfiguration,omitempty"`
}

type AudioChannelConfiguration struct {
	SchemeIdUri *string `xml:"schemeIdUri,attr,omitempty"`
	Value       *string `xml:"value,attr,omitempty"`
}

type audioChannelConfigurationMarshal struct {
	SchemeIdUri *string `xml:"schemeIdUri,attr,omitempty"`
	Value       *string `xml:"value,attr,omitempty"`
}

// Descriptor represents XSD's DescriptorType.
type DRMDescriptor struct {
	SchemeIDURI    *string `xml:"schemeIdUri,attr"`
	Value          *string `xml:"value,attr,omitempty"`
	CencDefaultKID *string `xml:"default_KID,attr,omitempty"`
	Cenc           *string `xml:"cenc,attr,omitempty"`
	Pssh           *Pssh   `xml:"pssh"`
}

type drmDescriptorMarshal struct {
	SchemeIDURI    *string      `xml:"schemeIdUri,attr"`
	Value          *string      `xml:"value,attr,omitempty"`
	CencDefaultKID *string      `xml:"cenc:default_KID,attr,omitempty"`
	Cenc           *string      `xml:"xmlns:cenc,attr,omitempty"`
	Pssh           *psshMarshal `xml:"cenc:pssh"`
}

// Pssh represents XSD's CencPsshType .
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

// modifyMPD generates true xml struct for MPD .
func modifyMPD(mpd *MPD) *mpdMarshal {
	return &mpdMarshal{
		XMLNS:                      copyobj.String(mpd.XMLNS),
		MinimumUpdatePeriod:        copyobj.String(mpd.MinimumUpdatePeriod),
		AvailabilityStartTime:      copyobj.String(mpd.AvailabilityStartTime),
		MediaPresentationDuration:  copyobj.String(mpd.MediaPresentationDuration),
		MinBufferTime:              copyobj.String(mpd.MinBufferTime),
		SuggestedPresentationDelay: copyobj.String(mpd.SuggestedPresentationDelay),
		TimeShiftBufferDepth:       copyobj.String(mpd.TimeShiftBufferDepth),
		PublishTime:                copyobj.String(mpd.PublishTime),
		Type:                       copyobj.String(mpd.Type),
		Profiles:                   mpd.Profiles,
		XSI:                        copyobj.String(mpd.XSI),
		SCTE35:                     copyobj.String(mpd.SCTE35),
		XSISchemaLocation:          copyobj.String(mpd.XSISchemaLocation),
		Cenc:                       copyobj.String(mpd.Cenc),
		ID:                         copyobj.String(mpd.ID),
		BaseURL:                    copyobj.String(mpd.BaseURL),
		Period:                     modifyPeriod(mpd.Period),
	}
}

func modifyPeriod(ps []Period) []periodMarshal {
	if ps == nil {
		return nil
	}
	pms := make([]periodMarshal, 0, len(ps))
	for _, p := range ps {
		period := periodMarshal{
			Duration:       copyobj.String(p.Duration),
			ID:             copyobj.String(p.ID),
			Start:          copyobj.String(p.Start),
			AdaptationSets: modifyAdaptationSets(p.AdaptationSets),
		}
		pms = append(pms, period)
	}

	return pms
}

func modifyAdaptationSets(as []*AdaptationSet) []*adaptationSetMarshal {
	if as == nil {
		return nil
	}
	asm := make([]*adaptationSetMarshal, 0, len(as))
	for _, a := range as {
		adaptationSet := &adaptationSetMarshal{
			BitstreamSwitching:      copyobj.Bool(a.BitstreamSwitching),
			Codecs:                  copyobj.String(a.Codecs),
			Lang:                    copyobj.String(a.Lang),
			ID:                      copyobj.String(a.ID),
			ContentType:             a.ContentType,
			MimeType:                a.MimeType,
			SegmentAlignment:        a.SegmentAlignment,
			StartWithSAP:            copyobj.UInt64(a.StartWithSAP),
			SubsegmentAlignment:     a.SubsegmentAlignment,
			SubsegmentStartsWithSAP: copyobj.UInt64(a.SubsegmentStartsWithSAP),
			Representations:         modifyRepresentations(a.Representations),
			ContentProtections:      modifyContentProtections(a.ContentProtections),
			Role:                    modifyRole(a.Role),
			Width:                   copyobj.String(a.Width),
			Height:                  copyobj.String(a.Height),
			MaxWidth:                copyobj.String(a.MaxWidth),
			MaxHeight:               copyobj.String(a.MaxHeight),
			Par:                     copyobj.String(a.Par),
			FrameRate:               copyobj.String(a.FrameRate),
			SupplementalProperty:    modifySupplementalProperty(a.SupplementalProperty),
		}
		asm = append(asm, adaptationSet)
	}
	return asm
}

func modifyRepresentations(rs []Representation) []representationMarshal {
	rsm := make([]representationMarshal, 0, len(rs))
	for _, r := range rs {
		representation := representationMarshal{
			AudioSamplingRate:         copyobj.String(r.AudioSamplingRate),
			Bandwidth:                 copyobj.UInt64(r.Bandwidth),
			Codecs:                    copyobj.String(r.Codecs),
			FrameRate:                 copyobj.String(r.FrameRate),
			Height:                    copyobj.UInt64(r.Height),
			ID:                        copyobj.String(r.ID),
			Width:                     copyobj.UInt64(r.Width),
			SegmentTemplate:           copySegmentTemplate(r.SegmentTemplate),
			SAR:                       copyobj.String(r.SAR),
			ContentProtections:        modifyContentProtections(r.ContentProtections),
			BaseURL:                   copyobj.String(r.BaseURL),
			MimeType:                  r.MimeType,
			AudioChannelConfiguration: modifyAudioChannelConfiguration(r.AudioChannelConfiguration),
		}
		rsm = append(rsm, representation)
	}
	return rsm
}

func copySegmentTemplate(st *SegmentTemplate) *SegmentTemplate {
	if st == nil {
		return nil
	}
	return &SegmentTemplate{
		Timescale:              copyobj.UInt64(st.Timescale),
		Media:                  copyobj.String(st.Media),
		Initialization:         copyobj.String(st.Initialization),
		StartNumber:            copyobj.UInt64(st.StartNumber),
		PresentationTimeOffset: copyobj.UInt64(st.PresentationTimeOffset),
		SegmentTimelineS:       copySegmentTimelineS(st.SegmentTimelineS),
	}
}

func copySegmentTimelineS(st []SegmentTimelineS) []SegmentTimelineS {
	stm := make([]SegmentTimelineS, 0, len(st))
	for _, s := range st {
		segmentTimelineS := SegmentTimelineS{
			T: s.T,
			D: s.D,
			R: copyobj.Int64(s.R),
		}
		stm = append(stm, segmentTimelineS)
	}
	return stm
}

func modifyContentProtections(ds []DRMDescriptor) []drmDescriptorMarshal {
	dsm := make([]drmDescriptorMarshal, 0, len(ds))
	for _, d := range ds {
		descriptor := drmDescriptorMarshal{
			CencDefaultKID: copyobj.String(d.CencDefaultKID),
			SchemeIDURI:    copyobj.String(d.SchemeIDURI),
			Value:          copyobj.String(d.Value),
			Cenc:           copyobj.String(d.Cenc),
			Pssh:           modifyPssh(d.Pssh),
		}
		dsm = append(dsm, descriptor)
	}
	return dsm
}

func modifyPssh(p *Pssh) *psshMarshal {
	if p == nil {
		return nil
	}
	return &psshMarshal{
		Cenc:  copyobj.String(p.Cenc),
		Value: copyobj.String(p.Value),
	}
}

func modifyRole(r *Role) *roleMarshal {
	if r == nil {
		return nil
	}
	return &roleMarshal{
		SchemeIdUri: copyobj.String(r.SchemeIdUri),
		Value:       copyobj.String(r.Value),
	}
}

func modifySupplementalProperty(s *SupplementalProperty) *supplementalPropertyMarshal {
	if s == nil {
		return nil
	}
	return &supplementalPropertyMarshal{
		SchemeIdUri: copyobj.String(s.SchemeIdUri),
		Value:       copyobj.String(s.Value),
	}
}

func modifyAudioChannelConfiguration(a *AudioChannelConfiguration) *audioChannelConfigurationMarshal {
	if a == nil {
		return nil
	}
	return &audioChannelConfigurationMarshal{
		SchemeIdUri: copyobj.String(a.SchemeIdUri),
		Value:       copyobj.String(a.Value),
	}
}
