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
	SendOfferAndReceiveAnswer(peerConnection *webrtc.PeerConnection) error
	// Called by the answerer in response to the offer
	ReceiveOfferAndSendAnswer(peerConnection *webrtc.PeerConnection) error
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

// Generic signaler with hook on every parts where sdp exchange is required
type genericSignaler struct {
	onSendOffer     func(sdp string) error
	onReceiveOffer  func() (string, error)
	onSendAnswer    func(sdp string) error
	onReceiveAnswer func() (string, error)
}

func (genericSignaler *genericSignaler) SendOfferAndReceiveAnswer(peerConnection *webrtc.PeerConnection) error {
	// Create the offer to be sent to the answerer
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		return fmt.Errorf("could not create offer for stdout signaling: %w", err)
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// The offer is always a local description
	peerConnection.SetLocalDescription(offer)

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	encodedOffer, err := encodeSdp(*peerConnection.LocalDescription())
	if err != nil {
		return fmt.Errorf("could not encode offer for stdout signaling: %w", err)
	}

	// Send to offer to the answerer
	err = genericSignaler.onSendOffer(encodedOffer)
	if err != nil {
		return err
	}

	// Wait for the answerer's response
	encodedAnswer, err := genericSignaler.onReceiveAnswer()
	if err != nil {
		return err
	}

	answer, err := decodeSdp(encodedAnswer)
	if err != nil {
		return fmt.Errorf("an error occured while decoding answer: %w", err)
	}

	// Save the answer
	err = peerConnection.SetRemoteDescription(*answer)
	if err != nil {
		return fmt.Errorf("could not set remote description %v: %w", answer, err)
	}

	return nil
}

func (genericSignaler *genericSignaler) ReceiveOfferAndSendAnswer(peerConnection *webrtc.PeerConnection) error {
	// Wait for the offerer's offer
	encodedOffer, err := genericSignaler.onReceiveOffer()
	if err != nil {
		return err
	}

	decodedSdp, err := decodeSdp(encodedOffer)
	if err != nil {
		return fmt.Errorf("an error occured while decoding offer: %w", err)
	}

	// Save the offer
	err = peerConnection.SetRemoteDescription(*decodedSdp)
	if err != nil {
		return fmt.Errorf("could not set remote description %v: %w", decodedSdp, err)
	}

	// Create an answer to be sent to the offerer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return fmt.Errorf("could not create answer: %w", err)
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Save the answer
	peerConnection.SetLocalDescription(answer)

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	encodedAnswer, err := encodeSdp(*peerConnection.LocalDescription())
	if err != nil {
		return fmt.Errorf("could not encode answer: %w", err)
	}

	// Send the answer to the offerer
	err = genericSignaler.onSendAnswer(encodedAnswer)
	if err != nil {
		return err
	}

	return nil
}
