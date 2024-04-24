package wrtc

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/pion/webrtc/v4"
)

// The SDP offers and answers can be exchange by multiple protocols
// since it is not part of the WRTC specification.
type signaler interface {
	// Called by the offerer for the answerer
	SendOfferAndReceiveAnswer() error
	// Called by the answerer in response to the offer
	ReceiveOfferAndSendAnswer() error
}

// Encodes the sdp in json then in base64
func encodeSdp(sdp webrtc.SessionDescription) (string, error) {
	encodedSdp, err := json.Marshal(sdp)
	if err != nil {
		return "", fmt.Errorf("could not marshal as json sdp %v: %w", sdp, err)
	}

	return base64.StdEncoding.EncodeToString(encodedSdp), nil
}

// Decode the sdp from base64 then from json
func decodeSdp(sdp string) (*webrtc.SessionDescription, error) {
	jsonSdp, err := base64.StdEncoding.DecodeString(sdp)
	if err != nil {
		return nil, fmt.Errorf("could not decode base64 encoded sdp %s: %w", sdp, err)
	}

	decodedSdp := webrtc.SessionDescription{}
	err = json.Unmarshal(jsonSdp, &decodedSdp)
	if err != nil {
		return nil, fmt.Errorf("could not decode json encoded sdp %s: %w", jsonSdp, err)
	}

	return &decodedSdp, nil
}
