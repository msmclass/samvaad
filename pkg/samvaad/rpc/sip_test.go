package rpc

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/msmclass/samvaad/pkg/proto/samvaad"
)

func TestNewCreateSIPParticipantRequest(t *testing.T) {
	r := &samvaad.CreateSIPParticipantRequest{
		SipTrunkId:          "trunk",
		SipCallTo:           "+3333",
		RoomName:            "room",
		ParticipantIdentity: "",
		ParticipantName:     "",
		ParticipantMetadata: "meta",
		ParticipantAttributes: map[string]string{
			"extra": "1",
		},
		Headers: map[string]string{
			"X-B": "B2",
			"X-C": "C",
		},
		Dtmf:              "1234#",
		PlayDialtone:      true,
		WaitUntilAnswered: true,
		MediaEncryption:   samvaad.SIPMediaEncryption_SIP_MEDIA_ENCRYPT_REQUIRE,
	}
	tr := &samvaad.SIPOutboundTrunkInfo{
		SipTrunkId:         "trunk",
		Address:            "sip.example.com",
		Numbers:            []string{"+1111"},
		DestinationCountry: "us",
		AuthUsername:       "user",
		AuthPassword:       "pass",
		Headers: map[string]string{
			"X-A": "A",
			"X-B": "B1",
		},
	}
	expAttrs1 := map[string]string{
		"extra":                    "1",
		samvaad.AttrSIPCallID:      "call-id",
		samvaad.AttrSIPTrunkID:     "trunk",
		samvaad.AttrSIPTrunkNumber: "+1111",
		samvaad.AttrSIPPhoneNumber: "+3333",
		samvaad.AttrSIPHostName:    "sip.example.com",
	}
	exp := &InternalCreateSIPParticipantRequest{
		ProjectId:             "p_123",
		SipCallId:             "call-id",
		SipTrunkId:            "trunk",
		Address:               "sip.example.com",
		Hostname:              "xyz.sip.samvaad.cloud",
		DestinationCountry:    "us",
		Number:                "+1111",
		CallTo:                "+3333",
		Username:              "user",
		Password:              "pass",
		RoomName:              "room",
		ParticipantIdentity:   "sip_+3333",
		ParticipantMetadata:   "meta",
		Token:                 "token",
		WsUrl:                 "url",
		Dtmf:                  "1234#",
		PlayDialtone:          true,
		ParticipantAttributes: expAttrs1,
		Headers: map[string]string{
			"X-A": "A",
			"X-B": "B2",
			"X-C": "C",
		},
		WaitUntilAnswered: true,
		MediaEncryption:   samvaad.SIPMediaEncryption_SIP_MEDIA_ENCRYPT_REQUIRE,
		Media: &samvaad.SIPMediaConfig{
			Encryption: new(samvaad.SIPMediaEncryption_SIP_MEDIA_ENCRYPT_REQUIRE),
		},
	}
	res, err := NewCreateSIPParticipantRequest("p_123", "call-id", "xyz.sip.samvaad.cloud", "url", "token", r, tr)
	require.NoError(t, err)
	require.True(t, proto.Equal(exp, res), "%v\nvs\n%v", exp, res)

	r.HidePhoneNumber = true
	r.MediaEncryption = 0
	r.Media = &samvaad.SIPMediaConfig{
		Encryption: new(samvaad.SIPMediaEncryption_SIP_MEDIA_ENCRYPT_ALLOW),
	}
	res, err = NewCreateSIPParticipantRequest("p_123", "call-id", "xyz.sip.samvaad.cloud", "url", "token", r, tr)
	require.NoError(t, err)
	exp = &InternalCreateSIPParticipantRequest{
		ProjectId:           "p_123",
		SipCallId:           "call-id",
		SipTrunkId:          "trunk",
		Address:             "sip.example.com",
		Hostname:            "xyz.sip.samvaad.cloud",
		DestinationCountry:  "us",
		Number:              "+1111",
		CallTo:              "+3333",
		Username:            "user",
		Password:            "pass",
		RoomName:            "room",
		Token:               "token",
		WsUrl:               "url",
		Dtmf:                "1234#",
		PlayDialtone:        true,
		ParticipantIdentity: "sip_+3333",
		ParticipantAttributes: map[string]string{
			"extra":                "1",
			samvaad.AttrSIPCallID:  "call-id",
			samvaad.AttrSIPTrunkID: "trunk",
		},
		ParticipantMetadata: "meta",
		Headers: map[string]string{
			"X-A": "A",
			"X-B": "B2",
			"X-C": "C",
		},
		WaitUntilAnswered: true,
		MediaEncryption:   samvaad.SIPMediaEncryption_SIP_MEDIA_ENCRYPT_ALLOW,
		Media: &samvaad.SIPMediaConfig{
			Encryption: new(samvaad.SIPMediaEncryption_SIP_MEDIA_ENCRYPT_ALLOW),
		},
	}
	require.True(t, proto.Equal(exp, res), "%v\nvs\n%v", exp, res)

	r.HidePhoneNumber = false
	r.SipNumber = tr.Numbers[0]
	r.Trunk = &samvaad.SIPOutboundConfig{
		Hostname:            tr.Address,
		Transport:           tr.Transport,
		DestinationCountry:  "us",
		AuthUsername:        tr.AuthUsername,
		AuthPassword:        tr.AuthPassword,
		HeadersToAttributes: tr.HeadersToAttributes,
		AttributesToHeaders: tr.AttributesToHeaders,
	}
	r.SipTrunkId = ""
	exp.SipTrunkId = ""
	for k, v := range tr.Headers {
		if _, ok := r.Headers[k]; !ok {
			r.Headers[k] = v
		}
	}
	exp.ParticipantAttributes = expAttrs1
	exp.ParticipantAttributes[samvaad.AttrSIPTrunkID] = ""
	res, err = NewCreateSIPParticipantRequest("p_123", "call-id", "xyz.sip.samvaad.cloud", "url", "token", r, nil)
	require.NoError(t, err)
	require.True(t, proto.Equal(exp, res), "%v\nvs\n%v", exp, res)
}

// Regression: trunk-level MediaEncryption must be honored when the request specifies
// neither MediaEncryption nor Media. A prior version called req.Upgrade() at the top of
// NewCreateSIPParticipantRequest, which pinned req.Media.Encryption to req.MediaEncryption (0)
// before the trunk was consulted, causing outbound INVITEs to omit SRTP and upstream
// providers (e.g. Twilio) to reject with 488 / 32208.
func TestNewCreateSIPParticipantRequest_TrunkOnlyEncryption(t *testing.T) {
	r := &samvaad.CreateSIPParticipantRequest{
		SipTrunkId: "trunk",
		SipCallTo:  "+3333",
		RoomName:   "room",
	}
	tr := &samvaad.SIPOutboundTrunkInfo{
		SipTrunkId:      "trunk",
		Address:         "sip.example.com",
		Numbers:         []string{"+1111"},
		MediaEncryption: samvaad.SIPMediaEncryption_SIP_MEDIA_ENCRYPT_REQUIRE,
	}
	res, err := NewCreateSIPParticipantRequest("p_123", "call-id", "xyz.sip.samvaad.cloud", "url", "token", r, tr)
	require.NoError(t, err)
	require.Equal(t, samvaad.SIPMediaEncryption_SIP_MEDIA_ENCRYPT_REQUIRE, res.MediaEncryption)
	require.NotNil(t, res.Media)
	require.NotNil(t, res.Media.Encryption)
	require.Equal(t, samvaad.SIPMediaEncryption_SIP_MEDIA_ENCRYPT_REQUIRE, *res.Media.Encryption)
}
