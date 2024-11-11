package detector

const (
	CandidateQueueSize int = 4
)

type DetectionBuffer interface {
	Append(index int, brightnessClassified, colorDiffClassified, btDiffClassified bool)

	Resolve() []DetectionBufferElement

	ResolveClassifiedIndex() []int
}

func CreateDetectionBuffer() DetectionBuffer {
	return &detectionBuffer{
		CandidateQueue:       make([]*DetectionBufferElement, 0),
		DetectionsCollection: make([]*DetectionBufferElement, 0, CandidateQueueSize),
	}
}

type detectionBuffer struct {
	CandidateQueue       []*DetectionBufferElement
	DetectionsCollection []*DetectionBufferElement
}

func (buffer *detectionBuffer) Append(index int, brightnessClassified bool, colorDiffClassified bool, btDiffClassified bool) {
	bufferElement := createDetectionBufferElement(index, brightnessClassified, colorDiffClassified, btDiffClassified)

	buffer.DetectionsCollection = append(buffer.DetectionsCollection, bufferElement)

	if len(buffer.CandidateQueue) < CandidateQueueSize {
		buffer.CandidateQueue = append(buffer.CandidateQueue, bufferElement)
	} else {
		buffer.CandidateQueue = append(buffer.CandidateQueue[1:], bufferElement)
	}

	if len(buffer.CandidateQueue) == CandidateQueueSize {
		buffer.ClassifyCandidateQueue()
	}
}

func (buffer *detectionBuffer) Resolve() []DetectionBufferElement {
	detections := make([]DetectionBufferElement, 0, len(buffer.DetectionsCollection))
	for _, detection := range buffer.DetectionsCollection {
		detections = append(detections, *detection)
	}

	return detections
}

func (buffer *detectionBuffer) ResolveClassifiedIndex() []int {
	classifiedIndices := make([]int, 0)
	for _, detection := range buffer.DetectionsCollection {
		if detection.FinalClassification() {
			classifiedIndices = append(classifiedIndices, detection.Index)
		}
	}

	return classifiedIndices
}

func (buffer *detectionBuffer) ClassifyCandidateQueue() {
	var (
		aElement  = buffer.CandidateQueue[0]
		bElement  = buffer.CandidateQueue[1]
		cElement  = buffer.CandidateQueue[2]
		dElement  = buffer.CandidateQueue[3]
		aDetected = aElement.ClassifiedViaWeights
		bDetected = bElement.ClassifiedViaWeights
		cDetected = cElement.ClassifiedViaWeights
		dDetected = dElement.ClassifiedViaWeights
	)

	if aDetected {
		aElement.SetCorrectedViaBuffer()
	}

	if bDetected || (aDetected && dDetected) || (aDetected && cDetected) {
		bElement.SetCorrectedViaBuffer()
	}

	if cDetected || (bDetected && dDetected) || (aDetected && dDetected) {
		cElement.SetCorrectedViaBuffer()
	}

	if dDetected {
		dElement.SetCorrectedViaBuffer()
	}
}

type DetectionBufferElement struct {
	Index                               int
	BrightnessClassified                bool
	ColorDifferenceClassified           bool
	BinaryThresholdDifferenceClassified bool
	ClassifiedViaWeights                bool
	CorrectedViaBuffer                  bool
}

func createDetectionBufferElement(index int, brightnessClassified, colorDiffClassified, btDiffClassified bool) *DetectionBufferElement {
	classified := brightnessClassified && colorDiffClassified && btDiffClassified

	return &DetectionBufferElement{
		Index:                               index,
		BrightnessClassified:                brightnessClassified,
		ColorDifferenceClassified:           colorDiffClassified,
		BinaryThresholdDifferenceClassified: btDiffClassified,
		ClassifiedViaWeights:                classified,
	}
}

func (e *DetectionBufferElement) SetCorrectedViaBuffer() {
	e.CorrectedViaBuffer = true
}

func (e *DetectionBufferElement) FinalClassification() bool {
	return e.ClassifiedViaWeights || (!e.ClassifiedViaWeights && e.CorrectedViaBuffer)
}
