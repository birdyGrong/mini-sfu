package sfu

import (
	"fmt"
	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v3"
	"strconv"
	"strings"
)

const (
	mediaNameAudio = "audio"
	mediaNameVideo = "video"
)

var (
	rtcpfb = []webrtc.RTCPFeedback{
		{Type: webrtc.TypeRTCPFBGoogREMB},
		{Type: webrtc.TypeRTCPFBCCM},
		{Type: webrtc.TypeRTCPFBNACK},
		{Type: "nack pli"},
		{Type: webrtc.TypeRTCPFBTransportCC},
	}
)

// MediaEngine handles stream codecs
type MediaEngine struct {
	webrtc.MediaEngine
}

func (e *MediaEngine) PopulateFromSDP(sd webrtc.SessionDescription) error {
	sdp := sdp.SessionDescription{}
	if err := sdp.Unmarshal([]byte(sd.SDP)); err != nil {
		return err
	}

	for _, md := range sdp.MediaDescriptions {
		if md.MediaName.Media != mediaNameAudio && md.MediaName.Media != mediaNameVideo {
			continue
		}

		for _, format := range md.MediaName.Formats {
			pt, err := strconv.Atoi(format)
			if err != nil {
				return fmt.Errorf("format parse error")
			}

			payloadType := uint8(pt)
			payloadCodec, err := sdp.GetCodecForPayloadType(payloadType)
			if err != nil {
				return fmt.Errorf("could not find codec for payload type %d", payloadType)
			}

			var codec *webrtc.RTPCodec
			switch {
			case strings.EqualFold(payloadCodec.Name, webrtc.Opus):
				codec = webrtc.NewRTPOpusCodec(payloadType, payloadCodec.ClockRate)
			case strings.EqualFold(payloadCodec.Name, webrtc.VP8):
				codec = webrtc.NewRTPVP8CodecExt(payloadType, payloadCodec.ClockRate, rtcpfb, payloadCodec.Fmtp)
			case strings.EqualFold(payloadCodec.Name, webrtc.VP9):
				codec = webrtc.NewRTPVP9CodecExt(payloadType, payloadCodec.ClockRate, rtcpfb, payloadCodec.Fmtp)
			case strings.EqualFold(payloadCodec.Name, webrtc.H264):
				codec = webrtc.NewRTPH264CodecExt(payloadType, payloadCodec.ClockRate, rtcpfb, payloadCodec.Fmtp)
			default:
				// ignoring other codecs
				continue
			}

			e.RegisterCodec(codec)
		}
	}

	// Use defaults for codecs not provided in sdp
	if len(e.GetCodecsByName(webrtc.Opus)) == 0 {
		codec := webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000)
		e.RegisterCodec(codec)
	}

	if len(e.GetCodecsByName(webrtc.VP8)) == 0 {
		codec := webrtc.NewRTPVP8CodecExt(webrtc.DefaultPayloadTypeVP8, 90000, rtcpfb, "")
		e.RegisterCodec(codec)
	}

	if len(e.GetCodecsByName(webrtc.VP9)) == 0 {
		codec := webrtc.NewRTPVP9CodecExt(webrtc.DefaultPayloadTypeVP9, 90000, rtcpfb, "")
		e.RegisterCodec(codec)
	}

	if len(e.GetCodecsByName(webrtc.H264)) == 0 {
		codec := webrtc.NewRTPH264CodecExt(webrtc.DefaultPayloadTypeH264, 90000, rtcpfb, "")
		e.RegisterCodec(codec)
	}

	return nil
}
