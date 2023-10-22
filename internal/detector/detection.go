package detector

import "github.com/Krzysztofz01/video-lightning-detector/internal/utils"

const (
	candidatesBufferSize int = 4
)

// A data structure that stores the detections and allows for automatic correction of missed detections thanks to a candidate buffer.
type DetectionBuffer interface {
	// Insert new a frame represented by the index and the detection status.
	Append(index int, detected bool)

	// Convert the detections collection into a ascending sorted slice of frames indexes.
	Resolve() []int
}

type detectionBuffer struct {
	detectionsBuffer []detectionBufferElement
	candidatesBuffer []detectionBufferElement
}

type detectionBufferElement struct {
	index    int
	detected bool
}

// Create a new detection buffer instance with the candidate buffer size set to four.
func CreateDetectionBuffer() DetectionBuffer {
	return &detectionBuffer{
		detectionsBuffer: make([]detectionBufferElement, 0),
		candidatesBuffer: make([]detectionBufferElement, 0, candidatesBufferSize),
	}
}

func (detection *detectionBuffer) Append(index int, detected bool) {
	if len(detection.candidatesBuffer) < candidatesBufferSize {
		detection.candidatesBuffer = append(detection.candidatesBuffer, detectionBufferElement{
			index:    index,
			detected: detected,
		})
	} else {
		detection.candidatesBuffer = append(detection.candidatesBuffer[1:], detectionBufferElement{
			index:    index,
			detected: detected,
		})
	}

	if len(detection.candidatesBuffer) == candidatesBufferSize {
		candidates := detection.getCandidateDetections()
		detection.handleBufferCarriage(candidates)
	}
}

func (detection *detectionBuffer) getCandidateDetections() []detectionBufferElement {
	var (
		a       = detection.candidatesBuffer[0]
		b       = detection.candidatesBuffer[1]
		c       = detection.candidatesBuffer[2]
		d       = detection.candidatesBuffer[3]
		results = make([]detectionBufferElement, 0, candidatesBufferSize)
	)

	if a.detected {
		results = append(results, a)
	}

	if b.detected || (a.detected && d.detected) || (a.detected && c.detected) {
		results = append(results, b)
	}

	if c.detected || (b.detected && d.detected) || (a.detected && d.detected) {
		results = append(results, c)
	}

	if d.detected {
		results = append(results, d)
	}

	return results
}

func (detection *detectionBuffer) handleBufferCarriage(candidateDetections []detectionBufferElement) {
	offsetRange := utils.MinInt(len(detection.detectionsBuffer), candidatesBufferSize)
	for _, candidate := range candidateDetections {
		isPresent := false
		for offset := 0; offset < offsetRange; offset += 1 {
			if detection.detectionsBuffer[len(detection.detectionsBuffer)-1-offset].index == candidate.index {
				isPresent = true
			}
		}

		if !isPresent {
			detection.detectionsBuffer = append(detection.detectionsBuffer, candidate)
		}
	}
}

func (detection *detectionBuffer) Resolve() []int {
	results := make([]int, 0, len(detection.detectionsBuffer))
	for _, detection := range detection.detectionsBuffer {
		results = append(results, detection.index)
	}

	// TODO: Do we need to sort it explicilty here?
	return results
}
